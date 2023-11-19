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
 *  iso9660.go - Implementa el sistema de fitxers ISO-9660.
 *
 */

package imgs

import (
  "errors"
  "fmt"
  "io"
  
  "github.com/adriagipas/imgcp/cdread"
  "github.com/adriagipas/imgcp/utils"
)




/***********/
/* ISO9660 */
/***********/


type _ISO_9660 struct {

  iso *cdread.ISO
  
}


func newISO_9660( cd cdread.CD, session int, track int ) (*_ISO_9660,error) {
  
  ret:= _ISO_9660{}
  var err error
  ret.iso,err= cdread.ReadISO ( cd, session, track )
  if err != nil { return nil,err }

  return &ret,nil
  
} // end newISO_9660


func newISO_9660_from_filename( file_name string ) (*_ISO_9660,error) {

  cd,err:= cdread.Open ( file_name )
  if err != nil { return nil,err }
  info:= cd.Info ()
  if len(info.Sessions)>1 || len(info.Tracks)>1 {
    return nil,fmt.Errorf ( "'%s' is not a ISO 9660 image file", file_name )
  }

  return newISO_9660 ( cd, 0, 0 )
  
} // end newISO_9660_from_filename


func (self *_ISO_9660) PrintInfo( file io.Writer, prefix string ) error {

  // Preparació
  P:= func(args... any) {
    fmt.Fprint ( file, prefix )
    fmt.Fprintln ( file, args... )
  }
  F:= func(format string,args... any) {
    fmt.Fprint ( file, prefix )
    fmt.Fprintf ( file, format, args... )
  }
  PrintDateTime:= func(field_name string,dt *cdread.ISO_DateTime) {
    if !dt.Empty {
      F("%s%s/%s/%s (%s:%s:%s.%s GMT %d)\n",
        field_name,
        dt.Day, dt.Month, dt.Year,
        dt.Hour, dt.Minute, dt.Second, dt.HSecond,
        dt.GMT)
    }
  }
  
  // Imprimeix
  P("ISO 9660")
  P("")
  F("Version:                       %d\n",self.iso.PrimaryVolume.Version)
  if len(self.iso.PrimaryVolume.SystemIdentifier)>0 {
    F("System Identifier:             %s\n",
      self.iso.PrimaryVolume.SystemIdentifier)
  }
  if len(self.iso.PrimaryVolume.VolumeIdentifier)>0 {
    F("Volume Identifier:             %s\n",
      self.iso.PrimaryVolume.VolumeIdentifier)
  }
  F("Volume Space Size:             %d (logical blocks)\n",
    self.iso.PrimaryVolume.VolumeSpaceSize)
  F("Volume Set Size:               %d (disks)\n",
    self.iso.PrimaryVolume.VolumeSetSize)
  F("Volume Sequence Number:        %d\n",
    self.iso.PrimaryVolume.VolumeSequenceNumber)
  F("Logical Block Size:            %d\n",
    self.iso.PrimaryVolume.LogicalBlockSize)
  if len(self.iso.PrimaryVolume.VolumeSetIdentifier)>0 {
    F("Volume Set Identifier:         %s\n",
      self.iso.PrimaryVolume.VolumeSetIdentifier)
  }
  if len(self.iso.PrimaryVolume.PublisherIdentifier)>0 {
    F("Publisher Identifier:          %s\n",
      self.iso.PrimaryVolume.PublisherIdentifier)
  }
  if len(self.iso.PrimaryVolume.DataPreparerIdentifier)>0 {
    F("Data Preparer Identifier:      %s\n",
      self.iso.PrimaryVolume.DataPreparerIdentifier)
  }
  if len(self.iso.PrimaryVolume.ApplicationIdentifier)>0 {
    F("Application Identifier:        %s\n",
      self.iso.PrimaryVolume.ApplicationIdentifier)
  }
  if len(self.iso.PrimaryVolume.CopyrightFileIdentifier)>0 {
    F("Copyright File Identifier:     %s\n",
      self.iso.PrimaryVolume.CopyrightFileIdentifier)
  }
  if len(self.iso.PrimaryVolume.AbstractFileIdentifier)>0 {
    F("Abstract File Identifier:      %s\n",
      self.iso.PrimaryVolume.AbstractFileIdentifier)
  }
  if len(self.iso.PrimaryVolume.BiblioFileIdentifier)>0 {
    F("Bibliographic File Identifier: %s\n",
      self.iso.PrimaryVolume.BiblioFileIdentifier)
  }
  PrintDateTime("Volume Creation:               ",
    &self.iso.PrimaryVolume.VolumeCreation)
  PrintDateTime("Volume Modification:           ",
    &self.iso.PrimaryVolume.VolumeModification)
  PrintDateTime("Volume Expiration:             ",
    &self.iso.PrimaryVolume.VolumeExpiration)
  PrintDateTime("Volume Effective:              ",
    &self.iso.PrimaryVolume.VolumeEffective)
  F("File Structure Version:        %d\n",
    self.iso.PrimaryVolume.FileStructureVersion)
  
  P("")
  
  return nil
  
} // end PrintInfo


