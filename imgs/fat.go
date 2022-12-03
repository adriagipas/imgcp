/*
 * Copyright 2022 Adrià Giménez Pastor.
 *
 * This file is part of adriagipas/imgcp.
 *
 * adriagipas/imgcp is free software: you can redistribute it and/or
 * modify it under the terms of the GNU General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * adriagipas/imgcp is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
 * General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with adriagipas/imgcp.  If not, see <https://www.gnu.org/licenses/>.
 */
/*
 *  fat.go - Estructures comuns a tots els sistemes de fitxer FAT.
 *
 */

package imgs

import (
  "errors"
  "fmt"
  "io"
  "os"
  "strings"
  "unicode/utf16"

  "github.com/adriagipas/imgcp/utils"
)


/***********/
/* FAT BPB */
/***********/

type _FAT_BPB struct {

  oem              string // OEM identifier. No és molt important
  bytes_per_sec    uint16 // Bytes per sector
  secs_per_clu     uint8  // Sectors per cluster
  reserved_secs    uint16 // Nombre de sectors reservats
  num_fat          uint8  // Nombre de File Allocation Tables (FAT's),
                          // típicament 2
  num_root_entries uint16 // Nombre d'entrades en el directori arrel.
  num_secs         uint32 // Nombre de sectors
  media_desc       uint8  // Media descriptor type
  secs_per_fat     uint16 // Nombre de sectors per FAT. Sols en FAT12/16
  secs_per_track   uint16 // Nombre de sectors per track
  num_heads        uint16 // Nombre de capçals
  num_hidden_sec   uint32 // Nombre de sectors ocults
  
}


// Ompli el contingut referent al BPB a partir de les dades que se li
// passen.
func (self *_FAT_BPB) read(data []byte) error {

  // Preparació
  var strb strings.Builder
  
  // Comprova JMP SHORT 3C NOP (o semblant)
  if data[0]!=0xeb && data[0]!=0xe9 && data[2]!=0x90 {
    return fmt.Errorf ( "Invalid FAT BPB: The first three bytes "+
      "(%02x %02x %02x) should be (eb 3c 90) or similar ",
      data[0], data[1], data[2] )
  }

  // Obté OEM
  strb.Write ( data[3:3+8] )
  self.oem= strb.String ()

  // Obté bytes per sec
  self.bytes_per_sec= uint16(data[0xb]) | (uint16(data[0xc])<<8)
  if self.bytes_per_sec == 0 {
    return errors.New ( "Invalid FAT BPB: bytes per sector is 0" )
  }

  // Obté sectors per cluster
  self.secs_per_clu= data[0xd]

  // Obté nombre de sectors reservats
  self.reserved_secs= uint16(data[0xe]) | (uint16(data[0xf])<<8)

  // Obté nombre de FATs
  self.num_fat= data[0x10]

  // Obté nombre d'entrades en el directori arrel
  self.num_root_entries= uint16(data[0x11]) | (uint16(data[0x12])<<8)

  // Obté nombre de sectors
  self.num_secs= uint32(data[0x13]) | (uint32(data[0x14])<<8)
  if self.num_secs == 0 {
    self.num_secs= uint32(data[0x20]) |
      (uint32(data[0x21])<<8) |
      (uint32(data[0x22])<<16) |
      (uint32(data[0x23])<<24)
    if self.num_secs == 0 {
      return errors.New ( "Invalid FAT BPB: number of sectors is 0" )
    }
  }

  // Obté media descriptor type
  self.media_desc= data[0x15]

  // Obté nombre de sectors per FAT
  self.secs_per_fat= uint16(data[0x16]) | (uint16(data[0x17])<<8)

  // Obté nombre de sectors per track
  self.secs_per_track= uint16(data[0x18]) | (uint16(data[0x19])<<8)

  // Obté nombre de capçals
  self.num_heads= uint16(data[0x1a]) | (uint16(data[0x1b])<<8)

  // Obté nombre de sectors ocults
  self.num_hidden_sec= uint32(data[0x1c]) |
      (uint32(data[0x1d])<<8) |
      (uint32(data[0x1e])<<16) |
      (uint32(data[0x1f])<<24)
  
  return nil
  
} // end read


