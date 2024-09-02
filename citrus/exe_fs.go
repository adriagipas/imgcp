/*
 * Copyright 2024 Adrià Giménez Pastor.
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
 * along with adriagipas/imgcp.  If not, see
 * <https://www.gnu.org/licenses/>.
 */
/*
 *  exe_fs.go - NCCH Executable Filesystem.
 */

package citrus

import (
  "bytes"
  "errors"
  "fmt"
  
  "github.com/adriagipas/imgcp/utils"
)


/*********/
/* TIPUS */
/*********/

type ExeFS_File struct {

  Name   string
  Offset uint32
  Size   uint32
  
}

type ExeFS struct {

  Files []ExeFS_File
  
  file_name  string
  offset     int64
  
}


/************/
/* FUNCIONS */
/************/

// mem ha de ser un slice de 16 bytes
func (self *ExeFS) addFile( mem []byte ) {
  
  file_name:= bytes.TrimRight ( mem[:8], "\000" )
  if len(file_name)>0 {
    offset:= uint32(mem[8]) |
      (uint32(mem[9])<<8) |
      (uint32(mem[10])<<16) |
      (uint32(mem[11])<<24)
    size:= uint32(mem[12]) |
      (uint32(mem[13])<<8) |
      (uint32(mem[14])<<16) |
      (uint32(mem[15])<<24)
    self.Files= append(
      self.Files,
      ExeFS_File{
        Name: string(file_name),
        Offset: offset,
        Size: size,
      },
    )
  }
  
} // end ExeFS.addFile


func newExeFS(
  file_name string,
  offset    int64,
  length    int64,
) (*ExeFS,error) {

  // Obri subfitxer.
  fd,err:= utils.NewSubfileReader ( file_name, offset, length )
  if err != nil { return nil,err }
  defer fd.Close ()
  
  // Llig capçalera.
  var buf [0x200]byte
  n,err:= fd.Read ( buf[:] )
  if err != nil {
    return nil,fmt.Errorf ( "Error while reading ExeFS header: %s", err )
  }
  if n != len(buf) {
    return nil,
      errors.New ( "Error while reading ExeFS header: not enough bytes" )
  }

  // Inicialitza
  ret:= ExeFS{
    Files: nil,
    file_name: file_name,
    offset: offset,
  }
  ret.Files= make([]ExeFS_File,0,10)

  // Afegeix fitxers
  for i:= 0; i < 10; i++ {
    ret.addFile ( buf[i*16:(i+1)*16])
  }
  
  return &ret,nil
  
} // end newExeFS


func (self *ExeFS) Open( file *ExeFS_File ) (*utils.SubfileReader,error) {
  return utils.NewSubfileReader (
    self.file_name,
    0x200 + self.offset + int64(uint64(file.Offset)),
    int64(uint64(file.Size)),
  )
} // end ExeFS.Open


func (self *ExeFS) OpenIndex( index int ) (*utils.SubfileReader,error) {
  return self.Open ( &self.Files[index] )
} // end ExeFS.OpenIndex
