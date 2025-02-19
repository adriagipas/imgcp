/*
 * Copyright 2025 Adrià Giménez Pastor.
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
 *  stfs.go - Secure Transacted File System.
 *
 */

package imgs

import (
  "bytes"
  "errors"
  "fmt"
  "io"
  "strconv"
  "strings"
  
  "github.com/adriagipas/imgcp/x360"
  "github.com/adriagipas/imgcp/utils"
)


/********/
/* STFS */
/********/

// Com que és sols lectura llisc al principi el contingut de les
// metadades.
type _STFS struct {
  state *x360.STFS
}


func newSTFS( file_name string ) (*_STFS,error) {
  
  ret:= _STFS{}
  var err error
  if ret.state,err= x360.NewSTFS ( file_name ); err != nil {
    return nil,err
  }
  
  return &ret,nil
  
} // end newSTFS


func (self *_STFS) PrintInfo( file io.Writer, prefix string ) error {

  // Preparació
  P:= func(args... any) {
    fmt.Fprint ( file, prefix )
    fmt.Fprintln ( file, args... )
  }
  F:= func(format string,args... any) {
    fmt.Fprint ( file, prefix )
    fmt.Fprintf ( file, format, args... )
  }
  PrintBytes:= func(title string, data []byte) {
    F("%s ",title)
    for i,v:= range data {
      if i%16 == 0 && i > 0 {
        fmt.Fprint ( file, "\n" )
        fmt.Fprint ( file, prefix )
        for i= 0; i < len(title)+1; i++ {
          fmt.Fprint ( file, " " )
        }
      }
      fmt.Fprintf ( file, "%02x ", uint8(v) )
    }
    fmt.Fprint ( file, "\n" )
  }
  
  P("Secure Transacted File System (STFS)")
  P("")
  F("Type:                                  %s\n",self.state.Type())
  if self.state.Header.Type == x360.STFS_TYPE_CONS {
    PrintBytes("Certificate Owner Console ID:         ",
      self.state.Header.CertOwnConsoleID[:])
    F("Certificate Owner Console Part Number: %s\n",
      self.state.Header.CertOwnConsolePartNumber)
    F("Certificate Owner Console Type:        %s\n",
      self.state.CertOwnConsoleType())
    F("Certificate Date of Generation:        %s\n",
      self.state.Header.CertDateGeneration)
    PrintBytes("Public Exponent:                      ",
      self.state.Header.PublicExponent[:])
    PrintBytes("Public Modulus:                       ",
      self.state.Header.PublicModulus[:])
    PrintBytes("Certificate Signature:                ",
      self.state.Header.CertSignature[:])
    PrintBytes("Signature:                            ",
      self.state.Header.Signature[:])
  } else {
    PrintBytes("Package Signature:                    ",
      self.state.Header.PackageSignature[:])
  }
  P("")
  PrintBytes("Content ID:                           ",
    self.state.Metadata.ContentID[:])
  F("Header Size:                           %d\n",
    self.state.Metadata.HeaderSize)
  F("Content Type:                          %s\n",self.state.ContentType())
  F("Metadata Version:                      %d\n",
    self.state.Metadata.MetadataVersion)
  if self.state.Metadata.ContentSize > 0 {
    F("Content Size:                          %s\n",
      utils.NumBytesToStr(uint64(self.state.Metadata.ContentSize)))
  }
  F("Media ID:                              %08x\n",self.state.Metadata.MediaID)
  F("Version:                               %d\n",self.state.Metadata.Version)
  F("Base Version:                          %d\n",
    self.state.Metadata.BaseVersion)
  F("Title ID:                              %08x\n",self.state.Metadata.TitleID)
  F("Platform:                              %s\n",self.state.Platform())
  F("Executable Type:                       %02x\n",
    self.state.Metadata.ExecutableType)
  F("Disc Number:                           %d\n",
    self.state.Metadata.DiscNumber)
  F("Disc in Set:                           %d\n",
    self.state.Metadata.DiscInSet)
  F("Save Game ID:                          %08x\n",
    self.state.Metadata.SaveGameID)
  PrintBytes("Console ID:                           ",
    self.state.Metadata.ConsoleID[:])
  PrintBytes("Profile ID:                           ",
    self.state.Metadata.ProfileID[:])
  if self.state.Metadata.DataFileCount > 0 {
    F("Data File Count:                       %d\n",
      self.state.Metadata.DataFileCount)
    F("Data File Combined Size:               %s\n",
      utils.NumBytesToStr(uint64(self.state.Metadata.DataFileCombSize)))
  }
  F("Volume Descriptor Type:                %s\n",
    self.state.DescriptorType())
  PrintBytes("Device ID:                            ",
    self.state.Metadata.DeviceID[:])
  F("Transfer Flags:                        %02x\n",
    self.state.Metadata.TransferFlags)
  if self.state.Metadata.MetadataVersion == 2 {
    PrintBytes("Series ID:                            ",
      self.state.Metadata.SeriesID[:])
    PrintBytes("Season ID:                            ",
      self.state.Metadata.SeasonID[:])
    F("Season Number:                         %d\n",
      self.state.Metadata.SeasonNumber)
    F("Episode Number:                        %d\n",
      self.state.Metadata.EpisodeNumber)
  }
  if self.state.Metadata.PublisherName != "" {
    F("Publisher Name:                        %s\n",
      self.state.Metadata.PublisherName)
  }
  if self.state.Metadata.TitleName != "" {
    F("Title Name:                            %s\n",
      self.state.Metadata.TitleName)
  }
  P("Display Name / Description:")
  for i:= 0; i < 12; i++ {
    if self.state.Metadata.DisplayName[i] != "" &&
      self.state.Metadata.DisplayDescription[i] != "" {
      P("")
      if self.state.Metadata.DisplayName[i] != "" {
        F(" - %s\n",self.state.Metadata.DisplayName[i])
      }
      if self.state.Metadata.DisplayDescription[i] != "" {
        F(" - %s\n",self.state.Metadata.DisplayDescription[i])
      }
    }
  }
  P("")
  
  return nil
  
} // end _STFS.PrintInfo


