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
 * along with adriagipas/imgcp.  If not, see
 * <https://www.gnu.org/licenses/>.
 */
/*
 *  stfs.go - Secure Transacted File System
 */

package x360

import (
  "bytes"
  "errors"
  "fmt"
  "io"
  "os"

  "golang.org/x/text/encoding"
  "golang.org/x/text/encoding/unicode"
)


/*********/
/* TIPUS */
/*********/

const (
  STFS_TYPE_CONS = 0
  STFS_TYPE_PIRS = 1
  STFS_TYPE_LIVE = 2

  CONTENT_TYPE_SAVED_GAME = 0
  CONTENT_TYPE_MARKETPLACE_CONTENT = 1
  CONTENT_TYPE_PUBLISHER = 2
  CONTENT_TYPE_XBOX360_TITLE = 3
  CONTENT_TYPE_IPTV_PAUSE_BUFFER = 4
  CONTENT_TYPE_INSTALLED_GAME = 5
  CONTENT_TYPE_XBOX_TITLE = 6
  CONTENT_TYPE_GAME_ON_DEMAND = 7
  CONTENT_TYPE_AVATAR_ITEM = 8
  CONTENT_TYPE_PROFILE = 9
  CONTENT_TYPE_GAMER_PICTURE = 10
  CONTENT_TYPE_THEME = 11
  CONTENT_TYPE_CACHE_FILE = 12
  CONTENT_TYPE_STORAGE_DOWNLOAD = 13
  CONTENT_TYPE_XBOX_SAVED_GAME = 14
  CONTENT_TYPE_XBOX_DOWNLOAD = 15
  CONTENT_TYPE_GAME_DEMO = 16
  CONTENT_TYPE_VIDEO = 17
  CONTENT_TYPE_GAME_TITLE = 18
  CONTENT_TYPE_INSTALLER = 19
  CONTENT_TYPE_GAME_TRAILER = 20
  CONTENT_TYPE_ARCADE_TITLE = 21
  CONTENT_TYPE_XNA = 22
  CONTENT_TYPE_LICENSE_STORE = 23
  CONTENT_TYPE_MOVIE = 24
  CONTENT_TYPE_TV = 25
  CONTENT_TYPE_MUSIC_VIDEO = 26
  CONTENT_TYPE_GAME_VIDEO = 27
  CONTENT_TYPE_PODCAST_VIDEO = 28
  CONTENT_TYPE_VIRAL_VIDEO = 29
  CONTENT_TYPE_COMMUNITY_GAME = 30
  CONTENT_TYPE_UNK = -1

  PLATFORM_XBOX360 = 0
  PLATFORM_PC = 1
  PLATFORM_UNK = -1
  
)

const _STFS_HEADER_SIZE = 0x22C
const _STFS_METADATA_SIZE = 0x94EE

type STFSHeader struct {
  
  Type int

  // LIVE/PIRS
  PackageSignature [0x100]byte

  // CONS
  CertOwnConsoleID         [5]byte
  CertOwnConsolePartNumber string
  CertOwnConsoleType       uint8
  CertDateGeneration       string
  PublicExponent           [4]byte
  PublicModulus            [0x80]byte
  CertSignature            [0x100]byte
  Signature                [0x80]byte
  
}

type STFSVolumeDescriptor struct {

  BlockSeparation            uint8
  FileTableBlockCount        int16
  FileTableBlockNumber       int32
  TopHashTableHash           [0x14]byte
  TotalAllocatedBlockCount   int32
  TotalUnallocatedBlockCount int32
  
}

type STFSMetadata struct {
  
  ContentID          [0x14]byte
  EntryID            uint32
  ContentType        int
  MetadataVersion    uint32
  ContentSize        int64
  MediaID            uint32
  Version            int32
  BaseVersion        int32
  TitleID            uint32
  Platform           int
  ExecutableType     uint8
  DiscNumber         uint8
  DiscInSet          uint8
  SaveGameID         uint32
  ConsoleID          [5]byte
  ProfileID          [8]byte
  Volume             STFSVolumeDescriptor
  DataFileCount      int32
  DataFileCombSize   int64
  DeviceID           [0x14]byte
  DisplayName        [12]string
  DisplayDescription [12]string
  PublisherName      string
  TitleName          string
  TransferFlags      uint8
  Thumbnail          []byte
  TitleThumbnail     []byte

  // Metadata Version 2
  SeriesID              [0x10]byte
  SeasonID              [0x10]byte
  SeasonNumber          int16
  EpisodeNumber         int16
  
}

