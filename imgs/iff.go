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
  "io"
  "os"
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
  return errors.New("PrintInfo: CAL IMPLEMENTAR")
} // end PrintInfo


func (self *_IFF_Chunk) GetRootDirectory() (Directory,error) {
  return nil,errors.New("GetRootDirectory: CAL IMPLEMENTAR");
} // end GetRootDirectory
