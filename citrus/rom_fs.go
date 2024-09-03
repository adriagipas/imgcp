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
 *  rom_fs.go - NCCH ROM Filesystem.
 */

package citrus

import (
  "errors"
  "fmt"
  
  "github.com/adriagipas/imgcp/utils"
  "golang.org/x/text/encoding/unicode"
)


/*********/
/* TIPUS */
/*********/

const _BASE_OFFSET= 0x1000

type RomFS_Directory struct {

  Name string // La cadena buida representa el root
  
  // Fitxer
  file_name   string
  file_offset int64
  file_size   int64

  // Taules
  directory_table uint32
  file_table      uint32
  file_data       uint32
  
  // Referències
  self    uint32
  parent  uint32
  sibling uint32
  child   uint32
  file    uint32
  
}

type RomFS_File struct {

  Name string // La cadena buida representa el root
  
  // Fitxer
  file_name   string
  file_offset int64
  file_size   int64

  // Taules
  directory_table uint32
  file_table      uint32
  file_data       uint32
  
  // Referències
  parent_dir uint32
  sibling    uint32

  // Fitxer
  offset uint64
  Size   uint64
  
}


/************/
/* FUNCIONS */
/************/

func (self *RomFS_File) Open() (*utils.SubfileReader,error) {

  offset:= _BASE_OFFSET + int64(uint64(self.file_data) + self.offset)
  size:= int64(self.Size)
  if offset < 0 || offset+size >= self.file_size {
    return nil,fmt.Errorf ( "File '%s' out of range (parent file)", self.Name )
  }
  return utils.NewSubfileReader (
    self.file_name,
    self.file_offset + offset,
    size,
  )
  
} // end RomFS_File.Open


// Si no en té torna nil sense error
func (self *RomFS_File) Sibling() (*RomFS_File,error) {

  if self.sibling == 0xFFFFFFFF { return nil,nil }
  return newRomFS_File (
    self.file_name, self.file_offset, self.file_size,
    self.directory_table, self.file_table, self.file_data,
    self.sibling )
  
} // end RomFS_File.Sibling


func newRomFS_File(
  
  file_name       string,
  file_offset     int64,
  file_length     int64,
  directory_table uint32,
  file_table      uint32,
  file_data       uint32,
  entry_offset    uint32,
  
) (*RomFS_File,error) {

  // Llig entry
  fd,err:= utils.NewSubfileReader ( file_name, file_offset, file_length )
  if err != nil { return nil,err }
  defer fd.Close ()
  real_file_offset:= _BASE_OFFSET + int64(uint64(file_table + entry_offset))
  if _,err:= fd.Seek ( real_file_offset, 0 ); err != nil {
    return nil,err
  }
  var buf [0x20]byte
  n,err:= fd.Read ( buf[:] )
  if err != nil {
    return nil,fmt.Errorf (
      "Error while reading File entry for offset %08X: %s",
      entry_offset, err )
  }
  if n != len(buf) {
    return nil,fmt.Errorf (
      "Error while reading File entry for offset %08X: not enough bytes",
      entry_offset )
  }
  
  // Llig offsets
  parent_dir:= uint32(buf[0]) |
    (uint32(buf[1])<<8) |
    (uint32(buf[2])<<16) |
    (uint32(buf[3])<<24)
  sibling:= uint32(buf[4]) |
    (uint32(buf[5])<<8) |
    (uint32(buf[6])<<16) |
    (uint32(buf[7])<<24)

  // Llig característiques fitxer
  if (buf[15]&0x80)!=0 || (buf[23]&0x80)!=0 {
    return nil,fmt.Errorf (
      "Error while reading File entry for offset %08X: file too large",
      entry_offset )
  }
  offset:= uint64(buf[8]) |
    (uint64(buf[9])<<8) |
    (uint64(buf[10])<<16) |
    (uint64(buf[11])<<24) |
    (uint64(buf[12])<<32) |
    (uint64(buf[13])<<40) |
    (uint64(buf[14])<<48) |
    (uint64(buf[15])<<56)
  size:= uint64(buf[16]) |
    (uint64(buf[17])<<8) |
    (uint64(buf[18])<<16) |
    (uint64(buf[19])<<24) |
    (uint64(buf[20])<<32) |
    (uint64(buf[21])<<40) |
    (uint64(buf[22])<<48) |
    (uint64(buf[23])<<56)
  
  // Llig grandària nom
  name_length:= uint32(buf[28]) |
    (uint32(buf[29])<<8) |
    (uint32(buf[30])<<16) |
    (uint32(buf[31])<<24)
  name:= ""
  if name_length > 0 {
    tmp:= make([]byte,name_length)
    n,err= fd.Read ( tmp[:] )
    if err != nil {
      return nil,fmt.Errorf (
        "Error while reading File name for offset %08X: %s",
        entry_offset, err )
    }
    if n != len(tmp) {
      return nil,fmt.Errorf (
        "Error while reading File name for offset %08X: not enough bytes",
        entry_offset )
    }
    dec:= unicode.UTF16(unicode.LittleEndian,unicode.IgnoreBOM).NewDecoder ()
    aux,err:= dec.Bytes ( tmp )
    if err != nil {
      return nil,fmt.Errorf (
        "Error while reading File name for offset %08X: %s",
        entry_offset, err )
    }
    name= string(aux)
  }
  
  // Inicialitza.
  ret:= RomFS_File{
    Name: name,
    file_name: file_name,
    file_offset: file_offset,
    file_size: file_length,
    directory_table: directory_table,
    file_table: file_table,
    file_data: file_data,
    parent_dir: parent_dir,
    sibling: sibling,
    offset: offset,
    Size: size,
  }

  return &ret,nil
  
} // end newRomFS_File