type STFS struct {

  Header    STFSHeader
  Metadata  STFSMetadata
  file_name string
  
}


/************/
/* FUNCIONS */
/************/

func _u16( v []byte ) uint16 {
  return (uint16(v[0])<<8) | uint16(v[1])
} // end _u16


func _u32( v []byte ) uint32 {
  return (uint32(v[0])<<24) |
    (uint32(v[1])<<16) |
    (uint32(v[2])<<8) |
    uint32(v[3])
} // end _u32


func _u64( v []byte ) uint64 {
  return (uint64(v[0])<<56) |
    (uint64(v[1])<<48) |
    (uint64(v[2])<<40) |
    (uint64(v[3])<<32) |
    (uint64(v[4])<<24) |
    (uint64(v[5])<<16) |
    (uint64(v[6])<<8) |
    uint64(v[7])
} // end _u64


func ctype2int( ctype uint32 ) int {

  var ret int
  switch ctype {
  case 0x0000001:
    ret= CONTENT_TYPE_SAVED_GAME
  case 0x0000002:
    ret= CONTENT_TYPE_MARKETPLACE_CONTENT
  case 0x0000003:
    ret= CONTENT_TYPE_PUBLISHER
  case 0x0001000:
    ret= CONTENT_TYPE_XBOX360_TITLE
  case 0x0002000:
    ret= CONTENT_TYPE_IPTV_PAUSE_BUFFER
  case 0x0004000:
    ret= CONTENT_TYPE_INSTALLED_GAME
  case 0x0005000:
    ret= CONTENT_TYPE_XBOX_TITLE
  case 0x0007000:
    ret= CONTENT_TYPE_GAME_ON_DEMAND
  case 0x0009000:
    ret= CONTENT_TYPE_AVATAR_ITEM
  case 0x0010000:
    ret= CONTENT_TYPE_PROFILE
  case 0x0020000:
    ret= CONTENT_TYPE_GAMER_PICTURE
  case 0x0030000:
    ret= CONTENT_TYPE_THEME
  case 0x0040000:
    ret= CONTENT_TYPE_CACHE_FILE
  case 0x0050000:
    ret= CONTENT_TYPE_STORAGE_DOWNLOAD
  case 0x0060000:
    ret= CONTENT_TYPE_XBOX_SAVED_GAME
  case 0x0070000:
    ret= CONTENT_TYPE_XBOX_DOWNLOAD
  case 0x0080000:
    ret= CONTENT_TYPE_GAME_DEMO
  case 0x0090000:
    ret= CONTENT_TYPE_VIDEO
  case 0x00A0000:
    ret= CONTENT_TYPE_GAME_TITLE
  case 0x00B0000:
    ret= CONTENT_TYPE_INSTALLER
  case 0x00C0000:
    ret= CONTENT_TYPE_GAME_TRAILER
  case 0x00D0000:
    ret= CONTENT_TYPE_ARCADE_TITLE
  case 0x00E0000:
    ret= CONTENT_TYPE_XNA
  case 0x00F0000:
    ret= CONTENT_TYPE_LICENSE_STORE
  case 0x0100000:
    ret= CONTENT_TYPE_MOVIE
  case 0x0200000:
    ret= CONTENT_TYPE_TV
  case 0x0300000:
    ret= CONTENT_TYPE_MUSIC_VIDEO
  case 0x0400000:
    ret= CONTENT_TYPE_GAME_VIDEO
  case 0x0500000:
    ret= CONTENT_TYPE_PODCAST_VIDEO
  case 0x0600000:
    ret= CONTENT_TYPE_VIRAL_VIDEO
  case 0x2000000:
    ret= CONTENT_TYPE_COMMUNITY_GAME
  default:
    ret= CONTENT_TYPE_UNK
  }

  return ret
    
} // ctype2int


