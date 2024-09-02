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
 * along with adriagipas/imgcp.  If not, see <https://www.gnu.org/licenses/>.
 */
/*
 *  citrus_ncch.go - Format citrus NCCH.
 *
 */

package imgs

import (
  "errors"
  "fmt"
  "io"
  "strings"
  
  "github.com/adriagipas/imgcp/citrus"
  "github.com/adriagipas/imgcp/utils"
)


/********/
/* NCCH */
/********/

const (
  _NCCH_PLAIN = 0
  _NCCH_LOGO  = 1
  _NCCH_EXEFS = 2
  _NCCH_ROMFS = 3
)

// Com que és sols lectura llisc al principi el contingut.
type _NCCH struct {
  state *citrus.NCCH
}


func newNCCH( state *citrus.NCCH ) (*_NCCH,error) {
  
  ret:= _NCCH{
    state: state,
  }
  
  return &ret,nil
  
} // end newCCI


func (self *_NCCH) PrintInfo( file io.Writer, prefix string ) error {

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
  P("Nintendo Content Container Header format")
  P("")
  F("Id.:          %016x",self.state.Header.Id)
  F("Maker Code:   %s",self.state.Header.MakerCode)
  F("Version:      %04x",self.state.Header.Version)
  F("Program Id.:  %016x",self.state.Header.ProgramId)
  F("Product Code: %s",self.state.Header.ProductCode)
  var platform string
  if self.state.Header.Platform == citrus.NCCH_PLATFORM_3DS {
    platform= "3DS"
  } else if self.state.Header.Platform == citrus.NCCH_PLATFORM_NEW3DS {
    platform= "New 3DS"
  } else {
    platform= "Unknown"
  }
  F("Platform:     %s",platform)
  var ftype string
  if self.state.Header.Type == citrus.NCCH_TYPE_CXI {
    ftype= "CXI"
  } else if self.state.Header.Type == citrus.NCCH_TYPE_CFA {
    ftype= "CFA"
  } else {
    ftype= "Unknown"
  }
  F("Type:         %s",ftype)
  P("Flags:\n")
  if (self.state.Header.Flags&citrus.NCCH_FLAGS_EXECUTABLE)!=0 {
    P("  - Executable")
  }
  if (self.state.Header.Flags&citrus.NCCH_FLAGS_DATA)!=0 {
    P("  - Data")
  }
  if (self.state.Header.Flags&citrus.NCCH_FLAGS_SYSTEM_UPDATE)!=0 {
    P("  - System update")
  }
  if (self.state.Header.Flags&citrus.NCCH_FLAGS_MANUAL)!=0 {
    P("  - Manual")
  }
  if (self.state.Header.Flags&citrus.NCCH_FLAGS_TRIAL)!=0 {
    P("  - Trial")
  }
  P("")
  
  return nil
  
} // end _NCCH.PrintInfo


// Com que no té subdirectoris i ja està carregat torna el propi
// objecte.
func (self *_NCCH) GetRootDirectory() (Directory,error) {
  return self,nil
} // end _NCCH.GetRootDirectory


func (self *_NCCH) MakeDir(name string) (Directory,error) {
  return nil,errors.New ( "Make directory not implemented for NCCH files" )
} // end Mkdir


func (self *_NCCH) GetFileWriter(name string) (FileWriter,error) {
  return nil,errors.New ( "Writing a file not implemented for NCCH files" )
}


func (self *_NCCH) Begin() (DirectoryIter,error) {

  var pos int
  if self.state.Header.Plain.Size != 0 {
    pos= _NCCH_PLAIN
  } else if self.state.Header.Logo.Size != 0 {
    pos= _NCCH_LOGO
  } else if self.state.Header.ExeFS.Size != 0 {
    pos= _NCCH_EXEFS
  } else if self.state.Header.RomFS.Size != 0 {
    pos= _NCCH_ROMFS
  } else {
    return nil,errors.New ( "NCCH file without content" )
  }
  ret:= _NCCH_DirIter{
    state: self.state,
    pos: pos,
  }

  return &ret,nil
  
} // end Begin


/**********************/
/* NCCH DIRECTORY ITER */
/**********************/

type _NCCH_DirIter struct {
  
  state *citrus.NCCH
  pos   int
  
}


func (self *_NCCH_DirIter) CompareToName(name string) bool {
  
  tmp:= strings.ToLower ( name )
  switch self.pos {
  case _NCCH_PLAIN:
    return tmp == "plain"
  case _NCCH_LOGO:
    return tmp == "logo"
  case _NCCH_EXEFS:
    return tmp == "exefs"
  case _NCCH_ROMFS:
    return tmp == "romfs"
  default:
    return false
  }
  
} // end CompareToName


func (self *_NCCH_DirIter) End() bool {
  return self.pos>=4
} // end End


func (self *_NCCH_DirIter) GetDirectory() (Directory,error) {

  switch self.pos {
    
  case _NCCH_EXEFS:
    state,err:= self.state.GetExeFS ()
    if err != nil { return nil,err }
    return &_ExeFS{
      state: state, // No deuria ser nil
    },nil
    
  case _NCCH_ROMFS:
    return nil,errors.New ( "GetDirectory ROMFS - TODO" )
    
  default:
    return nil,errors.New ( "_NCCH_DirIter.GetDirectory: WTF!!!" )
    
  }
  
} // end GetDirectory


