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
 *  fat1216.go - Conté el codi comú per als sistemes de fitxer FAT12 i
 *               FAT16.
 *
 */

package imgs

import (
  "errors"
  "fmt"
  "io"
  "os"
  "strings"
  "time"

  "github.com/adriagipas/imgcp/utils"
)


/******************/
/* FAT12/16 TABLE */
/******************/

type _FAT1216_Table interface {

  // Imprimeix info FAT table
  fPrintInfo(f *os.File, file io.Writer, prefix string, br *_FAT1216_BR) error

  // Torna el BAD Cluster
  badCluster() uint16

  // Nombre d'entrades en la taula
  length() uint16
  
  // Torna el següent cluster de la cadena.
  chain(ind uint16) uint16

  // Torna les dades internes
  getData() []byte

  // Escriu un valor en una entrada
  write(ind uint16, val uint16)
  
}


/************/
/* FAT12/16 */
/************/

// Segueix una aproximació lazzy
type _FAT1216 struct {

  file_name string
  offset    int64  // Offset primer byte
  length    uint64 // Grandària en bytes
  is_fat16  bool   // Si és cert indica FAT16 en cas contrari FAT12
  
  // Estructures internes inicialment no inicialitzades
  br_init      bool
  br           _FAT1216_BR
  fat          _FAT1216_Table
  fat_modified bool // Indica que la FAT s'ha d'escriure en el disc
  
}


func newFAT1216(
  
  file_name string,
  offset    int64,
  length    uint64,
  is_fat16  bool,
  
) (*_FAT1216,error) {

  // Comprovacions bàsiques
  if length == 0 {
    return nil,errors.New ( "Invalid zero length for a FAT12/16 partition" )
  }

  // Crea
  ret := _FAT1216 {
    file_name: file_name,
    offset: offset,
    length: length,
    is_fat16: is_fat16,

    br_init: false,
    fat: nil,
    
  }
  
  return &ret,nil
  
} // end newFAT1216


func (self *_FAT1216) GetRootDirectory() (Directory,error) {

  // Obri el fitxer
  f,err := os.Open ( self.file_name )
  if err != nil { return nil,err }
  
  // Llig el FAT Boot Record
  br,err := self.fGetBR ( f )
  if err != nil { return nil,err }

  // Llig el contingut
  tmp_secs := int64(br.bpb.reserved_secs) +
    int64(br.bpb.num_fat)*int64(br.bpb.secs_per_fat)
  offset := self.offset + tmp_secs*int64(br.bpb.bytes_per_sec)
  length := int64(br.bpb.num_root_entries)*32
  data := make ( []byte, length )
  if err := self.readBytes ( f, data, offset ); err != nil {
    return nil,fmt.Errorf ( "Error while reading root directory: %s", err )
  }

  // Offsets i flags modificat
  offsets := make ( []int64, 1 )
  offsets[0]= offset
  mod := make ( []bool, 1 )
  mod[0]= false
  
  // Crea objecte
  ret := _FAT1216_Directory{
    img: self,
    data: data,
    offs: offsets,
    mod: mod,
    is_root: true,
    dir_cluster: 0,
  }

  // Tanca
  f.Close ()
  
  return &ret,nil


} // end GetRootDirectory


func (self *_FAT1216) PrintInfo(
  file   io.Writer,
  prefix string,
) error {

  f,err := os.Open ( self.file_name )
  if err != nil { return err }
  if self.is_fat16 {
    fmt.Fprintf ( file, "%sFAT16 image\n", prefix )
  } else {
    fmt.Fprintf ( file, "%sFAT12 image\n", prefix )
  }
  fmt.Fprintln ( file, prefix, "" )
  err = self.fPrintInfo ( f, file, prefix )
  f.Close ()

  return err
  
} // end PrintInfo


func (self *_FAT1216) fAllocCluster(f *os.File) (uint16,error) {

  // Obte fat
  fat,err := self.fGetFAT ( f )
  if err != nil { return 0,err }

  // Busca el primer cluster buit
  var i uint16
  var cluster uint16= 0 // El 0 està prohibit, significa lliure
  for i= 2; i < fat.length (); i++ {
    if val := fat.chain ( i ); val == 0 {
      cluster= i
      break
    }
  }

  // Comprovacions
  if cluster == 0 {
    return 0,errors.New ( "Not enough space" )
  } else {
    fat.write ( cluster, fat.badCluster () + 1 ) // Fi de cadena
    self.fat_modified= true
  }
  
  return cluster,nil
  
} // end fAllocCluster


func (self *_FAT1216) fGetBR(f *os.File) (*_FAT1216_BR,error) {
  
  if !self.br_init {
    if err := self.br.read ( f, self.offset, self.length ); err != nil {
      return nil,fmt.Errorf ( "Unable to read FAT1216 BR: %s", err )
    }
    self.br_init= true
  }
  
  return &self.br,nil
  
} // end fGetBR


func (self *_FAT1216) fGetClusterSize(f *os.File) (int64,error) {
  
  br,err := self.fGetBR ( f )
  if err != nil { return -1,err }
  cluster_size := int64(br.bpb.secs_per_clu)*int64(br.bpb.bytes_per_sec)
  
  return cluster_size,nil
  
} // end fGetClusterSize


