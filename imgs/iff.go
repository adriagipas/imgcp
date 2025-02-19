/*
 * Copyright 2023-2025 Adrià Giménez Pastor.
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
 *  iff.go - Implementa suport per a fitxers "Interchange Format Files" (IFF).
 *
 */

package imgs

import (
  "errors"
  "fmt"
  "io"
  "os"
  "strconv"
  "strings"

  "github.com/adriagipas/imgcp/utils"
)


/*******/
/* IFF */
/*******/

// Segueix una aproximació lazzy.
type _IFF_Chunk struct {
  
  file_name string
  offset    int64  // Offset primer byte
  length    uint64 // Grandària en bytes
  
}


func newIFFChunk(file_name string, offset int64, length uint64) (*_IFF_Chunk){
  ret := _IFF_Chunk {
    file_name: file_name,
    offset: offset,
    length: length,
  }
  return &ret
} // end newIFFChunk


func newIFF(file_name string) (*_IFF_Chunk,error) {

  // Obté grandària
  f,err := os.Open ( file_name )
  if err != nil { return nil,err }
  info,err := f.Stat ()
  if err != nil { return nil,err }
  f.Close ()

  // Obté imatge
  return newIFFChunk ( file_name, 0, uint64(info.Size ()) ),nil
  
} // end newIFF


func (self *_IFF_Chunk) PrintInfo(file io.Writer, prefix string) error {

  // Obté informació capçalera
  f,err := os.Open ( self.file_name )
  if err != nil { return err }
  header,err := self.fReadHeader ( f )
  if err != nil { return err }
  f.Close ()

  // Imprimeix informació.
  var type_name string
  switch header.type_iff {
  case _IFF_FORM:
    type_name= "FORM"
  case _IFF_CAT:
    type_name= "CAT"
  case _IFF_LIST:
    type_name= "LIST"
  case _IFF_PROP:
    type_name= "PROP"
  }
  fmt.Fprintf ( file, "%sInterchange Format File (%s)\n", prefix, type_name )
  fmt.Fprintf ( file, "%sSize (bytes): %d\n", prefix, header.nbytes )
  fmt.Fprintf ( file, "%sIdentifier:   %c%c%c%c\n", prefix,
    header.id[0], header.id[1], header.id[2], header.id[3] )
  
  return nil
  
} // end PrintInfo


func (self *_IFF_Chunk) GetRootDirectory() (Directory,error) {

  // Obté informació capçalera
  f,err := os.Open ( self.file_name )
  if err != nil { return nil,err }
  header,err := self.fReadHeader ( f )
  if err != nil { return nil,err }
  f.Close ()
  
  // Crea Directory
  ret := _IFF_Directory{
    img: self,
    offset: self.offset+12, // Descarta capçalera
    nbytes: header.nbytes-4, // Descarta Type
  }
  
  return &ret,nil
  
} // end GetRootDirectory


func (self *_IFF_Chunk) fReadHeader(f *os.File) (*_IFF_Header,error) {

  var mem [4]byte
  buf := mem[:]
  
  // Llig tipus
  if err := utils.ReadBytes ( f, self.offset,
    int64(self.length), buf, self.offset ); err != nil {
    return nil,fmt.Errorf ( "Error while reading IFF chunk type: %s", err )
  }
  var type_iff int
  if buf[0]=='F' && buf[1]=='O' && buf[2]=='R' && buf[3]=='M' {
    type_iff= _IFF_FORM
  } else if buf[0]=='C' && buf[1]=='A' && buf[2]=='T' && buf[3]==' ' {
    type_iff= _IFF_CAT
  } else if buf[0]=='L' && buf[1]=='I' && buf[2]=='S' && buf[3]=='T' {
    type_iff= _IFF_LIST
  } else if buf[0]=='P' && buf[1]=='R' && buf[2]=='O' && buf[3]=='P' {
    type_iff= _IFF_PROP
  } else {
    return nil,fmt.Errorf ( "Unknown IFF chunk type: %c%c%c%c",
      buf[0], buf[1], buf[2], buf[3] )
  }

  // Llig Grandària
  if err := utils.ReadBytes ( f, self.offset,
    int64(self.length), buf, self.offset+4 ); err != nil {
    return nil,fmt.Errorf ( "Error while reading IFF chunk size: %s", err )
  }
  chunk_size := int32(
    (uint32(buf[0])<<24) |
      (uint32(buf[1])<<16) |
      (uint32(buf[2])<<8) |
      uint32(buf[3]))
  if chunk_size < 0 {
    return nil,fmt.Errorf ( "IFF chunk size is negative: %d", chunk_size )
  }

  // Llig identificador
  if err := utils.ReadBytes ( f, self.offset,
    int64(self.length), buf, self.offset+8 ); err != nil {
    return nil,fmt.Errorf ( "Error while reading IFF identifier: %s", err )
  }

  // Crea informació.
  ret := _IFF_Header {
    type_iff: type_iff,
    nbytes: chunk_size,
  }
  copy ( ret.id[:], buf )
  
  return &ret,nil
  
} // end fReadHeader


/**************/
/* IFF HEADER */
/**************/

const _IFF_FORM = 0
const _IFF_LIST = 1
const _IFF_CAT  = 2
const _IFF_PROP = 3

type _IFF_Header struct {

  type_iff int
  nbytes   int32 // No inclou byte padding
  id       [4]byte
  
}


/*************/
/* DIRECTORY */
/*************/

type _IFF_Directory struct {

  img    *_IFF_Chunk // Referència al chunk actual
  offset  int64  // Offset primer byte dades
  nbytes  int32 // Grandària en bytes dades
  
}