// Si no en té torna nil sense error
func (self *RomFS_Directory) Parent() (*RomFS_Directory,error) {

  if self.parent == self.self { return self,nil } // Cas especial
  if self.parent == 0xFFFFFFFF { return nil,nil }
  return newRomFS_Directory (
    self.file_name, self.file_offset, self.file_size,
    self.directory_table, self.file_table, self.file_data,
    self.parent )
  
} // end RomFS_Directory.Parent


// Si no en té torna nil sense error
func (self *RomFS_Directory) Sibling() (*RomFS_Directory,error) {

  if self.sibling == self.self { return self,nil } // Cas especial
  if self.sibling == 0xFFFFFFFF { return nil,nil }
  return newRomFS_Directory (
    self.file_name, self.file_offset, self.file_size,
    self.directory_table, self.file_table, self.file_data,
    self.sibling )
  
} // end RomFS_Directory.Sibling


// Si no en té torna nil sense error
func (self *RomFS_Directory) Child() (*RomFS_Directory,error) {

  if self.child == self.self { return self,nil } // Cas especial
  if self.child == 0xFFFFFFFF { return nil,nil }
  return newRomFS_Directory (
    self.file_name, self.file_offset, self.file_size,
    self.directory_table, self.file_table, self.file_data,
    self.child )
  
} // end RomFS_Directory.Child


// Si no en té torna nil sense error
func (self *RomFS_Directory) File() (*RomFS_File,error) {

  if self.file == 0xFFFFFFFF { return nil,nil }
  return newRomFS_File (
    self.file_name, self.file_offset, self.file_size,
    self.directory_table, self.file_table, self.file_data,
    self.file )
  
} // end RomFS_Directory.File


func newRomFS_Directory(
  
  file_name       string,
  offset          int64,
  length          int64,
  directory_table uint32,
  file_table      uint32,
  file_data       uint32,
  entry_offset    uint32,
  
) (*RomFS_Directory,error) {

  // Llig entry
  fd,err:= utils.NewSubfileReader ( file_name, offset, length )
  if err != nil { return nil,err }
  defer fd.Close ()
  dir_offset:= _BASE_OFFSET + int64(uint64(directory_table + entry_offset))
  if _,err:= fd.Seek ( dir_offset, 0 ); err != nil {
    return nil,err
  }
  var buf [0x18]byte
  n,err:= fd.Read ( buf[:] )
  if err != nil {
    return nil,fmt.Errorf (
      "Error while reading Directory entry for offset %08X: %s",
      entry_offset, err )
  }
  if n != len(buf) {
    return nil,fmt.Errorf (
      "Error while reading Directory entry for offset %08X: not enough bytes",
      entry_offset )
  }
  
  // Llig offsets
  parent:= uint32(buf[0]) |
    (uint32(buf[1])<<8) |
    (uint32(buf[2])<<16) |
    (uint32(buf[3])<<24)
  sibling:= uint32(buf[4]) |
    (uint32(buf[5])<<8) |
    (uint32(buf[6])<<16) |
    (uint32(buf[7])<<24)
  child:= uint32(buf[8]) |
    (uint32(buf[9])<<8) |
    (uint32(buf[10])<<16) |
    (uint32(buf[11])<<24)
  file:= uint32(buf[12]) |
    (uint32(buf[13])<<8) |
    (uint32(buf[14])<<16) |
    (uint32(buf[15])<<24)

  // Llig grandària nom
  name_length:= uint32(buf[20]) |
    (uint32(buf[21])<<8) |
    (uint32(buf[22])<<16) |
    (uint32(buf[23])<<24)
  name:= ""
  if name_length > 0 {
    tmp:= make([]byte,name_length)
    n,err= fd.Read ( tmp[:] )
    if err != nil {
      return nil,fmt.Errorf (
        "Error while reading Directory name for offset %08X: %s",
        entry_offset, err )
    }
    if n != len(tmp) {
      return nil,fmt.Errorf (
        "Error while reading Directory name for offset %08X: not enough bytes",
        entry_offset )
    }
    dec:= unicode.UTF16(unicode.LittleEndian,unicode.IgnoreBOM).NewDecoder ()
    aux,err:= dec.Bytes ( tmp )
    if err != nil {
      return nil,fmt.Errorf (
        "Error while reading Directory name for offset %08X: %s",
        entry_offset, err )
    }
    name= string(aux)
  }
  
  // Inicialitza.
  ret:= RomFS_Directory{
    Name: name,
    file_name: file_name,
    file_offset: offset,
    file_size: length,
    directory_table: directory_table,
    file_table: file_table,
    file_data: file_data,
    self: entry_offset,
    parent: parent,
    sibling: sibling,
    child: child,
    file: file,
  }

  return &ret,nil
  
} // end newRomFS_Directory


