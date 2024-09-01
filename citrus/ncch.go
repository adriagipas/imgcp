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
 *  ncch.go - Nintendo Content Container Header format.
 */

package citrus

import (
  "errors"
  "fmt"
  "io"
  "os"
  
  "github.com/adriagipas/imgcp/utils"
)


/*********/
/* TIPUS */
/*********/

const (
  NCCH_PLATFORM_3DS    = 0
  NCCH_PLATFORM_NEW3DS = 1
  NCCH_PLATFORM_UNK    = -1
)

const (
  NCCH_FLAGS_DATA          = 0x01
  NCCH_FLAGS_EXECUTABLE    = 0x02
  NCCH_FLAGS_SYSTEM_UPDATE = 0x04
  NCCH_FLAGS_MANUAL        = 0x08
  NCCH_FLAGS_TRIAL         = 0x10
)

const (
  NCCH_TYPE_CXI = 0
  NCCH_TYPE_CFA = 1
  NCCH_TYPE_UNK = -1
)

type NCCH_FileOffset struct {

  Offset     int64
  Size       int64
  HeaderSize int64
  
}

type NCCH_Header struct {

  Size        int64
  Id          uint64
  MakerCode   string
  Version     uint16
  ProgramId   uint64
  ProductCode string
  Platform    int
  Flags       uint8
  Type        int
  Plain       NCCH_FileOffset
  Logo        NCCH_FileOffset
  ExeFS       NCCH_FileOffset
  RomFS       NCCH_FileOffset
  
}

type NCCH struct {

  Header    NCCH_Header

  file_name string
  offset    int64
  
}


/************/
/* FUNCIONS */
/************/

// S'ha de passar el tros de memòria on està la informació. Si es
// passen 12 bytes llig també la grandària de la capçalera. Fica -1 si
// no té grandària de capçalera.
func newNCCH_FileOffset( mem []byte ) NCCH_FileOffset {
  
  ret:= NCCH_FileOffset{
    Offset: 0x200 * int64(uint32(mem[0]) |
      (uint32(mem[1])<<8) |
      (uint32(mem[2])<<16) |
      (uint32(mem[3])<<24)),
    Size: 0x200 * int64(uint32(mem[4]) |
      (uint32(mem[5])<<8) |
      (uint32(mem[6])<<16) |
      (uint32(mem[7])<<24)),
  }
  if len(mem)==12 {
    ret.HeaderSize= 0x200 * int64(uint32(mem[8]) |
      (uint32(mem[9])<<8) |
      (uint32(mem[10])<<16) |
      (uint32(mem[11])<<24))
  } else {
    ret.HeaderSize= -1
  }

  return ret
  
} // end NCCH_Header.getFileOffset


func (self *NCCH_Header) Read( fd io.Reader, file_size int64 ) error {

  // Llig capçalera.
  var buf [0x200]byte
  n,err:= fd.Read ( buf[:] )
  if err != nil {
    return fmt.Errorf ( "Error while reading NCCH header: %s", err )
  }
  if n != len(buf) {
    return errors.New ( "Error while reading NCCH header: not enough bytes" )
  }

  // Comprovacions
  if buf[0x100]!='N' || buf[0x101]!='C' || buf[0x102]!='C' || buf[0x103]!='H' {
    return fmt.Errorf ( "Not a NCCH file: wrong magic number (%c%c%c%c)",
    buf[0x100], buf[0x101], buf[0x102], buf[0x103] )
  }
  header_size:= uint32(buf[0x104]) |
    (uint32(buf[0x105])<<8) |
    (uint32(buf[0x106])<<16) |
    (uint32(buf[0x107])<<24)
  self.Size= int64(header_size)*0x200
  if self.Size != file_size {
    return fmt.Errorf ( "Mismatch between file size (%d) and the size "+
      "specified in the header (%d)", file_size, self.Size )
  }

  // Llig valors
  self.Id= uint64(buf[0x108]) |
    (uint64(buf[0x109])<<8) |
    (uint64(buf[0x10a])<<16) |
    (uint64(buf[0x10b])<<24) |
    (uint64(buf[0x10c])<<32) |
    (uint64(buf[0x10d])<<40) |
    (uint64(buf[0x10e])<<48) |
    (uint64(buf[0x10f])<<56)
  self.MakerCode= string(buf[0x110:0x112])
  self.Version= uint16(buf[0x112]) | (uint16(buf[0x113])<<8)
  self.ProgramId= uint64(buf[0x118]) |
    (uint64(buf[0x119])<<8) |
    (uint64(buf[0x11a])<<16) |
    (uint64(buf[0x11b])<<24) |
    (uint64(buf[0x11c])<<32) |
    (uint64(buf[0x11d])<<40) |
    (uint64(buf[0x11e])<<48) |
    (uint64(buf[0x11f])<<56)
  self.ProductCode= string(buf[0x150:0x160])
  switch buf[0x188+4] {
  case 0x01:
    self.Platform= NCCH_PLATFORM_3DS
  case 0x02:
    self.Platform= NCCH_PLATFORM_NEW3DS
  default:
    self.Platform= NCCH_PLATFORM_UNK
  }
  self.Flags= buf[0x188+5]

  // Fixa tipus
  if (self.Flags&NCCH_FLAGS_EXECUTABLE) != 0 {
    self.Type= NCCH_TYPE_CXI
  } else if (self.Flags&NCCH_FLAGS_DATA) != 0 {
    self.Type= NCCH_TYPE_CFA
  } else {
    self.Type= NCCH_TYPE_UNK
  }

  // Llig els offsets dels fitxers.
  self.Plain= newNCCH_FileOffset ( buf[0x190:0x198] )
  self.Logo= newNCCH_FileOffset ( buf[0x198:0x1a0] )
  self.ExeFS= newNCCH_FileOffset ( buf[0x1a0:0x1ac] )
  self.RomFS= newNCCH_FileOffset ( buf[0x1b0:0x1bc] )
  
  return nil
  
} // NCCH_Header.Read


func newNCCHSubfile(
  file_name string,
  offset int64,
  length int64,
) (*NCCH,error) {

  // Inicialitza
  ret:= NCCH{
    file_name: file_name,
    offset: offset,
  }

  // Llig capçalera
  fd,err:= utils.NewSubfileReader ( file_name, offset, length )
  if err != nil { return nil,err }
  defer fd.Close ()
  if err:= ret.Header.Read ( fd, length ); err != nil {
    return nil,err
  }
  
  return &ret,nil
  
} // end newNCCHSubfile


func NewNCCH( file_name string ) (*NCCH,error) {
  
  // Obté grandària.
  fd,err:= os.Open ( file_name )
  if err != nil { return nil,err }
  defer fd.Close ()
  info,err:= fd.Stat ()
  if err != nil { return nil,err }
  
  // Crea
  return newNCCHSubfile ( file_name, 0, info.Size () )
  
} // end NewNCCH


// Si no en té torna nil sense error.
func (self *NCCH) GetExeFS() (*ExeFS,error) {
  if self.Header.ExeFS.Size == 0 {
    return nil,nil
  } else {
    return newExeFS (
      self.file_name,
      self.offset + self.Header.ExeFS.Offset,
      self.Header.ExeFS.Size,
    )
  }
} // end GetExeFS


// Si no en té torna nil sense error.
func (self *NCCH) GetRomFS() (*RomFS_Directory,error) {
  if self.Header.RomFS.Size == 0 {
    return nil,nil
  } else {
    return openRomFS (
      self.file_name,
      self.offset + self.Header.RomFS.Offset,
      self.Header.RomFS.Size,
    )
  }
} // end GetRomFS