func (self *_FAT_BPB) fPrintfInfo(
  
  f      *os.File,
  file   io.Writer,
  prefix string,
  
) error {

  // Preparació
  P := func(args... any) {
    fmt.Fprint ( file, prefix )
    fmt.Fprintln ( file, args... )
  }
  F := func(format string, args... any) {
    fmt.Fprint ( file, prefix )
    fmt.Fprintf ( file, format, args... )
    fmt.Fprint ( file, "\n" )
  }
  
  // Imprimeix
  P("BIOS Parameter Block")
  P("--------------------")
  P("")
  F("  * OEM:                 '%s'", self.oem )
  F("  * BYTES/SECTOR:        %d", self.bytes_per_sec )
  F("  * SECTORS/CLUSTER:     %d", self.secs_per_clu )
  F("  * RESERVED SECTORS:    %d", self.reserved_secs )
  F("  * NUM. FAT:            %d", self.num_fat )
  F("  * ROOT DIR. ENTRIES:   %d", self.num_root_entries )
  F("  * NUM. SECTORS:        %d (%s)", self.num_secs,
    utils.NumBytesToStr ( uint64(self.num_secs)*uint64(self.bytes_per_sec) ))
  F("  * MEDIA DESC. TYPE:    %02Xh", self.media_desc )
  F("  * SECTORS/FAT:         %d", self.secs_per_fat )
  F("  * SECTORS/TRACK:       %d", self.secs_per_track )
  F("  * NUM. HEADS:          %d", self.num_heads )
  F("  * NUM. HIDDEN SECTORS: %d", self.num_hidden_sec )
  
  return nil
  
} // end fPrintfInfo


/*****************/
/* FAT_Directory */
/*****************/

type _FAT_Directory interface {

  getIter() *_FAT_DirectoryIter
  
}


/*********************/
/* FAT_DirectoryIter */
/*********************/

const FAT_DIR_READ_ONLY = 0x01
const FAT_DIR_HIDDEN    = 0x02
const FAT_DIR_SYSTEM    = 0x04
const FAT_DIR_VOLUME_ID = 0x08
const FAT_DIR_DIRECTORY = 0x10
const FAT_DIR_ARCHIVE   = 0x20

const FAT_DIR_LFN = FAT_DIR_READ_ONLY|FAT_DIR_HIDDEN|
  FAT_DIR_SYSTEM|FAT_DIR_VOLUME_ID

type _FAT_DirectoryIter struct {

  // Sobre les dades
  pos  int
  data []byte

  // Auxiliars
  strb          strings.Builder
  lname         string // Si està buit vol dir que no hi ha
  current_lname string // La que gasta el fitxer
  last_pos      uint8
  lbuf          [13]uint16 // Per a codificar els bytes
  
}

func (self *_FAT_DirectoryIter) getPosEntry() int {
  return self.pos
} // end getIter

// Comprova si l'iterador actual té més entrades o no.
func (self *_FAT_DirectoryIter) end() bool {
  if self.pos >= len(self.data) || self.data[self.pos]==0x00 {
    return true
  } else {
    return false
  }
}