func (self *_FAT1216) fGetDataOffset(f *os.File) (int64,error) {

  // Llig BR
  br,err := self.fGetBR ( f )
  if err != nil { return -1,err }

  // Calcula data_offset
  sec_size := int64(br.bpb.bytes_per_sec)
  fat_size := int64(br.bpb.num_fat)*int64(br.bpb.secs_per_fat)
  root_dir_secs := (int64(br.bpb.num_root_entries)*32 + sec_size - 1) / sec_size
  data_offset := self.offset +
    (int64(br.bpb.reserved_secs) + fat_size + root_dir_secs)*sec_size
  
  return data_offset,nil
  
} // end fGetDataOffset


func (self *_FAT1216) fGetFAT(f *os.File) (_FAT1216_Table,error) {

  if self.fat == nil {
    br,err := self.fGetBR ( f )
    if err != nil { return nil, err }
    if self.is_fat16 {
      self.fat,err= self.fReadFAT16 ( f, br )
      if err != nil { return nil, err }
    } else {
      self.fat,err= self.fReadFAT12 ( f, br )
      if err != nil { return nil, err }
    }
    self.fat_modified= false
  }
  
  return self.fat,nil
  
} // end fGetFAT


func (self *_FAT1216) fPrintInfo(
  
  f      *os.File,
  file   io.Writer,
  prefix string,
  
)  error {

  // Imprimeix el FAT Boot Record
  br,err := self.fGetBR ( f )
  if err != nil { return err }
  if err := br.fPrintfInfo ( f, file, prefix ); err != nil {
    return fmt.Errorf ( "Unable to print FAT12/16 BR: %s", err )
  }
  
  // Imprimeix informació FAT Table
  fat_table, err := self.fGetFAT ( f )
  if err != nil { return err }
  if err = fat_table.fPrintInfo ( f, file, prefix, br ); err != nil {
    return fmt.Errorf ( "Unable to print FAT12/16 table info: %s", err )
  }
  
  return nil
  
} // end fPrintInfo


func (self *_FAT1216) fReadFAT12(

  f  *os.File,
  br *_FAT1216_BR,
  
) (*_FAT12_Table,error) {
  
  // Adreça del primer sector de la taula
  first_fat_sector := self.offset +
    int64(br.bpb.reserved_secs)*int64(br.bpb.bytes_per_sec)
  fat_size := int64(br.bpb.secs_per_fat)*int64(br.bpb.bytes_per_sec)
  if fat_size%2 != 0 {
    return nil,fmt.Errorf ( "Wrong FAT12 table size: %d", fat_size )
  }

  // Reserva i llig
  var ret _FAT12_Table= make ( []byte, fat_size )
  if err := self.readBytes ( f, ret, first_fat_sector ); err != nil {
    return nil,fmt.Errorf ( "Error while reading FAT12 table: %s", err )
  }
  
  // Comprovacions semàntiques
  if tmp := (0xF00 | uint16(br.bpb.media_desc)); ret.get ( 0 ) != tmp {
    return nil,fmt.Errorf ( "FAT12[0] and media descriptor type differ:"+
      " %03X != %03X ", ret.get ( 0 ), tmp )
  }
  if ret.get ( 1 ) != 0xFFF {
    return nil,fmt.Errorf ( "FAT12[1] (%03X) != FFF", ret.get ( 1 ) )
  }
  
  return &ret,nil
  
} // end fReadFAT12


func (self *_FAT1216) fReadFAT16(

  f  *os.File,
  br *_FAT1216_BR,
  
) (*_FAT16_Table,error) {
  
  // Adreça del primer sector de la taula
  first_fat_sector := self.offset +
    int64(br.bpb.reserved_secs)*int64(br.bpb.bytes_per_sec)
  fat_size := int64(br.bpb.secs_per_fat)*int64(br.bpb.bytes_per_sec)
  if fat_size%2 != 0 {
    return nil,fmt.Errorf ( "Wrong FAT16 table size: %d", fat_size )
  }

  // Reserva i llig
  var ret _FAT16_Table= make ( []byte, fat_size )
  if err := self.readBytes ( f, ret, first_fat_sector ); err != nil {
    return nil,fmt.Errorf ( "Error while reading FAT16 table: %s", err )
  }

  // Comprovacions semàntiques
  if tmp := uint16(int16(int8(br.bpb.media_desc))); ret.get ( 0 ) != tmp {
    return nil,fmt.Errorf ( "FAT16[0] and media descriptor type differ:"+
      " %04X != %04X ", ret.get ( 0 ), tmp )
  }
  if ret.get ( 1 ) != 0xFFFF {
    return nil,fmt.Errorf ( "FAT16[1] (%04X) != FFFF", ret.get ( 1 ) )
  }
  
  return &ret,nil
  
} // end fReadFAT16


func (self *_FAT1216) fWriteFAT(f *os.File) (error) {

  // Si no s'ha modificat no fa res
  if self.fat == nil || !self.fat_modified {
    return nil
  }
  
  // Obté Boot Record
  br,err := self.fGetBR ( f )
  if err != nil { return err }

  // Escriu totes les còpies
  first_fat_sector := self.offset +
    int64(br.bpb.reserved_secs)*int64(br.bpb.bytes_per_sec)
  fat_data := self.fat.getData ()
  fat_size := int64(len(fat_data))
  for i := 0; i < int(br.bpb.num_fat); i++ {
    if err := self.writeBytes ( f, fat_data, first_fat_sector ); err != nil {
      return fmt.Errorf ( "Error while writing FAT table %d: %s", i+1, err )
    }
    first_fat_sector+= fat_size
  }
  
  // Actualitza
  self.fat_modified= false
  
  return nil
  
} // end fWriteFAT


// Llig bytes d'un fitxer fent comprovacions
func (self *_FAT1216) readBytes(

  f      *os.File,
  buf    []byte,
  offset int64,
  
) error {
  return utils.ReadBytes ( f, self.offset, int64(self.length), buf, offset )
} // readBytes


