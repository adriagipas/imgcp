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
  "fmt"
  "io"
  
  "github.com/adriagipas/imgcp/cdread"
)




/***********/
/* ISO9660 */
/***********/


type _ISO_9660 struct {

  iso *cdread.ISO
  
}


func newISO_9660( f cdread.TrackReader ) (*_ISO_9660,error) {
  
  ret:= _ISO_9660{}
  var err error
  ret.iso,err= cdread.ReadISO ( f )
  if err != nil { return nil,err }

  return &ret,nil
  
} // end newISO_9660


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