// L'offset és el primer byte del chunk
func (self *_IFF_Directory) initIter(
  
  it     *_IFF_DirectoryIter,
  offset int64,
  
) error {

  // Inicialitza
  it.dir= self
  it.offset= offset
  it.num= 0
  
  // Obri fitxer
  f,err := os.Open ( self.img.file_name )
  if err != nil { return err }

  // Llig ID
  if err := self.fReadBytes ( f, it.id[:], it.offset ); err != nil {
    return err
  }

  // Llig grandària bytes sense padding.
  var buf [4]byte
  if err := self.fReadBytes ( f, buf[:], it.offset+4 ); err != nil {
    return err
  }
  it.nbytes= int32(
    (uint32(buf[0])<<24) |
      (uint32(buf[1])<<16) |
      (uint32(buf[2])<<8) |
      uint32(buf[3]))
  if it.nbytes < 0 {
    return fmt.Errorf ( "IFF chunk size is negative: %d", it.nbytes )
  }

  // Tanca i torna valor
  f.Close ()
  
  return nil
  
} // end initIter


// Llig bytes de dins de les dades del "directori" (del chunk actual)

func (self *_IFF_Directory) fReadBytes(
  
  f      *os.File,
  buf    []byte,
  offset int64,
  
) error {
  return utils.ReadBytes ( f, self.offset, int64(self.nbytes), buf, offset )
} // end fReadBytes


func (self *_IFF_Directory) Begin() (DirectoryIter,error) {

  ret := _IFF_DirectoryIter{}
  if err:= self.initIter ( &ret, self.offset ); err != nil {
    return nil,err
  }

  return &ret,nil
  
} // end Begin


func (self *_IFF_Directory) MakeDir(name string) (Directory,error) {
  return nil,errors.New ( "Make directory not implemented for Interchange"+
    " Format Files (IFF)")
} // end MakeDir


func (self *_IFF_Directory) GetFileWriter(
  name string,
) (utils.FileWriter,error) {
  return nil,errors.New ( "Writing a file not implemented for Interchange"+
    " Format Files (IFF)")
} // end GetFileWriter


/******************/
/* DIRECTORY ITER */
/******************/

type _IFF_DirectoryIter struct {

  dir    *_IFF_Directory
  offset int64 // Offset on comença el ID del chunk actual
  id     [4]byte
  nbytes int32 // Bytes del chunk actual, no inclou padding
  num    int64 // Número d'entrada
  
}


func (self *_IFF_DirectoryIter) CompareToName(name string) bool {
  return strings.ToLower ( name ) == self.GetName ()
} // end CompareToName


func (self *_IFF_DirectoryIter) End() bool {

  end := self.dir.offset + int64(self.dir.nbytes)
  return self.offset >= end
  
} // end End


func (self *_IFF_DirectoryIter) GetDirectory() (Directory,error) {

  chunk := newIFFChunk (
    self.dir.img.file_name,
    self.offset,
    uint64(self.nbytes) + 8 )

  return chunk.GetRootDirectory ()
  
} // end GetDirectory


func (self *_IFF_DirectoryIter) GetFileReader() (utils.FileReader,error) {
  return utils.NewSubfileReader (
    self.dir.img.file_name,
    self.offset + 8,
    int64(self.nbytes) )
} // end GetFileReader


func (self *_IFF_DirectoryIter) GetName() string {
  return "e" + strconv.FormatInt ( self.num, 10 )
} // end GetName


func (self *_IFF_DirectoryIter) List(file io.Writer) error {

  P := func(args... any) {
    fmt.Fprint ( file, args... )
  }
  F := func(format string,args... any) {
    fmt.Fprintf ( file, format, args... )
  }
  
  // És o no directori
  it_type := self.Type ()
  if it_type==DIRECTORY_ITER_TYPE_DIR { P("d") } else { P("-") }

  P("  ")

  // Grandària
  size := utils.NumBytesToStr ( uint64(self.nbytes) )
  for i := 0; i < 10-len(size); i++ {
    P(" ")
  }
  P(size,"  ")

  // Tipus
  F("[%c%c%c%c]  ",self.id[0],self.id[1],self.id[2],self.id[3])

  // Nom
  F("e%d",self.num)
  
  P("\n")

  return nil
  
} // end List


func (self *_IFF_DirectoryIter) Next() error {

  // Següent offset
  new_offset := self.offset + 8 + int64(self.nbytes)
  if (self.nbytes&0x1) != 0 { new_offset++ } // Padding
  old_num := self.num

  // Carrega valors si no és end
  end := self.dir.offset + int64(self.dir.nbytes)
  if new_offset < end {
    if err := self.dir.initIter ( self, new_offset ); err != nil {
      return err
    }
  } else {
    self.offset= new_offset
  }
  
  // Incrementa comptador
  self.num= old_num + 1

  return nil
  
} // end Next


func (self *_IFF_DirectoryIter) Remove() error {
  return errors.New ( "Remove file not implemented for Interchange"+
    " Format Files (IFF)")
} // end Remove


func (self *_IFF_DirectoryIter) Type() int {

  id := self.id[:]
  if (id[0]=='F' && id[1]=='O' && id[2]=='R' && id[3]=='M') ||
    (id[0]=='C' && id[1]=='A' && id[2]=='T' && id[3]==' ') ||
    (id[0]=='L' && id[1]=='I' && id[2]=='S' && id[3]=='T') ||
    (id[0]=='P' && id[1]=='R' && id[2]=='O' && id[3]=='P') {
    return DIRECTORY_ITER_TYPE_DIR
  } else {
    return DIRECTORY_ITER_TYPE_FILE
  }
  
} // end Type