// Escriu bytes en un fitxer fent comprovacions
func (self *_FAT1216) writeBytes(

  f      *os.File,
  buf    []byte,
  offset int64,
  
) error {
  return utils.WriteBytes ( f, self.offset, int64(self.length), buf, offset )
} // writeBytes


/************************/
/* FAT12/16 FILE READER */
/************************/

type _FAT1216_FileReader struct {

  // Punters estructures externes
  f   *os.File // Fitxer d'on llegir
  img *_FAT1216  // Punter a la classe pare
  it  *_FAT_DirectoryIter // Metadades del fitxer

  // Estat intern
  data_offset  int64  // Offset on comencen les dades
  cluster_size int64  // Grandària d'un cluster
  cluster_data []byte // Dades del cluster acutal
  pos          int64  // Posició dins del cluster actual
  cluster      uint16 // cluster actual
  remain       uint32 // Bytes que falten per llegir
  
}


func (self *_FAT1216_FileReader) load_next_cluster() error {
  
  // Si està buit no fa res
  if self.remain == 0 { return nil }

  // Obté fat
  fat,err := self.img.fGetFAT ( self.f )
  if err != nil { return err }
  
  // Comprovacions
  if self.cluster <= 1 || self.cluster >= fat.badCluster () {
    return fmt.Errorf ( "Trying to read a file from an invalid"+
      " cluster number: %X", self.cluster )
  }
  
  // Llig cluster
  offset := self.data_offset + int64(self.cluster-2)*self.cluster_size
  if err := self.img.readBytes ( self.f,
    self.cluster_data, offset ); err != nil {
    return fmt.Errorf ( "Error while reading cluster %d: %s",
      self.cluster, err )
  }

  // Actualitza estat
  self.pos= 0
  self.cluster= fat.chain ( self.cluster )
  
  return nil
  
} // end load_next_cluster


func (self *_FAT1216_FileReader) Read(buf []byte) (int,error) {

  lbuf,pos := len(buf),0
  for ; pos < lbuf && self.remain > 0; {
    
    // Bytes a llegir
    remain_cluster := self.cluster_size-self.pos
    remain_buf := lbuf-pos
    var nbytes int
    if remain_cluster > int64(self.remain) {
      if remain_buf > int(self.remain) {
        nbytes= int(self.remain)
      } else {
        nbytes= remain_buf
      }
    } else { // remain_cluster < self.remain
      if remain_buf > int(remain_cluster) {
        nbytes= int(remain_cluster)
      } else {
        nbytes= remain_buf
      }
    }
    
    // Copia del cluster
    for i := 0; i < nbytes; i++ {
      buf[pos]= self.cluster_data[self.pos]
      pos++
      self.pos++
    }
    self.remain-= uint32(nbytes)
    
    // Llig el següent cluster si cal
    if self.pos == self.cluster_size {
      if err := self.load_next_cluster (); err != nil {
        return -1,err
      }
    }
    
  }
  
  return pos,nil
  
} // end Read


func (self *_FAT1216_FileReader) Close() {
  self.f.Close ()
}


/************************/
/* FAT12/16 FILE WRITER */
/************************/

type _FAT1216_FileWriter struct {

  // Punters estructures externes
  f    *os.File            // Fitxer d'on llegir
  img  *_FAT1216           // Punter a la classe pare
  pdir *_FAT1216_Directory // Directori que conté el fitxer
  
  // Estat intern
  entry        []byte // Entrada en el directori.
  data_offset  int64  // Offset on comencen les dades
  cluster_size int64  // Grandària d'un cluster
  cluster_data []byte // Dades del cluster actual
  pos          int64  // Posició dins del cluster actual
  cluster      uint16 // cluster actual
  size         uint32 // Grandària del fitxer. Inicialment 0
  
}


// Si CHAIN és true reserva un nou cluster
func (self *_FAT1216_FileWriter) write_cluster(chain bool) error {

  // Actualitza grandària
  // COMPTE!! por ser no estiga ple, per això fique self.pos
  tmp_size := int64(self.size) + self.pos
  if tmp_size > 0xFFFFFFFF {
    return errors.New ( "Error while writing cluster: file is too big" )
  }
  self.size= uint32(tmp_size)
  
  // Escriu
  offset := self.data_offset + int64(self.cluster-2)*self.cluster_size
  if err := self.img.writeBytes ( self.f,
    self.cluster_data[:self.pos], offset ); err != nil {
    return fmt.Errorf ( "Error while writing cluster %d: %s",
      self.cluster, err )
  }
  
  // Encadena
  if chain {
    
    // Obté fat
    fat,err := self.img.fGetFAT ( self.f )
    if err != nil { return err }

    // Crea nou cluster
    cluster,err := self.img.fAllocCluster ( self.f )
    if err != nil { return err }
    fat.write ( self.cluster, cluster )
    self.cluster= cluster
    self.pos= 0
    
  }
  
  return nil
  
} // end write_cluster


func (self *_FAT1216_FileWriter) Write(buf []byte) (int,error) {

  lbuf,pos := len(buf),0
  for ; pos < lbuf; {

    // Escriu cluster. Ho faig sempre just abans d'intentar copiar
    // alguna cosa. M'assegure del chain.
    if self.pos == self.cluster_size {
      if err := self.write_cluster ( true ); err != nil {
        return -1,err
      }
    }
    
    // Bytes a escriure
    var nbytes int
    remain_cluster := self.cluster_size-self.pos
    remain_buf := lbuf-pos
    if remain_buf > int(remain_cluster) {
      nbytes= int(remain_cluster)
    } else {
      nbytes= remain_buf
    }
    
    // Copia
    new_self_pos := self.pos+int64(nbytes)
    new_pos := pos+nbytes
    copy ( self.cluster_data[self.pos:new_self_pos], buf[pos:new_pos] )
    self.pos,pos= new_self_pos,new_pos
    
  }
  
  return pos,nil
  
} // end Write