func (self *_FAT_DirectoryIter) next() {

  // Incrementa
  self.pos+= 32

  if self.pos >= len(self.data) {
    return
  }
  
  // Long file names
  if self.getAttributes() == FAT_DIR_LFN {
    
    // Reseteja si és l'últim
    order := self.data[self.pos]
    pos := order&0x03f
    if order&0x40 == 0x40 {
      self.last_pos= pos
      self.lname= ""
    } else if pos+1 != self.last_pos { // Reseteje també
      self.lname= ""
      self.last_pos= 0
    }

    // Llig el nom llarg en el buffer
    end,buf,p := false,self.lbuf[:],0
    for i := 1; i < 11 && !end; i+= 2 {
      aux := uint16(self.data[self.pos+i]) |
        (uint16(self.data[self.pos+i+1])<<8)
      if aux == 0xffff { end= true } else { buf[p]= aux; p++ }
    }
    for i := 14; i < 26 && !end; i+= 2 {
      aux := uint16(self.data[self.pos+i]) |
        (uint16(self.data[self.pos+i+1])<<8)
      if aux == 0xffff { end= true } else { buf[p]= aux; p++ }
    }
    for i := 28; i < 32 && !end; i+= 2 {
      aux := uint16(self.data[self.pos+i]) |
        (uint16(self.data[self.pos+i+1])<<8)
      if aux == 0xffff { end= true } else { buf[p]= aux; p++ }
    }
    
    // Codifica a string
    if p > 0 {
      self.lname= string(utf16.Decode ( buf[:p] )) + self.lname
    }
    
  } else {
    self.current_lname= self.lname
    self.lname= ""
  }
  
} // end next

// Torna els attributs del fitxer actual
func (self *_FAT_DirectoryIter) getAttributes() uint8 {
  return self.data[self.pos+11]
}

// Ha de ser vàlid
func (self *_FAT_DirectoryIter) unused() bool {
  return self.data[self.pos]==0xe5
}

func (self *_FAT_DirectoryIter) getName() string {
  
  self.strb.Reset ()
  self.strb.Write ( self.data[self.pos:self.pos+8] )
  name := self.strb.String ()

  return name
}

func (self *_FAT_DirectoryIter) getExt() string {
  
  self.strb.Reset ()
  self.strb.Write ( self.data[self.pos+8:self.pos+11] )
  ext := self.strb.String ()

  return ext
}

func (self *_FAT_DirectoryIter) getLongName() string {
  return self.current_lname
} // end getLongName

func (self *_FAT_DirectoryIter) getSize() uint32 {
  return uint32(self.data[self.pos+28]) |
    (uint32(self.data[self.pos+29])<<8) |
    (uint32(self.data[self.pos+30])<<16) |
    (uint32(self.data[self.pos+31])<<24)
}

func (self *_FAT_DirectoryIter) getTime(pos int) (hh int,mm int,ss int) {
  val := uint16(self.data[self.pos+pos]) |
    (uint16(self.data[self.pos+pos+1])<<8)
  hh= int(val>>11)
  mm= int((val>>5)&0x3f)
  ss= int(val&0x1f)*2
  return hh,mm,ss
}

func (self *_FAT_DirectoryIter) getCreationTime() (hh int,mm int,ss int) {
  return self.getTime ( 14 )
}

func (self *_FAT_DirectoryIter) getModificationTime() (hh int,mm int,ss int) {
  return self.getTime ( 22 )
}

func (self *_FAT_DirectoryIter) getDate(pos int) (yy int,mm int,dd int) {
  val := uint16(self.data[self.pos+pos]) |
    (uint16(self.data[self.pos+pos+1])<<8)
  yy= (int(val>>9) + 80)%100
  mm= int((val>>5)&0xf)
  dd= int(val&0x1f)
  return yy,mm,dd
}

func (self *_FAT_DirectoryIter) getCreationDate() (yy int,mm int,dd int) {
  return self.getDate ( 16 )
}

func (self *_FAT_DirectoryIter) getModificationDate() (yy int,mm int,dd int) {
  return self.getDate ( 24 )
}

func (self *_FAT_DirectoryIter) getCluster16() uint16 {
  return uint16(self.data[self.pos+26]) | (uint16(self.data[self.pos+27])<<8)
}

