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
 * along with adriagipas/imgcp.  If not, see
 * <https://www.gnu.org/licenses/>.
 */
/*
 *  read_iso.go - Funcions per llegir tracks de CD en format ISO9660.
 */

package cdread

import (
  "errors"
  "fmt"
  "strings"
)




/****************/
/* PART PRIVADA */
/****************/

const LOGICAL_SECTOR_SIZE = 2048


func parse_date_time( data []byte, dt *ISO_DateTime ) {

  empty:= true
  
  // Year
  if data[0]!='0' || data[1]!='0' || data[2]!='0' || data[3]!='0' {
    dt.Year= string(data[:4])
    empty= false
  }
  // Month
  if data[4]!='0' || data[5]!='0' {
    dt.Month= string(data[4:6])
    empty= false
  }
  // Day
  if data[6]!='0' || data[7]!='0' {
    dt.Day= string(data[6:8])
    empty= false
  }
  // Hour
  if data[8]!='0' || data[9]!='0' {
    dt.Hour= string(data[8:10])
    empty= false
  }
  // Minute
  if data[10]!='0' || data[11]!='0' {
    dt.Minute= string(data[10:12])
    empty= false
  }
  // Second
  if data[12]!='0' || data[13]!='0' {
    dt.Second= string(data[12:14])
    empty= false
  }
  // HSecond
  if data[14]!='0' || data[15]!='0' {
    dt.HSecond= string(data[14:16])
    empty= false
  }
  // GMT Offset
  tmp:= int(data[16])
  if !empty || tmp != 0 {
    dt.GMT= tmp-48
  }
  dt.Empty= empty
  
} // end parse_date_time


func parse_int16_LSB_MSB(data []byte) uint16 {
  return uint16(data[0]) |
    (uint16(data[1])<<8)
} // end parse_int16_LSB_MSB


func parse_int32_LSB_MSB(data []byte) uint32 {
  return uint32(data[0]) |
    (uint32(data[1])<<8) |
    (uint32(data[2])<<16) |
    (uint32(data[3])<<24)
} // end parse_int32_LSB_MSB


func (self *ISO) readVolumeDescriptors() error {

  var buf [LOGICAL_SECTOR_SIZE]byte
  sector,end,num_pv:= int64(0x10),false,0
  for ; !end; sector++ {

    // Prova a llegir
    if err:= self.f.Seek ( sector ); err != nil {
      return err
    }
    if nbytes,err:= self.f.Read ( buf[:] ); err != nil {
      return err
    } else if nbytes != LOGICAL_SECTOR_SIZE {
      return fmt.Errorf ( "failed to read volume descriptor at sector %d",
        sector )
    }

    // Comprova tipus
    switch dtype:= buf[0]; dtype {
    case 0: // Boot record
      return fmt.Errorf ( "TODO - BOOT RECORD!!!" )
    case 1: // Primary volume
      if num_pv == 1 {
        return fmt.Errorf ( "found a second primary volume descriptor"+
          " at sector %d", sector )
      } else {
        num_pv= 1
      }
      if err:= self.readPrimaryVolume ( buf[:] ); err != nil {
        return err
      }
    case 2: // Supplementary volume
      return errors.New ( "supplementary volume descriptor not implemented" )
    case 3: // Volume partition
      return errors.New ( "volume partition descriptor not implemented" )
    case 255:
      end= true
    default:
      return fmt.Errorf ( "unknown volume descriptor type: %d", dtype )
      
    }
    
  }

  return nil
  
} // end readVolumeDescriptors