func _str( dec *encoding.Decoder, data []byte ) (ret string) {

  if aux,err:= dec.Bytes ( data ); err == nil {
    aux= bytes.TrimRight ( aux, "\000" )
    ret= string(aux)
  } else {
    data= bytes.TrimRight ( data, "\000" )
    ret= string(data)
  }
  
  return
  
} // end _str


func (self *STFSVolumeDescriptor) Read( v []byte ) {

  self.BlockSeparation= uint8(v[1])
  self.FileTableBlockCount= int16(_u16(v[2:]))
  tmp:= (uint32(v[4])<<16) | (uint32(v[5])<<8) | uint32(v[6])
  if tmp&0x800000 != 0 {
    tmp|= 0xFF000000
  }
  self.FileTableBlockNumber= int32(tmp)
  copy ( self.TopHashTableHash[:], v[7:27] )
  self.TotalAllocatedBlockCount= int32(_u32(v[27:]))
  self.TotalUnallocatedBlockCount= int32(_u32(v[31:]))
  
} // STFSVolumeDescriptor.Read


func (self *STFSMetadata) Read( fd io.Reader ) error {
  
  // Llig capçalera.
  var buf [_STFS_METADATA_SIZE]byte
  n,err:= fd.Read ( buf[:] )
  if err != nil {
    return fmt.Errorf ( "Error while reading STFS metadata: %s", err )
  }
  if n != len(buf) {
    return errors.New ( "Error while reading STFS metadata: not enough bytes" )
  }

  // Llig camps
  copy ( self.ContentID[:], buf[0x100:0x100+0x14] )
  self.EntryID= _u32(buf[0x114:])
  self.ContentType= ctype2int ( _u32(buf[0x118:]) )
  self.MetadataVersion= _u32(buf[0x11c:])
  self.ContentSize= int64(_u64(buf[0x120:]))
  self.MediaID= _u32(buf[0x128:])
  self.Version= int32(_u32(buf[0x12c:]))
  self.BaseVersion= int32(_u32(buf[0x130:]))
  self.TitleID= _u32(buf[0x134:])
  switch buf[0x138] {
  case 2:
    self.Platform= PLATFORM_XBOX360
  case 4:
    self.Platform= PLATFORM_PC
  default:
    self.Platform= PLATFORM_UNK
  }
  self.ExecutableType= uint8(buf[0x139])
  self.DiscNumber= uint8(buf[0x13a])
  self.DiscInSet= uint8(buf[0x13b])
  self.SaveGameID= _u32(buf[0x13c:])
  copy ( self.ConsoleID[:], buf[0x140:0x140+5] )
  copy ( self.ProfileID[:], buf[0x145:0x145+8] )
  if buf[0x14d] != 0x24 {
    return fmt.Errorf (
      "Error while reading STFS metadata: invalid Volume Descriptor size (%d)",
      uint8(buf[0x14d]) )
  }
  self.Volume.Read ( buf[0x14e:] )
  self.DataFileCount= int32(_u32(buf[0x171:]))
  self.DataFileCombSize= int64(_u64(buf[0x175:]))
  copy ( self.DeviceID[:], buf[0x1d1:0x1d1+0x14] )
  dec:= unicode.UTF16(unicode.BigEndian,unicode.IgnoreBOM).NewDecoder ()
  for i:= 0; i < 9; i++ {
    self.DisplayName[i]= _str(dec,buf[0x1e5+i*0x100:0x1e5+(i+1)*0x100])
  }
  for i:= 0; i < 9; i++ {
    self.DisplayDescription[i]= _str(dec,buf[0xae5+i*0x100:0xae5+(i+1)*0x100])
  }
  self.PublisherName= _str(dec,buf[0x13e5:0x13e5+0x80])
  self.TitleName= _str(dec,buf[0x1465:0x1465+0x80])
  self.TransferFlags= uint8(buf[0x14e5])
  img_size:= int32(_u32(buf[0x14e6:]))
  if img_size > 0 {
    self.Thumbnail= make([]byte,img_size)
    copy ( self.Thumbnail, buf[0x14ee:0x14ee+img_size] )
  }
  img_size= int32(_u32(buf[0x14ea:]))
  if img_size > 0 {
    self.TitleThumbnail= make([]byte,img_size)
    copy ( self.TitleThumbnail, buf[0x54ee:0x54ee+img_size] )
  }
  
  // Metadata Version 2
  if self.MetadataVersion == 2 {
    copy ( self.SeriesID[:], buf[0x185:0x185+0x10] )
    copy ( self.SeasonID[:], buf[0x195:0x195+0x10] )
    self.SeasonNumber= int16(_u16(buf[0x1a5:]))
    self.EpisodeNumber= int16(_u16(buf[0x1a9:]))
    for i:= 0; i < 3; i++ {
      self.DisplayName[i+9]= _str(dec,buf[0x51ee+i*0x100:0x51ee+(i+1)*0x100])
    }
    for i:= 0; i < 3; i++ {
      self.DisplayDescription[i+9]= _str(dec,
        buf[0x91ee+i*0x100:0x91ee+(i+1)*0x100])
    }
  }
  
  return nil

} // end STFSMetadata.Read