// Com que el directori arrel és especial (el gaste per accedir al
// thumbnails).
func (self *_STFS) GetRootDirectory() (Directory,error) {
  return self,nil
} // end _STFS.GetRootDirectory


func (self *_STFS) MakeDir(name string) (Directory,error) {
  return nil,errors.New ( "Creation of volumes is not supported" )
} // end Mkdir


func (self *_STFS) GetFileWriter(name string) (utils.FileWriter,error) {
  return nil,errors.New ( "Writing a file not implemented for STFS files" )
}


func (self *_STFS) Begin() (DirectoryIter,error) {

  ret:= _STFS_RootDirIter{
    state: self.state,
    current: 0,
  }

  return &ret,nil
  
} // end Begin


/*************************/
/* STFS THUMBNAIL READER */
/*************************/

type _STFS_ThumbnailReader struct {
  *bytes.Buffer
}


func (self _STFS_ThumbnailReader) Close() error {
  return nil
}


func newThumbnailReader( data []byte ) _STFS_ThumbnailReader {
  ret:= _STFS_ThumbnailReader{}
  ret.Buffer= bytes.NewBuffer ( data )
  return ret
}


/****************************/
/* STFS ROOT DIRECTORY ITER */
/****************************/

type _STFS_RootDirIter struct {

  state   *x360.STFS
  current int
  
}


func (self *_STFS_RootDirIter) CompareToName(name string) bool {
  
  switch self.current {
  case 0: // Volume
    num,err:= strconv.Atoi ( name )
    return err==nil && num == 0
  case 1: // Thumbnail
    return strings.ToLower ( name ) == "thumbnail"
  case 2: // Title Thumbnail
    return strings.ToLower ( name ) == "title_thumbnail"
  default:
    return false
  }
  
} // end CompareToName


func (self *_STFS_RootDirIter) End() bool {
  return self.current>=3
} // end End


func (self *_STFS_RootDirIter) GetDirectory() (Directory,error) {

  if self.current==0 {
    files,err:= self.state.FileList ()
    if err != nil { return nil,err }
    ret:= _STFS_Directory{
      state: self.state,
      files: files,
      dir: -1, // Root
    }
    return &ret,nil
  } else {
    return nil,errors.New ( "_STFS_RootDirIter.GetDirectory: WTF!!!" )
  }
  
} // end GetDirectory


func (self *_STFS_RootDirIter) GetFileReader() (utils.FileReader,error) {

  switch self.current {
  case 1: // Thumbnail
    return newThumbnailReader ( self.state.Metadata.Thumbnail ),nil
  case 2: // Title Thumbnail
    return newThumbnailReader ( self.state.Metadata.TitleThumbnail ),nil
  default:
    return nil,errors.New ( "_STFS_RootDirIter.GetFileReader: WTF!!" )
  }
  
} // end GetFileReader


func (self *_STFS_RootDirIter) GetName() string {
  
  switch self.current {
  case 0:
    return "0"
  case 1:
    return "thumbnail"
  case 2:
    return "title_thumbail"
  default:
    return "???"
  }
  
} // end GetName


func (self *_STFS_RootDirIter) List( file io.Writer ) error {

  P:= func(args... any) {
    fmt.Fprint ( file, args... )
  }
  
  // Directori
  if self.current == 0 { P("d") } else { P("-") }
  P("  ")

  // Grandària
  var size string
  switch self.current {
  case 1: // Thumbnail
    size= utils.NumBytesToStr ( uint64(len(self.state.Metadata.Thumbnail)) )
  case 2: // Title Thumbnail
    size= utils.NumBytesToStr (
      uint64(len(self.state.Metadata.TitleThumbnail)) )
  default:
    size= ""
  }
  for i := 0; i < 10-len(size); i++ {
    P(" ")
  }
  P(size,"  ")

  // Nom
  P(self.GetName ())
  
  P("\n")
  
  return nil

} // end List


