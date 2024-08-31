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
 * along with adriagipas/imgcp.  If not, see
 * <https://www.gnu.org/licenses/>.
 */
/*
 *  cci.go - CTR Cart Image format.
 */

package citrus

import (
  "errors"
  "fmt"
  "os"
)


/*********/
/* TIPUS */
/*********/

type CCIHeader struct {

  NCSDHeader
  PartitionIDs [8]uint64
  TitleVersion uint16
  CardRevision uint16
  TitleID      uint64 // Algo de CVer ??
  VersionCVer  uint16
  
}

type CCI struct {

  Header    CCIHeader
  file_name string
  
}


/************/
/* FUNCIONS */
/************/

func (self *CCIHeader) Read( fd *os.File ) error {

  // Capçalera NCSD
  if err:= self.NCSDHeader.Read ( fd ); err != nil {
    return err
  }
  if self.Partitions[0].Type != NCSD_PARTITION_TYPE_NCCH ||
    self.Partitions[1].Type != NCSD_PARTITION_TYPE_NCCH ||
    (self.Partitions[2].Type != NCSD_PARTITION_TYPE_NCCH &&
      self.Partitions[2].Type != NCSD_PARTITION_TYPE_UNUSED) ||
    self.Partitions[3].Type != NCSD_PARTITION_TYPE_UNUSED ||
    self.Partitions[4].Type != NCSD_PARTITION_TYPE_UNUSED ||
    self.Partitions[5].Type != NCSD_PARTITION_TYPE_UNUSED ||
    (self.Partitions[6].Type != NCSD_PARTITION_TYPE_NCCH &&
      self.Partitions[6].Type != NCSD_PARTITION_TYPE_UNUSED) ||
    self.Partitions[7].Type != NCSD_PARTITION_TYPE_NCCH {
    return errors.New (
      "Error while reading CCI header: invalid partition types" )
  }

  // Llig resta capçalera CCI
  if _,err:= fd.Seek ( 0x160, 0 ); err != nil {
    return fmt.Errorf ( "Error while reading CCI header: %s", err )
  }
  var buf [0x10a0]byte
  n,err:= fd.Read ( buf[:] )
  if err != nil {
    return fmt.Errorf ( "Error while reading CCI header: %s", err )
  }
  if n != len(buf) {
    return errors.New ( "Error while reading CCI header: not enough bytes" )
  }

  // Partition ID Table
  for i:= 0; i < 8; i++ {
    self.PartitionIDs[i]= uint64(buf[0x30+i*8]) |
      (uint64(buf[0x31+i*8])<<8) |
      (uint64(buf[0x32+i*8])<<16) |
      (uint64(buf[0x33+i*8])<<24) |
      (uint64(buf[0x34+i*8])<<32) |
      (uint64(buf[0x35+i*8])<<40) |
      (uint64(buf[0x36+i*8])<<48) |
      (uint64(buf[0x37+i*8])<<56)
  }

  // Altres
  self.TitleVersion= uint16(buf[0x1b0]) | (uint16(buf[0x1b1])<<8)
  self.CardRevision= uint16(buf[0x1b2]) | (uint16(buf[0x1b3])<<8)
  self.TitleID= uint64(buf[0x1c0]) |
      (uint64(buf[0x1c1])<<8) |
      (uint64(buf[0x1c2])<<16) |
      (uint64(buf[0x1c3])<<24) |
      (uint64(buf[0x1c4])<<32) |
      (uint64(buf[0x1c5])<<40) |
      (uint64(buf[0x1c6])<<48) |
      (uint64(buf[0x1c7])<<56)
  self.VersionCVer= uint16(buf[0x1c8]) | (uint16(buf[0x1c9])<<8)
  
  return nil
  
} // end CCIHeader.Read


func NewCCI( file_name string ) (*CCI,error) {

  // Inicialitza.
  ret:= CCI{
    file_name: file_name,
  }
  
  // Llig capçalera.
  fd,err:= os.Open ( file_name )
  if err != nil {
    return nil,err
  }
  defer fd.Close ()
  if err:= ret.Header.Read ( fd ); err != nil {
    return nil,err
  }
  
  return &ret,nil
  
} // end NewCCI


func (self *CCI) GetNCCHPartition( ind int ) (*NCCH,error) {

  if self.Header.Partitions[ind].Type != NCSD_PARTITION_TYPE_NCCH {
    return nil,fmt.Errorf ( "Partition %d is not a NCCH partition", ind )
  }
  return newNCCHSubfile (
    self.file_name,
    self.Header.Partitions[ind].Offset,
    self.Header.Partitions[ind].Size,
  )
  
} // CCI.GetNCCHPartition