func (self *_FAT_DirectoryIter) list(file io.Writer) {
  
  P := func(args... any) {
    fmt.Fprint ( file, args... )
  }
  F := func(format string,args... any) {
    fmt.Fprintf ( file, format, args... )
  }
  
  attr := self.getAttributes ()
  
  // Attributs
  if (attr&FAT_DIR_DIRECTORY) != 0 { P("d") } else { P("-") }
  if (attr&FAT_DIR_HIDDEN) != 0 { P("h") } else { P("-") }
  if (attr&FAT_DIR_SYSTEM) != 0 { P("s") } else { P("-") }
  if (attr&FAT_DIR_VOLUME_ID) != 0 { P("v") } else { P("-") }
  if (attr&FAT_DIR_READ_ONLY) != 0 { P("-") } else { P("w") }
  
  P("  ")
  
  // Grandària
  size := utils.NumBytesToStr ( uint64(self.getSize ()) )
  for i := 0; i < 10-len(size); i++ {
    P(" ")
  }
  P(size,"  ")
  
  // Date
  year,month,day := self.getModificationDate ()
  F("%02d/%02d/%02d  ",day,month,year)
  
  // Time
  hh,mm,ss := self.getModificationTime ()
  F("%02d:%02d:%02d  ",hh,mm,ss)
  
  // Nom
  short_name := strings.TrimSpace ( self.getName () )
  ext := strings.TrimSpace ( self.getExt () )
  if ext != "" {
    short_name+= "."+ext
  }
  long_name := strings.TrimSpace ( self.getLongName () )
  if long_name == "" { P(short_name) } else { P(long_name) }
  
  P("\n")
  
} // end list


/*****************/
/* FAT FIND PATH */
/*****************/

type _FAT_FindPath struct {
  is_dir  bool
  dir     _FAT_Directory
  file_it *_FAT_DirectoryIter
}


/*********/
/* UTILS */
/*********/

// Aquesta funció comprova si un nom de fitxer compleix amb
// l'estàndard 8.3, i si ho fa torna el nom preparat per a una
// entrada.
func FAT_GetFileName83(file_name string) ([]byte,error) {

  // Formatació prèvia.
  file_name= strings.ToUpper ( strings.TrimSpace ( file_name ) )
  if file_name == "" || file_name == "." {
    return nil,errors.New ( "Empty name" )
  }

  // Comprova format
  // NOTA!! Aparentment els caràcters >=128 valen però de moment passe
  // d'ells. En eixe aspecte per codificar el 0xe5 s'empra el 0x5
  for _,c := range file_name {
    if (c < 'A' || c > 'Z' ) && (c < '0' || c > '9' ) && c != '!' &&
      c != '#' && c != '$' && c != '%' && c != '&' && c != '\'' &&
      c != '(' && c != ')' && c != '-' && c != '@' && c != '^' &&
      c != '_' && c != '`' && c != '{' && c != '}' && c != '~' && c != '.' {
      return nil,fmt.Errorf ( "Character not supported in 8.3 name: %s",
        file_name )
    }
  }

  // Separa en nom i extensió
  var mem [11]byte
  var i int
  ret := mem[:]
  tokens := strings.Split ( file_name, "." )
  if len(tokens) > 2 {
    return nil,fmt.Errorf ( "Wrong file name format: %s", file_name )
  }
  // --> Nom
  token := tokens[0]
  if len(token) == 0 {
    return nil,fmt.Errorf ( "File name without name: %s", file_name )
  }
  if len(token) > 8 {
    return nil,fmt.Errorf ( "File name too long: %s", file_name )
  }
  for i= 0; i < len(token); i++ {
    ret[i]= token[i]
  }
  for ; i < 8; i++ {
    ret[i]= ' '
  }
  // --> Extensió
  i= 0;
  if len(tokens) == 2 {
    token= tokens[1]
    if len(token) > 3 {
      return nil,fmt.Errorf ( "File extension too long: %s", file_name )
    }
    for ; i < len(token); i++ {
      ret[i+8]= token[i]
    }
  }
  for ; i < 3; i++ {
    ret[i+8]= ' '
  }
  
  return ret,nil
  
} // end FAT_GetFileName83
