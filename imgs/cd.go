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
 *  cd.go - CD-Rom image.
 *
 */

package imgs

import (
  "errors"
  "io"
  
  "github.com/adriagipas/imgcp/cdread"
)




/******/
/* CD */
/******/

// Segueix una aproximació lazzy
type _CD struct {

  file_name string
  cd        cdread.CD
  
}


func newCD( file_name string ) (*_CD,error) {

  ret:= _CD{
    file_name : file_name,
  }
  var err error
  if ret.cd,err= cdread.Open ( file_name ); err != nil {
    return nil,err
  }

  return &ret,nil
  
} // newCD


func (self *_CD) PrintInfo( file io.Writer, prefix string ) error {
  return errors.New("TODO - CD.PrintInfo")
} // end PrintInfo


func (self *_CD) GetRootDirectory() (Directory,error) {
  return nil,errors.New("TODO - CD.GetRootDirectory")
} // end GetRootDirectory