func (self *STFSHeader) ReadCons( buf []byte ) error {

  copy ( self.CertOwnConsoleID[:], buf[0x6:0x6+0x5] )
  self.CertOwnConsolePartNumber= string(buf[0xb:0xb+0x14])
  self.CertOwnConsoleType= buf[0x1f]
  self.CertDateGeneration= string(buf[0x20:0x20+0x8])
  copy ( self.PublicExponent[:], buf[0x28:0x28+0x4] )
  copy ( self.PublicModulus[:], buf[0x2c:0x2c+0x80] )
  copy ( self.CertSignature[:], buf[0xac:0xac+0x100] )
  copy ( self.Signature[:], buf[0x1ac:0x1ac+0x80] )
  
  return nil
  
} // end STFSHeader.ReadCons


func (self *STFSHeader) ReadPirsLive( buf []byte ) error {

  copy ( self.PackageSignature[:], buf[0x4:0x4+0x100] )
  
  return nil
  
} // end STFSHeader.ReadPirsLive


func (self *STFSHeader) Read( fd io.Reader ) error {
  
  // Llig capçalera.
  var buf [_STFS_HEADER_SIZE]byte
  n,err:= fd.Read ( buf[:] )
  if err != nil {
    return fmt.Errorf ( "Error while reading STFS header: %s", err )
  }
  if n != len(buf) {
    return errors.New ( "Error while reading STFS header: not enough bytes" )
  }

  // Tipus
  if buf[0]=='C' && buf[1]=='O' && buf[2]=='N' && buf[3]==' ' {
    self.Type= STFS_TYPE_CONS
  } else if buf[0]=='P' && buf[1]=='I' && buf[2]=='R' && buf[3]=='S' {
    self.Type= STFS_TYPE_PIRS
  } else if buf[0]=='L' && buf[1]=='I' && buf[2]=='V' && buf[3]=='E' {
    self.Type= STFS_TYPE_LIVE
  } else {
    return fmt.Errorf (
      "Error while reading STFS Header: unknown type '%c%c%c%c'",
    buf[0], buf[1], buf[2], buf[3] )
  }
  
  // Llig contingut capçalera
  if self.Type == STFS_TYPE_CONS {
    return self.ReadCons ( buf[:] )
  } else {
    return self.ReadPirsLive ( buf[:] )
  }
  
} // end STFSHeader.Read


func NewSTFS( file_name string ) (*STFS,error) {

  // Inicialitza
  ret:= STFS{
    file_name: file_name,
  }
  
  // Llig capçalera i metadades.
  fd,err:= os.Open ( file_name )
  if err != nil {
    return nil,err
  }
  defer fd.Close ()
  if err:= ret.Header.Read ( fd ); err != nil {
    return nil,err
  }
  if err:= ret.Metadata.Read ( fd ); err != nil {
    return nil,err
  }
  
  return &ret,nil
  
} // end NewSTFS