func (self *ISO) readPrimaryVolume( data []byte ) error {

  // Signatura
  if data[1]!='C' || data[2]!='D' || data[3]!='0' ||
    data[4]!='0' || data[5]!='1' {
    return errors.New ( "Volume descriptor signature 'CD001' not "+
      "found in primary descriptor" )
  }

  // Versió
  self.PrimaryVolume.Version= uint8(data[6])

  // Identificadors
  self.PrimaryVolume.SystemIdentifier=
    strings.TrimRight ( string(data[8:40]), " " )
  self.PrimaryVolume.VolumeIdentifier=
    strings.TrimRight ( string(data[40:72]), " " )

  // Grandàries
  self.PrimaryVolume.VolumeSpaceSize= parse_int32_LSB_MSB ( data[80:88] )
  self.PrimaryVolume.VolumeSetSize= parse_int16_LSB_MSB ( data[120:124] )
  self.PrimaryVolume.VolumeSequenceNumber= parse_int16_LSB_MSB ( data[124:128] )
  self.PrimaryVolume.LogicalBlockSize= parse_int16_LSB_MSB ( data[128:132] )

  // Root directory record
  copy(self.PrimaryVolume.root_dir_record[:],data[156:190])

  // Més identificadors
  self.PrimaryVolume.VolumeSetIdentifier=
    strings.TrimRight ( string(data[190:318]), " " )
  self.PrimaryVolume.PublisherIdentifier=
    strings.TrimRight ( string(data[318:446]), " " )
  self.PrimaryVolume.DataPreparerIdentifier=
    strings.TrimRight ( string(data[446:574]), " " )
  self.PrimaryVolume.ApplicationIdentifier=
    strings.TrimRight ( string(data[574:702]), " " )
  self.PrimaryVolume.CopyrightFileIdentifier=
    strings.TrimRight ( string(data[702:739]), " " )
  self.PrimaryVolume.AbstractFileIdentifier=
    strings.TrimRight ( string(data[739:776]), " " )
  self.PrimaryVolume.BiblioFileIdentifier=
    strings.TrimRight ( string(data[776:813]), " " )

  // Dates
  parse_date_time ( data[813:830], &self.PrimaryVolume.VolumeCreation )
  parse_date_time ( data[830:847], &self.PrimaryVolume.VolumeModification )
  parse_date_time ( data[847:864], &self.PrimaryVolume.VolumeExpiration )
  parse_date_time ( data[864:881], &self.PrimaryVolume.VolumeEffective )

  // FileStructureVersion
  self.PrimaryVolume.FileStructureVersion= uint8(data[881])
  
  return nil
  
} // end error




/****************/
/* PART PÚBLICA */
/****************/

type ISO_DateTime struct {

  Year    string
  Month   string
  Day     string
  Hour    string
  Minute  string
  Second  string
  HSecond string
  GMT     int  // Intervals de 15 minuts des de -48 (oest) fins 52 (est)
  Empty   bool
  
}

type ISO_PrimaryVolume struct {

  // Part pública
  Version                uint8
  SystemIdentifier       string
  VolumeIdentifier       string
  VolumeSpaceSize        uint32 // Number of Logical Blocks in which the
                                // volume is recorded.
  VolumeSetSize          uint16 // The size of the set in this logical
                                // volume (number of disks).
  VolumeSequenceNumber   uint16 // The number of this disk in the Volume Set.
  LogicalBlockSize       uint16 // The size in bytes of a logical
                                // block. NB: This means that a
                                // logical block on a CD could be
                                // something other than 2 KiB!
  VolumeSetIdentifier    string // Identifier of the volume set of
                                // which this volume is a member.
  PublisherIdentifier    string // The volume publisher. For extended
                                // publisher information, the first
                                // byte should be 0x5F, followed by
                                // the filename of a file in the root
                                // directory. If not specified, all
                                // bytes should be 0x20.
  DataPreparerIdentifier string // The identifier of the person(s) who
                                // prepared the data for this
                                // volume. For extended preparation
                                // information, the first byte should
                                // be 0x5F, followed by the filename
                                // of a file in the root directory. If
                                // not specified, all bytes should be
                                // 0x20.
  ApplicationIdentifier  string  // Identifies how the data are
                                 // recorded on this volume. For
                                 // extended information, the first
                                 // byte should be 0x5F, followed by
                                 // the filename of a file in the root
                                 // directory. If not specified, all
                                 // bytes should be 0x20.  IGNORE coses
                                 // de la Path Table
  CopyrightFileIdentifier string // Filename of a file in the root
                                 // directory that contains copyright
                                 // information for this volume
                                 // set. If not specified, all bytes
                                 // should be 0x20.
  AbstractFileIdentifier  string // Filename of a file in the root
                                 // directory that contains abstract
                                 // information for this volume
                                 // set. If not specified, all bytes
                                 // should be 0x20.
  BiblioFileIdentifier    string // Filename of a file in the root
                                 // directory that contains
                                 // bibliographic information for this
                                 // volume set. If not specified, all
                                 // bytes should be 0x20.
  VolumeCreation          ISO_DateTime
  VolumeModification      ISO_DateTime
  VolumeExpiration        ISO_DateTime
  VolumeEffective         ISO_DateTime
  FileStructureVersion    uint8
  
  // Part privada
  root_dir_record [34]byte
  
}

type ISO struct {

  // Públic
  PrimaryVolume ISO_PrimaryVolume
  
  // Privat
  f TrackReader
  
}


func ReadISO( f TrackReader ) (*ISO,error) {

  ret:= ISO{
    f : f,
  }
  
  // Parse volume descriptors.
  if err:= ret.readVolumeDescriptors (); err != nil {
    return nil,err
  }
  
  return &ret,nil
  
} // end ReadISO
