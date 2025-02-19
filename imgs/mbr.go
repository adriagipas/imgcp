/*
 * Copyright 2022-2025 Adrià Giménez Pastor.
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
 *  mbr.go - HDD with master boot record.
 *
 */

package imgs;

import (
  "errors"
  "fmt"
  "io"
  "os"
  "strconv"

  "github.com/adriagipas/imgcp/utils"
)


/*************/
/* CONSTANTS */
/*************/

const SEC_SIZE = 512

// Partition Types
const PTYPE_FAT16  = 0x04
const PTYPE_FAT16B = 0x06


/*******/
/* MBR */
/*******/

// Segueix una aproximació lazzy
type _MBR struct {
  
  file_name  string
  
}


func newMBR(file_name string) *_MBR {
  ret := _MBR {
    file_name : file_name,
    }
  return &ret
} // end newMBR


// Comprova que es una partició vàlida.
func (self *_MBR) checkPath (
  
  path []string,
  
) (int,error) {

  // No es pot accedir a l'arrel en un MBR
  if len(path) == 0 {
    return -1,errors.New ( "'/' is not a valid path for an image containing"+
      " multiples partitions; please specify the number of a partition,"+
      " p.e.: /0" )
  }

  // Comprova si és un número de partició
  num,err := strconv.Atoi ( path[0] )
  if err != nil {
    return -1,fmt.Errorf ( "'%s' is not a partition number", path[0] )
  }
  if num < 0 || num >= 4 {
    return -1,fmt.Errorf ( "Invalid partition number (%d). Valid partition"+
      " numbers are in range [0,3]", num )
  }

  return num,nil
  
} // end checkPath


// Mètode privat que rep el descriptor del fixer ja obert i llig el
// contingut del MBR.
func (self *_MBR) getContent(f *os.File) (*_MBRContent,error) {

  // Obté info i comprovacions sobre grandària
  info,err := f.Stat ()
  if err != nil { return nil, err }
  if info.IsDir() {
    return nil,fmt.Errorf("'%s' is a directory",self.file_name)
  }
  if info.Size()<=0 || info.Size()%SEC_SIZE != 0 {
    return nil,fmt.Errorf("Wrong size (%d) for '%s'",info.Size(),self.file_name)
  }

  // Llig el MBR
  var buf [SEC_SIZE]byte
  nbytes,err := f.Read ( buf[:] )
  if err != nil { return nil,err }
  if nbytes != SEC_SIZE {
    return nil,fmt.Errorf("Unable to read the MBR from '%s'",self.file_name)
  }
  if buf[0x1FE] != 0x55 || buf[0x1FF] != 0xaa {
    return nil,fmt.Errorf("'%s' does not contain a valid MBR",self.file_name)
  }

  // Crea el contingut
  ret := _MBRContent {}
  ret.partitions[0].read ( buf[0x1be:0x1be+16] )
  ret.partitions[1].read ( buf[0x1ce:0x1ce+16] )
  ret.partitions[2].read ( buf[0x1de:0x1de+16] )
  ret.partitions[3].read ( buf[0x1ee:0x1ee+16] )

  // Comprova grandària particions, i si són absurdes invalida
  // partició.
  for i := 0; i < 4; i++ {
    pe := &ret.partitions[i]
    if pe.lba == 0 ||
      int64(uint64(pe.lba+pe.num_sectors)*512) > info.Size() {
      pe.valid= false
    }
  }
  
  return &ret,nil
  
} // end getContent