func (self *_NCCH_DirIter) GetFileReader() (FileReader,error) {

  switch self.pos {
    
  case _NCCH_PLAIN:
    return self.state.GetPlain ()
    
  case _NCCH_LOGO:
    return self.state.GetLogo ()
    
  default:
    return nil,errors.New ( "_NCCH_DirIter.GetFileReader: WTF!!!" )
    
  }
  
} // end GetFileReader


func (self *_NCCH_DirIter) GetName() string {
  switch self.pos {
  case _NCCH_PLAIN:
    return "plain"
  case _NCCH_LOGO:
    return "logo"
  case _NCCH_EXEFS:
    return "exeFS"
  case _NCCH_ROMFS:
    return "romFS"
  default:
    return "???"
  }
} // end GetName


func (self *_NCCH_DirIter) List( file io.Writer ) error {

  P:= func(args... any) {
    fmt.Fprint ( file, args... )
  }
  F:= func(format string,args... any) {
    fmt.Fprintf ( file, format, args... )
  }
  
  // És o no directori
  it_type:= self.Type ()
  if it_type==DIRECTORY_ITER_TYPE_DIR { P("d") } else { P("-") }

  P("  ")

  // Grandària
  var nbytes int64
  switch self.pos {
  case _NCCH_PLAIN:
    nbytes= self.state.Header.Plain.Size
  case _NCCH_LOGO:
    nbytes= self.state.Header.Logo.Size
  case _NCCH_EXEFS:
    nbytes= self.state.Header.ExeFS.Size
  case _NCCH_ROMFS:
    nbytes= self.state.Header.RomFS.Size
  }
  size := utils.NumBytesToStr ( uint64(nbytes) )
  for i := 0; i < 10-len(size); i++ {
    P(" ")
  }
  P(size,"  ")

  // Nom
  F("%s",self.GetName ())
  
  P("\n")

  return nil
  
} // end List


func (self *_NCCH_DirIter) Next() error {

  if !self.End () {
    for self.pos++; self.pos < 4; self.pos++ {
      switch self.pos {
      case _NCCH_PLAIN: // <-- Innecessari
        if self.state.Header.Plain.Size != 0 { return nil }
      case _NCCH_LOGO:
        if self.state.Header.Logo.Size != 0 { return nil }
      case _NCCH_EXEFS:
        if  self.state.Header.ExeFS.Size != 0 { return nil }
      case _NCCH_ROMFS:
        if  self.state.Header.RomFS.Size != 0 { return nil }
      }
    }
  }
  
  return nil
  
} // end Next


func (self *_NCCH_DirIter) Remove() error {
  return errors.New ( "Remove file not implemented for NCCH files" )
} // end Remove


func (self *_NCCH_DirIter) Type() int {
  switch self.pos {
  case _NCCH_PLAIN, _NCCH_LOGO:
    return DIRECTORY_ITER_TYPE_FILE
  default:
    return DIRECTORY_ITER_TYPE_DIR
  }
} // end Type


/*********/
/* EXEFS */
/*********/

// Com que és sols lectura llisc al principi el contingut.
type _ExeFS struct {
  state *citrus.ExeFS
}

func (self *_ExeFS) MakeDir(name string) (Directory,error) {
  return nil,errors.New (
    "Make directory not implemented for NCCH.ExeFS files" )
} // end Mkdir


func (self *_ExeFS) GetFileWriter(name string) (FileWriter,error) {
  return nil,errors.New (
    "Writing a file not implemented for NCCH.ExeFS files" )
}


func (self *_ExeFS) Begin() (DirectoryIter,error) {

  ret:= _ExeFS_DirIter{
    state: self.state,
    pos: 0,
  }

  return &ret,nil
  
} // end Begin


/************************/
/* EXEFS DIRECTORY ITER */
/************************/

type _ExeFS_DirIter struct {
  
  state *citrus.ExeFS
  pos   int
  
}


func (self *_ExeFS_DirIter) CompareToName(name string) bool {
  return name == self.state.Files[self.pos].Name
} // end CompareToName


func (self *_ExeFS_DirIter) End() bool {
  return self.pos>=len(self.state.Files)
} // end End


func (self *_ExeFS_DirIter) GetDirectory() (Directory,error) {
  return nil,errors.New ( "_ExeFS_DirIter.GetDirectory: WTF!!!" )
} // end GetDirectory


func (self *_ExeFS_DirIter) GetFileReader() (FileReader,error) {
  return self.state.OpenIndex ( self.pos )
} // end GetFileReader


func (self *_ExeFS_DirIter) GetName() string {
  return self.state.Files[self.pos].Name
} // end GetName


func (self *_ExeFS_DirIter) List( file io.Writer ) error {

  P:= func(args... any) {
    fmt.Fprint ( file, args... )
  }
  F:= func(format string,args... any) {
    fmt.Fprintf ( file, format, args... )
  }
  
  // És o no directori
  P("-")

  P("  ")

  // Grandària
  size := utils.NumBytesToStr ( uint64(self.state.Files[self.pos].Size) )
  for i := 0; i < 10-len(size); i++ {
    P(" ")
  }
  P(size,"  ")

  // Nom
  F("%s",self.GetName ())
  
  P("\n")
  
  return nil
  
} // end List


func (self *_ExeFS_DirIter) Next() error {

  if !self.End () {
    self.pos++
  }
  
  return nil
  
} // end Next


func (self *_ExeFS_DirIter) Remove() error {
  return errors.New ( "Remove file not implemented for NCCH.ExeFS files" )
} // end Remove


func (self *_ExeFS_DirIter) Type() int {
  return DIRECTORY_ITER_TYPE_FILE
} // end Type