func (self *_FAT1216_FileWriter) Close() error {

  // Escriu dades pendents.
  if self.pos > 0 {
    if err := self.write_cluster ( false ); err != nil {
      return err
    }
  }
  
  // Actualitza grandària fitxer
  self.entry[28]= uint8(self.size)
  self.entry[29]= uint8(self.size>>8)
  self.entry[30]= uint8(self.size>>16)
  self.entry[31]= uint8(self.size>>24)
  
  // Escriu resta estructures en el disc
  if err := self.pdir.fWrite ( self.f ); err != nil {
    return err
  }
  if err := self.img.fWriteFAT ( self.f ); err != nil {
    return err
  }

  // Tanca
  self.f.Close ()
  
  return nil
  
} // end Close


/****************/
/* FAT12/16 BR */
/****************/

const FAT1216_BR_SIZE = 512

type _FAT1216_BR struct {
  
  bpb    _FAT_BPB // BIOS Parameter Block
  number uint8    // Número de dispositiu, no és molt util
  id     uint32   // VolumeID. No és molt important
  label  string   // Etiqueta del volum
  sys_id string   // Identificació de sistema. Aparentment no cal
                  // fiar-se
  esign  uint8
}

// Omplie el contingut referent al BR
func (self *_FAT1216_BR) read(
  
  f      *os.File, // Fitxer d'on llegir
  offset int64,   // Primer byte de la partició
  length uint64,  // Últim byte de la partició
  
) error {

  // Preparació
  var strb strings.Builder
  
  // Intenta llegit el primer sector
  var buf [FAT1216_BR_SIZE]byte
  if length < FAT1216_BR_SIZE {
    return fmt.Errorf("Not enough bytes (%d) to read the FAT Boot Record",
      length)
  }
  if offset < 0 {
    return fmt.Errorf("Unable to read FAT Boot Record: invalid offset (%d)",
      offset)
  }
  tmp_off,err := f.Seek ( offset, 0 )
  if err != nil { return err }
  if tmp_off != offset {
    return errors.New("Unexpected error occurred while reading FAT Boot Record")
  }
  nbytes,err := f.Read ( buf[:] )
  if err != nil { return err }
  if nbytes != FAT1216_BR_SIZE {
    return errors.New("Unable to read FAT Boot Record")
  }
  
  // Llig el BIOS BR
  if err := self.bpb.read ( buf[:] ); err != nil {
    return err
  }

  // Obté número
  self.number= buf[0x24]

  // Comprova signatura
  self.esign= buf[0x26]
  
  // Obté VolumeId
  self.id= uint32(buf[0x27]) |
    (uint32(buf[0x28])<<8) |
    (uint32(buf[0x29])<<16) |
    (uint32(buf[0x2a])<<24)

  // Obté label
  strb.Write ( buf[0x2b:0x2b+11] )
  self.label= strb.String ()

  // Obté system identifier
  strb.Reset ()
  strb.Write ( buf[0x36:0x36+8] )
  self.sys_id= strb.String ()

  // Comprova bootable signature
  if buf[0x1fe]!=0x55 && buf[0x1ff]!=0xaa {
    return fmt.Errorf ( "Invalid bootable partiture signature (%02X%02Xh)"+
      " in FAT12/16 Extended Boot Record", buf[0x1ff], buf[0x1fe] )
  }
  
  return nil
  
} // end read


func (self *_FAT1216_BR) fPrintfInfo(
  
  f      *os.File,
  file   io.Writer,
  prefix string,
  
) error {

  // Imprimeix BPB
  if err := self.bpb.fPrintfInfo ( f, file, prefix ); err != nil {
    return err
  }

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

  // Imprimeix Extended Boot Record
  if self.esign == 0x28 || self.esign == 0x29 {
    P("")
    P("Extended Boot Record")
    P("--------------------")
    P("")
    F("  * DRIVE NUMBER: %02Xh", self.number )
    F("  * VOLUME ID:    %08Xh", self.id )
    if self.esign == 0x29 {
      F("  * VOLUM LABEL:  '%s'", self.label )
      F("  * SYSTEM ID:    '%s'", self.sys_id )
    }
  }
  
  return nil
  
} // end fPrintfInfo


/**********************/
/* FAT12/16 DIRECTORY */
/**********************/

type _FAT1216_Directory struct {

  img          *_FAT1216 // Referència a imatge
  offs         []int64   // Offset de cada cluster i nombre de clusters
  mod          []bool    // Per a cada cluster indica si ha sigut o no
                         // modificat.
  data         []byte    // Contingut
  is_root      bool      // Indica que és un directori root i per tant no es
                         // pot redimensionar.
  last_cluster uint16    // Sols si no és root
  dir_cluster  uint16    // Cluster del directori actual
}