func (self *_MBR) PrintInfo(file io.Writer, prefix string) error {

  // Obté continguts
  f,err := os.Open ( self.file_name )
  if err != nil { return err }
  cont,err := self.getContent ( f )
  if err != nil { return err }
  
  // Preparació impressió
  P := fmt.Fprintln
  F := fmt.Fprintf
  
  // Imprimeix
  P(file,prefix, "Image with Master Boot Record (MBR)")
  P(file,"")
  P(file,prefix, "Partitions:")
  for i := 0; i < 4; i++ {
    e := &cont.partitions[i]
    if e.valid {
      P(file,"")
      F(file,"%s  %d)\n",prefix,i)
      P(file,"")
      F(file,"%s    BOOT:         ",prefix)
      if e.active {
        P(file,"Yes")
      } else {
        P(file,"No")
      }
      F(file,"%s    TYPE:         %s\n", prefix, ptype2str ( e.ptype ) )
      F(file,"%s    NUM. SECTORS: %d (%s)\n",
        prefix,e.num_sectors,
        utils.NumBytesToStr ( uint64(e.num_sectors)*SEC_SIZE ))
      F(file,"%s    LBA:          %08Xh\n",prefix, e.lba)
      F(file,"%s    FIRST SECTOR: %s\n",prefix,e.first_sector.toString ())
      F(file,"%s    LAST SECTOR:  %s\n",prefix,e.last_sector.toString ())
      P(file,"")
      err := self.fPrintInfoPartition ( f, e, file, prefix+"    " )
      if err != nil { return err }
    }
  }
  
  // Tanca
  f.Close ()
  
  return nil
  
} // end PrintInfo


func (self *_MBR) fPrintInfoPartition(

  f      *os.File,
  pe     *_PartitionEntry,
  file   io.Writer,
  prefix string,
  
) error {

  // Preparació
  offset := int64(pe.lba)*SEC_SIZE
  length := uint64(pe.num_sectors)*SEC_SIZE
  
  // Imprimeix
  switch pe.ptype {
    
  case PTYPE_FAT16, PTYPE_FAT16B:
    img,err := newSubimgFAT16 ( self.file_name, offset, length )
    if err != nil { return err }
    if err := img.fPrintInfo ( f, file, prefix ); err != nil {
      return err
    }

  default:
    return fmt.Errorf ( "Unknown partition type %02X", pe.ptype )
  }
  
  return nil
  
} // end printInfoPartition


func (self *_MBR) GetRootDirectory() (Directory,error) {

  // Obté continguts
  f,err := os.Open ( self.file_name )
  if err != nil { return nil,err }
  cont,err := self.getContent ( f )
  if err != nil { return nil,err }

  // Crea
  ret := _MBR_Directory{
    img: self,
    content: cont,
  }
  
  // Tanca
  f.Close ()
  
  return &ret,nil
  
} // end GetRootDirectory


/***************/
/* MBR CONTENT */
/***************/

type _MBRContent struct {
  partitions [4]_PartitionEntry // Entrades paritions
}


/*******/
/* CHS */
/*******/

type _CHS struct {
  C uint16
  H uint8
  S uint8
}

func (chr *_CHS) toString() string {
  return fmt.Sprintf ( "C:%04d H:%02d S:%02d", chr.C, chr.H, chr.S )
}


/*******************/
/* PARTITION ENTRY */
/*******************/

type _PartitionEntry struct {
  
  valid        bool   // Indica que és una entrada vàlida
  active       bool   // Indica que és una partició activa
  num_sectors  uint32 // Nombre de sectors
  first_sector _CHS   // Adreça absoluta primer sector
  last_sector  _CHS   // Adreça absoluta últim sector
  lba          uint32 // LBA del primer sector
  ptype        uint8
  
}

// S'utilitza per omplir una PartitionEntry amb dades.
func (pe *_PartitionEntry) read(data []byte) {
  
  // Inicialment és valid
  pe.valid= true
  pe.active= (data[0]&0x80)!=0

  // Grandària
  pe.num_sectors= uint32(data[0xc]) |
    (uint32(data[0xd])<<8) |
    (uint32(data[0xe])<<16) |
    (uint32(data[0xf])<<24)
  if pe.num_sectors == 0 {
    pe.valid= false
    return
  }
  
  // CHS primer sector
  pe.first_sector.C= (uint16(data[2]&0xC0)<<2) | uint16(data[3])
  pe.first_sector.H= data[1]
  pe.first_sector.S= data[2]&0x3F
  if pe.first_sector.S == 0 {
    pe.valid= false
    return
  }
  
  // CHS últim sector
  pe.last_sector.C= (uint16(data[6]&0xC0)<<2) | uint16(data[7])
  pe.last_sector.H= data[5]
  pe.last_sector.S= data[6]&0x3F
  if pe.last_sector.S == 0 {
    pe.valid= false
    return
  }

  // LBA
  pe.lba= uint32(data[0x8]) |
    (uint32(data[0x9])<<8) |
    (uint32(data[0xa])<<16) |
    (uint32(data[0xb])<<24)

  // Partition type
  pe.ptype= data[4]
  
} // end read