func (self *STFS) Type() string {

  switch self.Header.Type {
  case STFS_TYPE_CONS:
    return "CONS"
  case STFS_TYPE_PIRS:
    return "PIRS"
  case STFS_TYPE_LIVE:
    return "LIVE"
  default:
    return "Unknown"
  }
  
} // end STFS.Type


func (self *STFS) CertOwnConsoleType() string {
  if self.Header.Type == STFS_TYPE_CONS {
    switch self.Header.CertOwnConsoleType {
    case 1:
      return "Devkit"
    case 2:
      return "Retail"
    default:
      return fmt.Sprintf ( "Unknown (%02x)", self.Header.CertOwnConsoleType )
    }
  } else {
    return "None"
  }
  
} // end STFS.CertOwnConsoleType


func (self *STFS) ContentType() string {
  switch self.Metadata.ContentType {
  case CONTENT_TYPE_SAVED_GAME:
    return "Saved Game"
  case CONTENT_TYPE_MARKETPLACE_CONTENT:
    return "Marketplace Content"
  case CONTENT_TYPE_PUBLISHER:
    return "Publisher"
  case CONTENT_TYPE_XBOX360_TITLE:
    return "Xbox 360 Title"
  case CONTENT_TYPE_IPTV_PAUSE_BUFFER:
    return "IPTV Pause Buffer"
  case CONTENT_TYPE_INSTALLED_GAME:
    return "Installed Game"
  case CONTENT_TYPE_XBOX_TITLE:
    return "Xbox Title"
  case CONTENT_TYPE_GAME_ON_DEMAND:
    return "Game on Demand"
  case CONTENT_TYPE_AVATAR_ITEM:
    return "Avatar item"
  case CONTENT_TYPE_PROFILE:
    return "Profile"
  case CONTENT_TYPE_GAMER_PICTURE:
    return "Gamer Picture"
  case CONTENT_TYPE_THEME:
    return "Theme"
  case CONTENT_TYPE_CACHE_FILE:
    return "Cache File"
  case CONTENT_TYPE_STORAGE_DOWNLOAD:
    return "Storage Download"
  case CONTENT_TYPE_XBOX_SAVED_GAME:
    return "Xbox Saved Game"
  case CONTENT_TYPE_XBOX_DOWNLOAD:
    return "Xbox Download"
  case CONTENT_TYPE_GAME_DEMO:
    return "Game Demo"
  case CONTENT_TYPE_VIDEO:
    return "Video"
  case CONTENT_TYPE_GAME_TITLE:
    return "Game Title"
  case CONTENT_TYPE_INSTALLER:
    return "Installer"
  case CONTENT_TYPE_GAME_TRAILER:
    return "Game Trailer"
  case CONTENT_TYPE_ARCADE_TITLE:
    return "Arcade Title"
  case CONTENT_TYPE_XNA:
    return "XNA"
  case CONTENT_TYPE_LICENSE_STORE:
    return "License Store"
  case CONTENT_TYPE_MOVIE:
    return "Movie"
  case CONTENT_TYPE_TV:
    return "TV"
  case CONTENT_TYPE_MUSIC_VIDEO:
    return "Music Video"
  case CONTENT_TYPE_GAME_VIDEO:
    return "Game Video"
  case CONTENT_TYPE_PODCAST_VIDEO:
    return "Podcast Video"
  case CONTENT_TYPE_VIRAL_VIDEO:
    return "Viral Video"
  case CONTENT_TYPE_COMMUNITY_GAME:
    return "Community Game"
  default:
    return "Unknown"
  }
} // end STFS.ContentType


func (self *STFS) Platform() string {
  switch self.Metadata.Platform {
  case PLATFORM_XBOX360:
    return "Xbox 360"
  case PLATFORM_PC:
    return "PC"
  default:
    return "Unknown"
  }
} // end STFS.Platform