// Aquest mètode ompli una nova entrada amb el nom indicat i torna el
// cluster i l'entrada del directory al que apunta. Si és is_dir es
// marca com a directori i si no com a fitxer normal. Si cal es
// redimensionarà el directori.
func (self *_FAT1216_Directory) fNewEntry(

  f      *os.File,
  name   string,
  is_dir bool,
  
) (uint16,[]byte,error) {

  // Comprova el nom
  file_name,err := FAT_GetFileName83 ( name )
  if err != nil {
    return 0,nil,fmt.Errorf ( "Only 8.3 file names supported: %s", err )
  }
  if is_dir && file_name[8]!=' ' {
    return 0,nil,fmt.Errorf ( "Extension is not supported for directories: %s",
    name )
  }
  
  // Busca la primera entrada unused o aplega al final. Ho faig a mà
  // perquè casi que és més fàcil.
  var pos int
  for pos= 0;
  pos < len(self.data) && self.data[pos]!=0x00 && self.data[pos]!=0xe5;
  pos+= 32 {
  }
  var next_pos_modified bool= false
  next_pos :=  pos+32
  if pos >= len(self.data) { // Redimensiona
    if self.is_root {
      return 0,nil,errors.New ( "Root directory is full" )
    } else if err := self.fResize ( f ); err != nil {
      return 0,nil,err
    }
    if pos >= len(self.data) {
      return 0,nil,errors.New ( "Unexpected error occurred while creating"+
        " a new entry" )
    }
    if next_pos < len(self.data) { // Fixa el nou 0
      self.data[next_pos]= 0x00
      next_pos_modified= true
    }
    
  } else if self.data[pos] == 0x00 { // Final del directory
    if next_pos < len(self.data) { // Fixa el nou 0
      self.data[next_pos]= 0x00
      next_pos_modified= true
    }
  }

  // Ompli l'entry
  entry := self.data[pos:next_pos]
  if len(entry)!=32 {
    return 0,nil,errors.New ( "Unexpected error occurred while creating"+
      " a new entry" )
  }
  // --> Nom
  copy ( entry, file_name )
  // --> Attributs
  if is_dir {
    entry[11]= FAT_DIR_DIRECTORY
  } else {
    entry[11]= FAT_DIR_ARCHIVE
  }
  // --> Times.
  entry[13]= 0x00 // Temps de creació???
  t := time.Now ()
  //   ---> Data
  year,month,day := (t.Year()+20)%100,t.Month(),t.Day()
  date := uint16(day&0x1f) | (uint16(month&0xf)<<5) | (uint16(year&0x7f)<<9)
  entry[16],entry[17]= uint8(date),uint8(date>>8)
  entry[18],entry[19]= uint8(date),uint8(date>>8)
  entry[24],entry[25]= uint8(date),uint8(date>>8)
  //   ---> Temps
  hh,mm,ss := t.Hour(),t.Minute(),t.Second()/2
  time := uint16(ss&0x1f) | (uint16(mm&0x3f)<<5) | (uint16(hh&0x1f)<<11)
  entry[14],entry[15]= uint8(time),uint8(time>>8)
  entry[22],entry[23]= uint8(time),uint8(time>>8)
  // --> Cluster
  entry[20],entry[21]= 0x00,0x00
  cluster,err := self.img.fAllocCluster ( f )
  if err != nil { return 0,nil,err }
  entry[26]= uint8(cluster)
  entry[27]= uint8(cluster>>8)
  // --> Size (inicialitze a 0)
  entry[28],entry[29],entry[30],entry[31]= 0x00,0x00,0x00,0x00
  // --> Marca com a modificat
  block_size := len(self.data)/len(self.mod)
  block_ind := pos/block_size
  self.mod[block_ind]= true
  if next_pos_modified {
    block_ind= next_pos/block_size
    self.mod[block_ind]= true
  }
  
  return cluster,entry,nil
  
} // end fNewEntry


// Aquesta funció es crida per a augmentar en 1 el nombre de clusters
// del directori.
func (self *_FAT1216_Directory) fResize(f *os.File) error {

  if self.is_root {
    return errors.New ( "Root directory cannot be resized" )
  }

  // Obté un cluster nou
  new_c,err := self.img.fAllocCluster ( f )
  if err != nil { return err }

  // Redimensiona
  // --> Data
  // IMPORTANT!!! make inicialitza a 0, això es convenient perquè quan
  // el primer byte d'una entrada és 0 vol dir que no hi han més.
  cluster_size,err := self.img.fGetClusterSize ( f )
  if err != nil { return err }
  new_data := make ( []byte, cluster_size + int64(len(self.data)) )
  copy ( new_data, self.data )
  self.data= new_data
  // --> Offsets
  new_offsets := make ( []int64, 1 + len(self.offs) )
  copy ( new_offsets, self.offs )
  self.offs= new_offsets
  // --> Modified
  new_mod := make ( []bool, 1 + len(self.mod) )
  copy ( new_mod, self.mod )
  self.mod= new_mod

  // Actualitza valors
  // --> New offset
  data_offset,err := self.img.fGetDataOffset ( f )
  if err != nil { return err }
  offset := data_offset + int64(new_c-2)*cluster_size
  self.offs[len(self.offs)-1]= offset
  // --> Mod
  self.mod[len(self.mod)-1]= false
  // --> new cluster
  fat,err := self.img.fGetFAT ( f )
  if err != nil { return err }
  fat.write ( self.last_cluster, new_c )
  self.last_cluster= new_c
  
  return nil
  
} // end fResize


func (self *_FAT1216_Directory) fWrite(f *os.File) error {

  block_size := uint64(len(self.data)) / uint64(len(self.mod))
  for i := 0; i < len(self.mod); i++ {
    if self.mod[i] {
      buf := self.data[block_size*uint64(i):block_size*uint64(i+1)]
      if err := self.img.writeBytes ( f, buf, self.offs[i] ); err != nil {
        return fmt.Errorf ( "Error while writing directory entries: %s", err )
      }
      self.mod[i]= false
    }
  }
  
  return nil
  
} // end fWrite