// Obté el tipus de la partició
func ptype2str(ptype uint8) string {
  switch ptype {
  case PTYPE_FAT16B:   return "FAT16B    "
  case PTYPE_FAT16:    return "FAT16     "
  default: return fmt.Sprintf("UNK (%02X)",ptype)
  }
} // end ptype2str


/*************/
/* DIRECTORY */
/*************/

type _MBR_Directory struct {

  img     *_MBR
  content *_MBRContent
  
}

func (self *_MBR_Directory) Begin() (DirectoryIter,error) {

  var pos int
  for pos= 0; pos < 4 && !self.content.partitions[pos].valid; pos++ {
  }
  ret := _MBR_DirectoryIter{
    pdir: self,
    p: pos,
  }
  
  return &ret,nil
  
} // end Begin


func (self *_MBR_Directory) MakeDir(name string) (Directory,error) {
  return nil,errors.New ( "Creation of partitions is not supported" )
} // end Mkdir


func (self *_MBR_Directory) GetFileWriter(
  name string,
) (utils.FileWriter,error) {
  return nil,errors.New ( "Files cannot be created outside of a partition" )
}


/******************/
/* DIRECTORY ITER */
/******************/

type _MBR_DirectoryIter struct {

  pdir *_MBR_Directory // Directori pare
  p    int             // Partició actual
  
}


func (self *_MBR_DirectoryIter) CompareToName(name string) bool {
  
  if num,err := strconv.Atoi ( name ); err == nil && num == self.p {
    return true
  } else {
    return false
  }
  
} // end CompareToName


func (self *_MBR_DirectoryIter) End() bool {
  return self.p >= 4
}


func (self *_MBR_DirectoryIter) GetDirectory() (Directory,error) {

  // Comprovacions
  if self.End() {
    return nil,errors.New ( "Trying to obtain a directory from a"+
      " ended iterator" )
  }

  // Preparació
  pe := &self.pdir.content.partitions[self.p]
  offset := int64(pe.lba)*SEC_SIZE
  length := uint64(pe.num_sectors)*SEC_SIZE

  // Obté el directori
  switch pe.ptype {

  case PTYPE_FAT16, PTYPE_FAT16B:
    img,err := newSubimgFAT16 ( self.pdir.img.file_name, offset, length )
    if err != nil { return nil,err }
    return img.GetRootDirectory ()
    
  default:
    return nil,fmt.Errorf ( "Unknown partition type %02X", pe.ptype )
    
  }
  
} // end GetDirectory


func (self *_MBR_DirectoryIter) GetFileReader() (utils.FileReader,error) {
  return nil,errors.New ( "A partition cannot be accessed as a file" )
} // end GetFileReader


func (self *_MBR_DirectoryIter) GetName() string {
  return strconv.FormatInt ( int64(self.p), 10 )
} // end GetName


func (self *_MBR_DirectoryIter) List(file io.Writer) error {

  fmt.Fprintf ( file, "partition  " )

  // Tipus
  fmt.Fprintf ( file, "%s  ",
    ptype2str ( self.pdir.content.partitions[self.p].ptype ))
  
  // Grandària
  p_secs := uint64(self.pdir.content.partitions[self.p].num_sectors)
  size := utils.NumBytesToStr ( p_secs*SEC_SIZE )
  for i := 0; i < 10-len(size); i++ {
    fmt.Fprintf ( file, " " )
  }
  fmt.Fprintf ( file, "%s  ", size )
  
  // Nom
  fmt.Fprintf ( file, "%d\n", self.p )

  return nil
  
} // end List


func (self *_MBR_DirectoryIter) Next() error {
  
  self.p++
  for ; self.p < 4 && !self.pdir.content.partitions[self.p].valid; self.p++ {
  }
  
  return nil
  
} // end Next


func (self *_MBR_DirectoryIter) Remove() error {
  return errors.New ( "Partitions cannot be removed" )
}


func (self *_MBR_DirectoryIter) Type() int {
  return DIRECTORY_ITER_TYPE_DIR_SPECIAL
}
