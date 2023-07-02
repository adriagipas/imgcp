/*
 * Copyright 2023 Adrià Giménez Pastor.
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
  return nil,errors.New("GetRootDirectory: CAL IMPLEMENTAR");
} // end GetRootDirectory


func (self *_IFF_Chunk) fReadHeader(f *os.File) (*_IFF_Header,error) {

  var mem [4]byte
  buf := mem[:]
  
  // Llig tipus
  if err := utils.ReadBytes ( f, self.offset,
    int64(self.length), buf, 0 ); err != nil {
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
    int64(self.length), buf, 4 ); err != nil {
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
    int64(self.length), buf, 8 ); err != nil {
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
