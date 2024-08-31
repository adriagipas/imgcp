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
 *  ncsd.go - NCSD format.
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

const (
  NCSD_PARTITION_TYPE_NCCH     = 0
  NCSD_PARTITION_TYPE_MBR      = 1
  NCSD_PARTITION_TYPE_FIRM     = 2
  NCSD_PARTITION_TYPE_GBA_SAVE = 3
  NCSD_PARTITION_TYPE_UNUSED   = -1
)

type NCSD_Partition struct {
  
  Type   int
  Offset int64
  Size   int64
  
}

type NCSDHeader struct {

  Size       int64
  Partitions [8]NCSD_Partition
  MediaID    uint64
  
}


/************/
/* FUNCIONS */
/************/

func (self *NCSDHeader) Read( fd *os.File ) error {

  // Rebobina i obté grandària
  if _,err:= fd.Seek ( 0, 0 ); err != nil {
    return err
  }
  info,err:= fd.Stat ()
  if err != nil {
    return err
  }

  // Llig capçalera
  var buf [0x160]byte
  n,err:= fd.Read ( buf[:] )
  if err != nil {
    return fmt.Errorf ( "Error while reading NCSD header: %s", err )
  }
  if n != len(buf) {
    return errors.New ( "Error while reading NCSD header: not enough bytes" )
  }

  // Comprovacions
  if buf[0x100]!='N' || buf[0x101]!='C' || buf[0x102]!='S' || buf[0x103]!='D' {
    return fmt.Errorf ( "Not a NCSD image: wrong magic number (%c%c%c%c)",
    buf[0x100], buf[0x101], buf[0x102], buf[0x103] )
  }
  file_size:= info.Size ()
  header_size:= uint32(buf[0x104]) |
    (uint32(buf[0x105])<<8) |
    (uint32(buf[0x106])<<16) |
    (uint32(buf[0x107])<<24)
  self.Size= int64(header_size)*0x200
  if self.Size != file_size {
    return fmt.Errorf ( "Mismatch between image size (%d) and the size "+
      "specified in the header (%d)", file_size, self.Size )
  }
  
  // Llig informació particions
  for i:= 0; i < 8; i++ {

    // Tipus de partició
    switch pt:= buf[0x110+i]; pt {
    case 0:
      self.Partitions[i].Type= NCSD_PARTITION_TYPE_NCCH
    case 1:
      self.Partitions[i].Type= NCSD_PARTITION_TYPE_MBR
    case 2:
      self.Partitions[i].Type= NCSD_PARTITION_TYPE_FIRM
    case 3:
      self.Partitions[i].Type= NCSD_PARTITION_TYPE_GBA_SAVE
    default:
      return fmt.Errorf ( "Error while reading NCSD header: unknown partition"+
        " type (%d) for partition %d", pt, i )
    }

    // Offset i grandària
    tmp:= buf[0x120+i*8:0x120+(i+1)*8]
    self.Partitions[i].Offset= 0x200 * int64(uint32(tmp[0]) |
      (uint32(tmp[1])<<8) |
      (uint32(tmp[2])<<16) |
      (uint32(tmp[3])<<24))
    self.Partitions[i].Size= 0x200 * int64(uint32(tmp[4]) |
      (uint32(tmp[5])<<8) |
      (uint32(tmp[6])<<16) |
      (uint32(tmp[7])<<24))

    // Si està buit modifica tipus
    if self.Partitions[i].Size == 0 {
      self.Partitions[i].Type= NCSD_PARTITION_TYPE_UNUSED
    }

    // Comprovacions bàsiques.
    if self.Partitions[i].Offset+self.Partitions[i].Size > file_size {
      return fmt.Errorf ( "Error while reading NCSD header: partition %d"+
        " ([%d,%d+d[) is out of image boundaries ([%d,%d[) than file image",
        self.Partitions[i].Offset,
        self.Partitions[i].Offset+self.Partitions[i].Size,
        0, file_size )
    }
    
  }

  // Media ID
  self.MediaID= uint64(buf[0x108]) |
    (uint64(buf[0x109])<<8) |
    (uint64(buf[0x10a])<<16) |
    (uint64(buf[0x10b])<<24) |
    (uint64(buf[0x10c])<<32) |
    (uint64(buf[0x10d])<<40) |
    (uint64(buf[0x10e])<<48) |
    (uint64(buf[0x10f])<<56)

  return nil
  
} // end Read