func (self *_FAT1216_Directory) begin() (*_FAT1216_DirectoryIter,error) {

  // Crea
  ret := _FAT1216_DirectoryIter{
    pdir: self,
    it: _FAT_DirectoryIter{
      pos: 0,
      data: self.data,
    },
  }

  // Ignora unused o LFN
  for ; !ret.it.end () &&
    (ret.it.unused () || ret.it.getAttributes () == FAT_DIR_LFN);
  ret.it.next () {
  }

  return &ret,nil
  
} // end begin


func (self *_FAT1216_Directory) Begin() (DirectoryIter,error) {
  return self.begin ()
} // end Begin


func (self *_FAT1216_Directory) MakeDir(name string) (Directory,error) {

  // Obri fitxer
  f,err := os.OpenFile ( self.img.file_name, os.O_RDWR, 0666 )
  if err != nil {
    return nil,fmt.Errorf ( "Unable to open for writing '%s': %s",
    self.img.file_name, err )
  }

  // Comprova si ja existeix
  it,err := self.begin ()
  for ; err == nil && !it.End() && !it.CompareToName ( name ); err= it.Next () {
  }
  if err != nil {
    return nil,err
  } else if !it.End() { // Hem trobat el possible directori
    if it.Type () != DIRECTORY_ITER_TYPE_DIR {
      return nil,fmt.Errorf ( "Trying to create directory '%s' over"+
        " an existing file", name )
    } else {
      return it.GetDirectory ()
    }
  }
  
  // Crea nova entrada
  new_cluster,_,err := self.fNewEntry ( f, name, true )
  if err != nil { return nil,err }
  
  // Data
  t := time.Now ()
  year,month,day := (t.Year()+20)%100,t.Month(),t.Day()
  date := uint16(day&0x1f) | (uint16(month&0xf)<<5) | (uint16(year&0x7f)<<9)
  // Temps
  hh,mm,ss := t.Hour(),t.Minute(),t.Second()/2
  time := uint16(ss&0x1f) | (uint16(mm&0x3f)<<5) | (uint16(hh&0x1f)<<11)

  // Reserva memòria per a un cluster
  cluster_size,err := self.img.fGetClusterSize ( f )
  if err != nil { return nil,err }
  if cluster_size < 32*3 {
    return nil,fmt.Errorf ( "Cluster size is too small: %d", cluster_size )
  }
  dir_data := make ( []byte, cluster_size )
  
  // Create empty dir (3 entrades -> 32*3)
  // --> Entrada 1
  entry1 := dir_data[:32]
  //   --> Nom
  entry1[0]= '.'
  for i := 1; i < 11; i++ { entry1[i]= ' ' }
  //   --> Tipus i altres
  entry1[11]= FAT_DIR_DIRECTORY
  entry1[13]= 0x00
  //   --> Times
  entry1[16],entry1[17]= uint8(date),uint8(date>>8)
  entry1[18],entry1[19]= uint8(date),uint8(date>>8)
  entry1[24],entry1[25]= uint8(date),uint8(date>>8)
  entry1[14],entry1[15]= uint8(time),uint8(time>>8)
  entry1[22],entry1[23]= uint8(time),uint8(time>>8)
  //   --> Cluster, El . apunta al directori nou
  entry1[20],entry1[21]= 0x00,0x00
  entry1[26],entry1[27]= uint8(new_cluster),uint8(new_cluster>>8)
  //   --> Size
  entry1[28],entry1[29],entry1[30],entry1[31]= 0x00,0x00,0x00,0x00
  // --> Entrada 2
  entry2 := dir_data[32:64]
  copy ( entry2, entry1 )
  //   --> Nom
  entry2[1]= '.' // ..
  //   --> Cluster .. apunta al pare
  entry2[26],entry2[27]= uint8(self.dir_cluster),uint8(self.dir_cluster>>8)
  // --> Entrada 3
  entry3 := dir_data[64:96]
  copy ( entry3, entry1 )
  //   --> Nom
  entry3[0]= 0x00 // No hi han més entrades
  
  // Crea directori
  data_offset,err := self.img.fGetDataOffset ( f )
  if err != nil { return nil,err }
  offsets := []int64 { data_offset + int64(new_cluster-2)*cluster_size }
  mods := []bool { true }
  ret := _FAT1216_Directory{
    img: self.img,
    offs: offsets,
    mod: mods,
    data: dir_data,
    is_root: false,
    last_cluster: new_cluster,
    dir_cluster: new_cluster,
  }

  // Escriu en el disc
  if err := ret.fWrite ( f ); err != nil {
    return nil,err
  }
  if err := self.fWrite ( f ); err != nil {
    return nil,err
  }
  if err := self.img.fWriteFAT ( f ); err != nil {
    return nil,err
  }
  
  // Tanca fitxer
  f.Close ()
  
  return &ret,nil
  
} // end MakeDir


