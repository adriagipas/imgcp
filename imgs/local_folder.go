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
 *  local_folder.go - Carpeta local. S'utilitza per a opder copiar
 *                    fitxers entre el disc dur local i les imatges.
 *
 */

package imgs

import (
  "errors"
  "fmt"
  "io"
  "os"
  "path"
)


/****************/
/* LOCAL FOLDER */
/****************/

type _LocalFolder struct {

  file_name string
  
}


func newLocalFolder(file_name string) (*_LocalFolder,error) {

  ret := _LocalFolder{
    file_name: file_name,
  }

  return &ret,nil
  
} // end newLocalFolder


func (self *_LocalFolder) PrintInfo(file io.Writer, prefix string) error {
  
  fmt.Fprintf ( file, "%sLOCAL FOLDER: %s\n", prefix, self.file_name )

  return nil
  
} // end PrintInfo


func (self *_LocalFolder) GetRootDirectory() (Directory,error) {

  ret := _LocalFolder_Directory{
    img: self,
    dir_name: self.file_name,
  }
  
  return &ret,nil
  
} // end GetRootDirectory


/**************************/
/* LOCAL FOLDER DIRECTORY */
/**************************/

type _LocalFolder_Directory struct {

  img      *_LocalFolder
  dir_name string
  
}


func (self *_LocalFolder_Directory) Begin() (DirectoryIter,error) {

  // Obri la carpeta
  f,err := os.Open ( self.dir_name )
  if err != nil { return nil,err }

  // Obté entrades
  entries,err := f.ReadDir ( 0 )
  if err != nil { return nil,err }

  // Tanca
  f.Close ()

  // Retorna
  ret := _LocalFolder_DirectoryIter{
    pdir: self,
    entries: entries,
    pos: 0,
  }
  
  return &ret,nil
  
} // end Begin


func (self *_LocalFolder_Directory) MakeDir(name string) (Directory,error) {

  // Crea directori
  new_path := path.Join ( self.dir_name, name )
  if err := os.MkdirAll ( new_path, 0777 ); err != nil {
    return nil,err
  }

  // Torna
  ret := _LocalFolder_Directory{
    img: self.img,
    dir_name: new_path,
  }

  return &ret,nil
  
} // end MakeDir


func (self *_LocalFolder_Directory) GetFileWriter(
  name string,
) (FileWriter,error) {

  new_path := path.Join ( self.dir_name, name )
  f,err := os.OpenFile ( new_path, os.O_RDWR|os.O_CREATE, 0666 )
  
  return f,err
  
} // end GetFileWriter


/*******************************/
/* LOCAL FOLDER DIRECTORY ITER */
/*******************************/

type _LocalFolder_DirectoryIter struct {

  pdir    *_LocalFolder_Directory  // Directori pare
  entries []os.DirEntry            // Entrades actuals
  pos     int                      // Posició
  
}


func (self *_LocalFolder_DirectoryIter) CompareToName(name string) bool {
  return self.entries[self.pos].Name () == name
} // end CompareToName


func (self *_LocalFolder_DirectoryIter) End() bool {
  return self.pos>=len(self.entries)
} // end End


func (self *_LocalFolder_DirectoryIter) GetDirectory() (Directory,error) {

  new_path := path.Join ( self.pdir.dir_name, self.entries[self.pos].Name () )
  ret := _LocalFolder_Directory{
    img: self.pdir.img,
    dir_name: new_path,
  }

  return &ret,nil
  
} // end GetDirectory


func (self *_LocalFolder_DirectoryIter) GetFileReader() (FileReader,error) {

  path := path.Join ( self.pdir.dir_name, self.entries[self.pos].Name () )
  f,err := os.Open ( path )

  return f,err
  
} // end GetFileReader


func (self *_LocalFolder_DirectoryIter) GetName() string {
  return self.entries[self.pos].Name ()
} // end GetName


func (self *_LocalFolder_DirectoryIter) List(file io.Writer) error {
  return errors.New ( "List operation is not implemented for local folders" )
} // end List


func (self *_LocalFolder_DirectoryIter) Next() error {
  self.pos++
  return nil
} // end Next


func (self *_LocalFolder_DirectoryIter) Remove() error {
  return errors.New ( "Remove operation is not implemented for local folders" )
} // end Remove


func (self *_LocalFolder_DirectoryIter) Type() int {
  
  fmode := self.entries[self.pos].Type ()
  if fmode.IsDir () {
    return DIRECTORY_ITER_TYPE_DIR
  } else if fmode.IsRegular () {
    return DIRECTORY_ITER_TYPE_FILE
  } else {
    return DIRECTORY_ITER_TYPE_SPECIAL
  }
  
} // end Type
