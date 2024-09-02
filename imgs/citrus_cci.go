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
 *  citrus_cci.go - Format citrus CCI.
 *
 */

package imgs

import (
  "errors"
  "fmt"
  "io"
  "strconv"
  
  "github.com/adriagipas/imgcp/citrus"
  "github.com/adriagipas/imgcp/utils"
)


/*******/
/* CCI */
/*******/

// Com que és sols lectura llisc al principi el contingut.
type _CCI struct {
  state *citrus.CCI
}


func newCCI( file_name string ) (*_CCI,error) {

  ret:= _CCI{}
  var err error
  if ret.state,err= citrus.NewCCI ( file_name ); err != nil {
    return nil,err
  }
  
  return &ret,nil
  
} // end newCCI


func (self *_CCI) PrintInfo( file io.Writer, prefix string ) error {

  // Preparació impressió
  P := fmt.Fprintln
  F := fmt.Fprintf

  // Imprimeix
  P(file,prefix, "CTR Cart Image (CCI)")
  P(file,"")
  F(file,"%s Media Id.:     %016x\n",prefix,self.state.Header.MediaID)
  F(file,"%s Title Version: %04x\n",prefix,self.state.Header.TitleVersion)
  F(file,"%s Card Revision: %04x\n",prefix,self.state.Header.CardRevision)
  F(file,"%s Title Id.:     %016x\n",prefix,self.state.Header.TitleID)
  F(file,"%s Version CVer:  %04x\n",prefix,self.state.Header.VersionCVer)
  P(file,prefix, "Partitions:")
  for i:= 0; i < 8; i++ {
    p:= &self.state.Header.Partitions[i]
    if p.Type != citrus.NCSD_PARTITION_TYPE_UNUSED {
      P(file,"")
      F(file,"%s  %d)\n",prefix,i)
      P(file,"")
      F(file,"%s    TYPE:       %s\n", prefix,
        citrus.NCSD_ptype2str ( p.Type ) )
      F(file,"%s    OFFSET:     %016x\n", prefix, p.Offset )
      F(file,"%s    SIZE:       %s\n", prefix,
        utils.NumBytesToStr ( uint64(p.Size) ) )
      P(file,"")
      err := self.fPrintInfoNCCHPartition ( i, file, prefix+"    " )
      if err != nil { return err }
    }
  }
  
  return nil
  
} // end _CCI.PrintInfo


func (self *_CCI) fPrintInfoNCCHPartition(

  ind    int,
  file   io.Writer,
  prefix string,
  
) error {

  // Obté estat
  state,err:= self.state.GetNCCHPartition ( ind )
  if err != nil { return err }

  // NCCH
  ncch,err:= newNCCH ( state )
  if err != nil { return err }

  // Imprimeix
  return ncch.PrintInfo ( file, prefix )
  
} // end fPrintInfoNCCHPartition


// Com que no té subdirectoris i ja està carregat torna el propi
// objecte.
func (self *_CCI) GetRootDirectory() (Directory,error) {
  return self,nil
} // end _CCI.GetRootDirectory


func (self *_CCI) MakeDir(name string) (Directory,error) {
  return nil,errors.New ( "Creation of partitions is not supported" )
} // end Mkdir


func (self *_CCI) GetFileWriter(name string) (FileWriter,error) {
  return nil,errors.New ( "Writing a file not implemented for CCI files" )
}


func (self *_CCI) Begin() (DirectoryIter,error) {

  start:= 0
  for ; start < 8 &&
    self.state.Header.Partitions[start].Type==
      citrus.NCSD_PARTITION_TYPE_UNUSED;
  start++ {
  }
  ret:= _CCI_DirIter{
    state: self.state,
    current: start, // Sempre serà el 0
  }

  return &ret,nil
  
} // end Begin


/**********************/
/* CCI DIRECTORY ITER */
/**********************/

type _CCI_DirIter struct {
  
  state   *citrus.CCI
  current int
  
}


func (self *_CCI_DirIter) CompareToName(name string) bool {
  if num,err := strconv.Atoi ( name ); err == nil && num == self.current {
    return true
  } else {
    return false
  }
} // end CompareToName


func (self *_CCI_DirIter) End() bool {
  return self.current>=8
} // end End


func (self *_CCI_DirIter) GetDirectory() (Directory,error) {

  // Obté estat del NCCH
  state,err:= self.state.GetNCCHPartition ( self.current )
  if err != nil { return nil,err }

  // NCCH
  ncch,err:= newNCCH ( state )
  if err != nil { return nil,err }

  // Crea
  return ncch.GetRootDirectory ()
  
} // end GetDirectory


func (self *_CCI_DirIter) GetFileReader() (FileReader,error) {
  return nil,errors.New ( "_CCI_DirIter.GetFileReader: WTF!!" )
} // end GetFileReader


func (self *_CCI_DirIter) GetName() string {
  return strconv.FormatInt ( int64(self.current), 10 )
} // end GetName


func (self *_CCI_DirIter) List( file io.Writer ) error {

  fmt.Fprintf ( file, "partition  " )

  // Tipus
  fmt.Fprintf ( file, "%s  ",
    citrus.NCSD_ptype2str ( self.state.Header.Partitions[self.current].Type ) )
  
  // Grandària
  size:= utils.NumBytesToStr (
    uint64(self.state.Header.Partitions[self.current].Size) )
  for i := 0; i < 10-len(size); i++ {
    fmt.Fprintf ( file, " " )
  }
  fmt.Fprintf ( file, "%s  ", size )
  
  // Nom
  fmt.Fprintf ( file, "%d\n", self.current )

  return nil
  
} // end List


func (self *_CCI_DirIter) Next() error {

  if !self.End () {
    for self.current++; self.current < 8; self.current++ {
      if self.state.Header.Partitions[self.current].Type != 
        citrus.NCSD_PARTITION_TYPE_UNUSED {
        return nil
      }
    }
  }
  
  return nil
  
} // end Next


func (self *_CCI_DirIter) Remove() error {
  return errors.New ( "Remove file not implemented for CCI images" )
} // end Remove


func (self *_CCI_DirIter) Type() int {
  return DIRECTORY_ITER_TYPE_DIR
} // end Type