func (self *_FAT1216_Directory) GetFileWriter(name string) (FileWriter,error) {

  // Obri fitxer
  f,err := os.OpenFile ( self.img.file_name, os.O_RDWR, 0666 )
  if err != nil {
    return nil,fmt.Errorf ( "Unable to open for writing '%s': %s",
    self.img.file_name, err )
  }
  
  // Cerca si existeix el fitxer
  var file_cluster uint16
  var file_entry []byte
  it,err := self.begin ()
  for ; err == nil && !it.End() && !it.CompareToName ( name ); err= it.Next () {
  }
  if err != nil {
    return nil,err

  } else if it.End() { // Cal crear fitxer nou
    file_cluster,file_entry,err= self.fNewEntry ( f, name, false )
    if err != nil { return nil,err }
    
  } else { // Ja existeix
    file_cluster,file_entry,err= it.fOverwriteFile( f )
    if err != nil { return nil,err }
  }

  // Crea FileWriter
  cluster_size,err := self.img.fGetClusterSize ( f )
  if err != nil { return nil,err }
  if cluster_size == 0 { return nil,errors.New ( "Cluster size is 0" ) }
  data_offset,err := self.img.fGetDataOffset ( f )
  if err != nil { return nil,err }
  ret := _FAT1216_FileWriter{
    f: f,
    img: self.img,
    pdir: self,
    entry: file_entry,
    data_offset: data_offset,
    cluster_size: cluster_size,
    cluster_data: make ( []byte, cluster_size ),
    pos: 0,
    cluster: file_cluster,
    size: 0,
  }
  
  return &ret,nil
  
} // end GetFileWriter


/***************************/
/* FAT12/16 DIRECTORY ITER */
/***************************/

type _FAT1216_DirectoryIter struct {

  pdir *_FAT1216_Directory
  it    _FAT_DirectoryIter
  
}


// Torna Cluster,Entry,Error
func (self *_FAT1216_DirectoryIter) fOverwriteFile(
  f *os.File,
) (uint16,[]byte,error) {

  // Comprovacions
  attr := self.it.getAttributes ()
  if attr==FAT_DIR_LFN ||
    (attr&(FAT_DIR_SYSTEM|FAT_DIR_VOLUME_ID|FAT_DIR_DIRECTORY)) != 0 {
      return 0,nil,errors.New ( "Trying to overwrite a directory"+
        " or special file" )
  }
  if attr&(FAT_DIR_ARCHIVE|FAT_DIR_READ_ONLY) != FAT_DIR_ARCHIVE {
      return 0,nil,errors.New ( "Only non read-only regular files"+
        " can be overwritten" )
  }

  // Llig la taula fat
  fat,err := self.pdir.img.fGetFAT ( f )
  if err != nil { return 0,nil,err }
  
  // Obté cluster i neteja
  file_cluster := self.it.getCluster16 ()
  p := fat.chain ( file_cluster )
  for ; p < fat.badCluster () && p > 1 ; {
    q := p
    p= fat.chain ( p )
    fat.write ( q, 0 ) // Allibera
  }
  fat.write ( file_cluster, fat.badCluster () + 1 ) // Últim cluster

  // Obté entry i actualitza
  pos_entry := self.it.getPosEntry ()
  block_size := len(self.pdir.data)/len(self.pdir.mod)
  block_ind := pos_entry/block_size
  self.pdir.mod[block_ind]= true
  file_entry := self.pdir.data[pos_entry:pos_entry+32]
  // --> Grandària a 0
  file_entry[28]= 0x00
  file_entry[29]= 0x00
  file_entry[30]= 0x00
  file_entry[31]= 0x00
  // --> Times
  t := time.Now ()
  //  --> Data
  year,month,day := (t.Year()+20)%100,t.Month(),t.Day()
  date := uint16(day&0x1f) | (uint16(month&0xf)<<5) | (uint16(year&0x7f)<<9)
  file_entry[18],file_entry[19]= uint8(date),uint8(date>>8)
  file_entry[24],file_entry[25]= uint8(date),uint8(date>>8)
  //  --> Temps
  hh,mm,ss := t.Hour(),t.Minute(),t.Second()/2
  time := uint16(ss&0x1f) | (uint16(mm&0x3f)<<5) | (uint16(hh&0x1f)<<11)
  file_entry[22],file_entry[23]= uint8(time),uint8(time>>8)

  return file_cluster,file_entry,nil
  
} // end fOverwriteFile


func (self *_FAT1216_DirectoryIter) CompareToName(name string) bool {

  // Normalitza nom
  name= strings.ToLower ( name )
  
  // Obté long_name
  long_name := self.it.getLongName ()
  if long_name != "" {
    long_name= strings.ToLower ( strings.TrimSpace ( long_name ) )
  }

  // Comprova long_name
  if long_name == name {
    return true

  } else {

    // Intenta amb el nom curt
    short_name := self.it.getName ()
    short_name= strings.ToLower ( strings.TrimSpace ( short_name ) )
    ext := strings.TrimSpace ( self.it.getExt () )
    if ext != "" {
      short_name+= "." + strings.ToLower ( ext)
    }

    return name == short_name
    
  }

} // end CompareToName
  

func (self *_FAT1216_DirectoryIter) End() bool {
  return self.it.end ()
}