func openRomFS(
  file_name string,
  offset    int64,
  length    int64,
) (*RomFS_Directory,error) {

  // Obri subfitxer.
  fd,err:= utils.NewSubfileReader ( file_name, offset, length )
  if err != nil { return nil,err }
  defer fd.Close ()

  // Llig capçalera.
  var buf [0x5c]byte
  n,err:= fd.Read ( buf[:] )
  if err != nil {
    return nil,fmt.Errorf ( "Error while reading RomFS header: %s", err )
  }
  if n != len(buf) {
    return nil,
      errors.New ( "Error while reading RomFS header: not enough bytes" )
  }
  if buf[0]!='I' || buf[1]!='V' || buf[2]!='F' || buf[3]!='C' ||
    buf[4]!=0x00 || buf[5]!=0x00 || buf[6]!=0x01 || buf[7]!=0x00 {
    return nil,fmt.Errorf (
      "Not a RomFs file: wrong magic number (%c%c%c%c%d%d%d%d)",
    buf[0], buf[1], buf[2], buf[3], buf[4], buf[5], buf[6], buf[7] )
  }
  
  // Obté offset taula directories
  if _,err:= fd.Seek ( _BASE_OFFSET+0xc, 0 ); err != nil {
    return nil,fmt.Errorf (
      "Error while trying to locate the directory table: %s", err )
  }
  var tmp [4]byte
  n,err= fd.Read ( tmp[:] )
  if err != nil {
    return nil,fmt.Errorf (
      "Error while trying to locate the directory table: %s", err )
  }
  if n != len(tmp) {
    return nil,errors.New (
      "Error while trying to locate the directory table: not enough bytes" )
  }
  directory_table:= uint32(tmp[0]) |
    (uint32(tmp[1])<<8) |
    (uint32(tmp[2])<<16) |
    (uint32(tmp[3])<<24)

  // Obté offset taula fitxers
  if _,err:= fd.Seek ( _BASE_OFFSET+0x1c, 0 ); err != nil {
    return nil,fmt.Errorf (
      "Error while trying to locate the file table: %s", err )
  }
  n,err= fd.Read ( tmp[:] )
  if err != nil {
    return nil,fmt.Errorf (
      "Error while trying to locate the file table: %s", err )
  }
  if n != len(tmp) {
    return nil,errors.New (
      "Error while trying to locate the file table: not enough bytes" )
  }
  file_table:= uint32(tmp[0]) |
    (uint32(tmp[1])<<8) |
    (uint32(tmp[2])<<16) |
    (uint32(tmp[3])<<24)

  // Obté offset dades fitxers
  if _,err:= fd.Seek ( _BASE_OFFSET+0x24, 0 ); err != nil {
    return nil,fmt.Errorf (
      "Error while trying to locate the file data offset: %s", err )
  }
  n,err= fd.Read ( tmp[:] )
  if err != nil {
    return nil,fmt.Errorf (
      "Error while trying to locate the file data offset: %s", err )
  }
  if n != len(tmp) {
    return nil,errors.New (
      "Error while trying to locate the file data offset: not enough bytes" )
  }
  file_data:= uint32(tmp[0]) |
    (uint32(tmp[1])<<8) |
    (uint32(tmp[2])<<16) |
    (uint32(tmp[3])<<24)
  
  // Crea directory
  return newRomFS_Directory (
    file_name, offset, length, directory_table, file_table, file_data, 0 )
  
} // end openRomFS