func (self *_ISO_9660) GetRootDirectory() (Directory,error) {

  ret:= _ISO_9660_Directory{}
  var err error
  ret.dir,err= self.iso.Root ()
  if err != nil { return nil,err }
  
  return &ret,nil
  
} // end GetRootDirectory


/*************/
/* DIRECTORY */
/*************/

type _ISO_9660_Directory struct {
  dir *cdread.ISO_Directory
}


func (self *_ISO_9660_Directory) Begin() (DirectoryIter,error) {


  tmp,err:= self.dir.Begin ()
  if err != nil { return nil,err }
  ret:= _ISO_9660_DirIter{
    ISO_DirectoryIter : *tmp,
  }

  return &ret,nil
  
} // end Begin


func (self *_ISO_9660_Directory) MakeDir(name string) (Directory,error) {
  return nil,errors.New ( "Make directory not implemented for ISO 9660"+
    " image files")
} // end MakeDir


func (self *_ISO_9660_Directory) GetFileWriter(name string) (FileWriter,error) {
  return nil,errors.New ( "Writing a file not implemented for ISO 9660"+
    " image files")
} // end GetFileWriter


type _ISO_9660_DirIter struct {
  cdread.ISO_DirectoryIter
}


func (self *_ISO_9660_DirIter) CompareToName(name string) bool {
  return name == self.GetName ()
} // end CompareToName


func (self *_ISO_9660_DirIter) GetDirectory() (Directory,error) {

  ret:= _ISO_9660_Directory{}
  var err error
  ret.dir,err= self.ISO_DirectoryIter.GetDirectory ()
  if err != nil { return nil,err }
  
  return &ret,nil
  
} // end GetDirectory


func (self *_ISO_9660_DirIter) GetFileReader() (FileReader,error) {
  return self.ISO_DirectoryIter.GetFileReader ()
} // end GetFileReader


func (self *_ISO_9660_DirIter) GetName() string {
  return self.Id ()
} // end GetName


func (self *_ISO_9660_DirIter) List( file io.Writer ) error {

  P:= func(args... any) {
    fmt.Fprint ( file, args... )
  }
  F:= func(format string,args... any) {
    fmt.Fprintf ( file, format, args... )
  }
  
  // Flags.
  flags:= self.Flags ()
  if (flags&cdread.FILE_FLAGS_DIRECTORY) != 0 { P("d") } else { P("-") }
  if (flags&cdread.FILE_FLAGS_EXISTENCE) != 0 { P("h") } else { P("-") }
  if (flags&cdread.FILE_FLAGS_ASSOCIATED_FILE) != 0 { P("s") } else { P("-") }
  P("  ")

  // Grandària
  size := utils.NumBytesToStr ( uint64(self.Size ()) )
  for i := 0; i < 10-len(size); i++ {
    P(" ")
  }
  P(size,"  ")

  // Date
  dt:= self.DateTime ()
  if !dt.Empty {
    F("%02d/%02d/%04d  ",dt.Day,dt.Month,dt.Year)
  } else {
    P("??/??/????  ")
  }

  // Time
  if !dt.Empty {
    F("%02d:%02d:%02d (GMT %-02d)  ",dt.Hour,dt.Minute,dt.Second,dt.GMT)
  } else {
    P("??:??:?? (GMT  ??)  ")
  }

  // Nom
  P(self.GetName ())

  P("\n")

  return nil
  
} // end List


func (self *_ISO_9660_DirIter) Remove() error {
  return errors.New ( "Remove file not implemented for ISO 9660 images" )
} // end Remove


func (self *_ISO_9660_DirIter) Type() int {

  var ret int
  flags:= self.Flags ()
  if (flags&cdread.FILE_FLAGS_DIRECTORY) != 0 {
    name:= self.GetName ()
    if name == "." || name == ".." {
      ret= DIRECTORY_ITER_TYPE_DIR_SPECIAL
    } else {
      ret= DIRECTORY_ITER_TYPE_DIR
    }
  } else {
    ret= DIRECTORY_ITER_TYPE_FILE
  }

  return ret
  
} // end Type
