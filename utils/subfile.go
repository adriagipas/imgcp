/*
 * Copyright 2022-2024 Adrià Giménez Pastor.
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
 *  subfile.go - Per a llegir/escriure fitxers que formen part d'un
 *               altre fitxer.
 *
 */

package utils;

import (
  "errors"
  "os"
)


/******************/
/* SUBFILE READER */
/******************/

type SubfileReader struct {

  f           *os.File
  data_offset int64
  data_length int64
  pos         int64 // Posició actual
  
}


func (self *SubfileReader) Close() error {
  self.f.Close ()
  return nil
} // end Close


func (self *SubfileReader) Read(buf []byte) (int,error) {

  // Calcula el que queda
  remain := self.data_length-(self.pos-self.data_offset)
  if remain <= 0 { return 0,nil }

  // Reajusta buffer
  lbuf := int64(len(buf))
  var sbuf []byte
  if lbuf > remain {
    sbuf= buf[:remain]
  } else {
    sbuf= buf
  }

  // Llig
  if err := ReadBytes ( self.f, self.data_offset,
    self.data_length, sbuf, self.pos ); err != nil {
    return -1,err
  }
  ret := len(sbuf)
  self.pos+= int64(ret)

  return ret,nil
  
} // end Read


func (self *SubfileReader) Seek( offset int64, whence int ) (int64,error) {

  if whence != 0 {
    return -1,errors.New ( "SubfileReader.Seek only supports whence=0" )
  }
  if offset < 0 || offset >= self.data_length {
    return -1,errors.New ( "offset out of range" )
  }
  self.pos= self.data_offset + offset

  return offset,nil
  
} // end Seek


func NewSubfileReader(
  
  file_name   string,
  data_offset int64,
  data_length int64,
  
) (*SubfileReader,error) {

  // Obri fitxer.
  f,err := os.Open ( file_name )
  if err != nil { return nil,err }

  // Crea SubfileReader
  ret := SubfileReader{
    f: f,
    data_offset: data_offset,
    data_length: data_length,
    pos: data_offset,
  }

  return &ret,nil
  
} // end NewSubfileReader