func (self *_FAT1216_DirectoryIter) GetDirectory() (Directory,error) {

  // Comprovacions
  if self.End() {
    return nil,errors.New ( "Trying to obtain a directory from a"+
      " ended iterator" )
  }
  if self.Type () != DIRECTORY_ITER_TYPE_DIR &&
    self.Type () != DIRECTORY_ITER_TYPE_DIR_SPECIAL {
    return nil,errors.New ( "Trying to obtain a directory from a"+
      " non directory entry" )
  }

  // Prepara
  img := self.pdir.img
  cluster := self.it.getCluster16 ()

  // Obri el fitxer
  f,err := os.Open ( img.file_name )
  if err != nil { return nil,err }
  
  // Llig la taula fat
  fat,err := img.fGetFAT ( f )
  if err != nil { return nil,err }

  // Comprovació inicial
  if cluster >= fat.badCluster () {
    return nil,fmt.Errorf ( "Trying to read a FAT12/16 directory from an"+
      " invalid cluster number: %X", cluster )
  }
  
  // Calcula nombre de clusters
  num,tmpc := 0,cluster
  for ; tmpc < fat.badCluster (); {
    if tmpc == 0 || tmpc == 1 {
      return nil,fmt.Errorf ( "%d is a reserved cluster", tmpc )
    } else if tmpc >= fat.length () {
      return nil,fmt.Errorf ( "Cluster %d is out of bounds", tmpc )
    } else {
      num+= 1
      tmpc= fat.chain ( tmpc )
    }
  }
  if tmpc == fat.badCluster () {
    return nil,fmt.Errorf ( "Found bad cluster in a chain started"+
      " in cluster %d", cluster )
  }

  // Crea objecte
  cluster_size,err := img.fGetClusterSize ( f )
  if err != nil { return nil,err }
  if cluster_size == 0 {
    return nil,errors.New ( "Cluster size is 0" )
  }
  nbytes := int64(num)*cluster_size
  data := make ( []byte, nbytes )
  offsets := make ( []int64, num )
  mod := make ( []bool, num )
  
  // Llig clusters
  data_offset,err := img.fGetDataOffset ( f )
  if err != nil { return nil,err }
  var last_cluster uint16= 0
  for p := 0; cluster < fat.badCluster (); p++ {
    buf := data[int64(p)*cluster_size:int64(p+1)*cluster_size]
    offset := data_offset + int64(cluster-2)*cluster_size
    if err := img.readBytes ( f, buf, offset ); err != nil {
      return nil,fmt.Errorf ( "Error while reading directory from"+
        " cluster %d: %s", cluster, err )
    }
    last_cluster= cluster
    cluster= fat.chain ( cluster )
    offsets[p]= offset
    mod[p]= false
  }

  // Crea directori
  ret := _FAT1216_Directory{
    img: img,
    data: data,
    offs: offsets,
    mod: mod,
    is_root: false,
    last_cluster: last_cluster,
    dir_cluster: self.it.getCluster16 (),
  }
  
  // Tanca fitxer
  f.Close ()

  return &ret,nil

} // end GetDirectory


func (self *_FAT1216_DirectoryIter) GetFileReader() (FileReader,error) {

  // Comprovacions
  if self.End() {
    return nil,errors.New ( "Trying to obtain a file reader from a"+
      " ended iterator" )
  }
  if self.Type () != DIRECTORY_ITER_TYPE_FILE {
    return nil,errors.New ( "Trying to obtain a file reader from a"+
      " non file entry" )
  }

  // Prepara
  img := self.pdir.img

  // Obri el fitxer
  f,err := os.Open ( img.file_name )
  if err != nil { return nil,err }
  
  // Calcula valors
  data_offset,err := img.fGetDataOffset ( f )
  if err != nil { return nil,err }
  cluster_size,err := img.fGetClusterSize ( f )
  if err != nil { return nil,err }
  cluster_data := make ( []byte, cluster_size )
  
  // Crea FileReader
  ret := _FAT1216_FileReader{
    f: f,
    img: img,
    it: &self.it,
    data_offset: data_offset,
    cluster_size: cluster_size,
    cluster_data: cluster_data[:],
    pos: 0,
    cluster: self.it.getCluster16 (),
    remain: self.it.getSize (),
  }
  
  // Carrega el següent cluster
  if err := ret.load_next_cluster (); err != nil {
    return nil,err
  }
  
  return &ret,nil
  
} // end GetFileReader


func (self *_FAT1216_DirectoryIter) GetName() string {

  // Obté long_name
  long_name := strings.TrimSpace ( self.it.getLongName () )
  if long_name != "" {
    return long_name
  }

  // Intenta amb el nom curt
  short_name := self.it.getName ()
  short_name= strings.TrimSpace ( short_name )
  ext := strings.TrimSpace ( self.it.getExt () )
  if ext != "" {
    short_name+= "." + ext
  }
  short_name= strings.ToUpper ( short_name )

  return short_name
  
} // end GetName


func (self *_FAT1216_DirectoryIter) List(file io.Writer) error {
  self.it.list ( file )
  return nil
}
  

func (self *_FAT1216_DirectoryIter) Next() error {

  self.it.next ()

  // Ignora entrades buides o inservibles
  for ; !self.it.end () &&
    (self.it.unused () ||
      self.it.getAttributes () == FAT_DIR_LFN);
  self.it.next () {
  }
  
  return nil
  
} // end Next
  

func (self *_FAT1216_DirectoryIter) Type() int {

  attr := self.it.getAttributes ()
  if attr == FAT_DIR_LFN {
    return DIRECTORY_ITER_TYPE_SPECIAL
    
  } else if (attr&(FAT_DIR_VOLUME_ID|FAT_DIR_SYSTEM)) != 0 {
    return DIRECTORY_ITER_TYPE_SPECIAL
    
  } else if (attr&FAT_DIR_DIRECTORY) != 0 {
    name := self.GetName ()
    if name == "." || name == ".." {
      return DIRECTORY_ITER_TYPE_DIR_SPECIAL
    } else {
      return DIRECTORY_ITER_TYPE_DIR
    }
    
  } else if (attr&FAT_DIR_ARCHIVE) != 0 {
    return DIRECTORY_ITER_TYPE_FILE
    
  } else { // Per si de cas
    return DIRECTORY_ITER_TYPE_SPECIAL
    
  }
  
} // end Type