func (self *_STFS_RootDirIter) Next() error {

  if !self.End () {
    self.current++
    if self.current==1 && self.state.Metadata.Thumbnail==nil {
      self.current++
    }
    if self.current==2 && self.state.Metadata.TitleThumbnail==nil {
      self.current++
    }
  }
  
  return nil
  
} // end Next


func (self *_STFS_RootDirIter) Remove() error {
  return errors.New ( "Remove file not implemented for STFS images" )
} // end Remove


func (self *_STFS_RootDirIter) Type() int {
  if self.current == 0 {
    return DIRECTORY_ITER_TYPE_DIR
  } else {
    return DIRECTORY_ITER_TYPE_FILE
  }
} // end Type


/******************/
/* STFS DIRECTORY */
/******************/

type _STFS_Directory struct {

  state *x360.STFS
  files []x360.STFS_FileEntry
  dir   int
  
}


func (self *_STFS_Directory) Begin() (DirectoryIter,error) {
  
  ret:= _STFS_DirIter{
    dir: self,
  }
  for ret.pos= 0;
  ret.pos < len(self.files) && self.files[ret.pos].PathIndicator != self.dir;
  ret.pos++ {
  }
  
  return &ret,nil
  
} // end Begin


func (self *_STFS_Directory) GetFileWriter(
  name string,
) (utils.FileWriter,error) {
  return nil,errors.New ( "Writing a file not implemented for STFS files" )
}


func (self *_STFS_Directory) MakeDir(name string) (Directory,error) {
  return nil,errors.New ( "Creation of volumes is not supported" )
} // end Mkdir


/**********************/
/* STF DIRECTORY ITER */
/**********************/

type _STFS_DirIter struct {
  dir *_STFS_Directory
  pos int
}


func (self *_STFS_DirIter) CompareToName(name string) bool {
  return !self.End () && name==self.dir.files[self.pos].Name
} // end CompareToName


func (self *_STFS_DirIter) End() bool {
  return self.pos>=len(self.dir.files)
} // end End


func (self *_STFS_DirIter) GetDirectory() (Directory,error) {

  if !self.End () && self.dir.files[self.pos].IsDirectory {
    ret:= _STFS_Directory{
      state: self.dir.state,
      files: self.dir.files,
      dir: self.pos,
    }
    return &ret,nil
  } else {
    return nil,errors.New ( "_STFS_DirIter.GetDirectory: WTF!!!" )
  }
  
} // end GetDirectory


func (self *_STFS_DirIter) GetFileReader() (utils.FileReader,error) {
  
  if !self.End () && !self.dir.files[self.pos].IsDirectory {
    file:= &self.dir.files[self.pos]
    return self.dir.state.Open ( file.StartingBlock, file.NumBlocks,
      file.Consecutive, file.Size )
  } else {
    return nil,errors.New ( "_STFS_DirIter.GetFileReader: WTF!!" )
  }
  
} // end GetFileReader


func (self *_STFS_DirIter) GetName() string {
  return self.dir.files[self.pos].Name
} // end GetName


func (self *_STFS_DirIter) List(file io.Writer) error {

  P := func(args... any) {
    fmt.Fprint ( file, args... )
  }
  F := func(format string,args... any) {
    fmt.Fprintf ( file, format, args... )
  }

  entry:= &self.dir.files[self.pos]
  
  // Attributs
  if entry.IsDirectory { P("d") } else { P("-") }
  
  P("  ")
  
  // Grandària
  size := utils.NumBytesToStr ( uint64(entry.Size) )
  for i := 0; i < 10-len(size); i++ {
    P(" ")
  }
  P(size,"  ")
  
  // Date
  F("%s  ",entry.GetUpdateTimestamp())
  
  // Nom
  P(entry.Name)
  
  P("\n")

  return nil
  
} // end List


func (self *_STFS_DirIter) Next() error {
  
  if !self.End () {
    for self.pos++;
    self.pos < len(self.dir.files) &&
      self.dir.files[self.pos].PathIndicator != self.dir.dir;
    self.pos++ {
    }
  }
  
  return nil
  
} // end Next


func (self *_STFS_DirIter) Remove() error {
  return errors.New ( "Remove file not implemented for STFS images" )
} // end Remove


func (self *_STFS_DirIter) Type() int {
  if self.dir.files[self.pos].IsDirectory {
    return DIRECTORY_ITER_TYPE_DIR
  } else {
    return DIRECTORY_ITER_TYPE_FILE
  }
} // end Type
