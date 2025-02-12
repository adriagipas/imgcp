/*
 * Copyright 2023-2025 Adrià Giménez Pastor.
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
  "io"
  "strings"
)




/****************/
/* PART PRIVADA */
/****************/

const LOGICAL_SECTOR_SIZE = 2048


// FUNCIONS ////////////////////////////////////////////////////////////////////

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


func parse_date_time_record( data []byte, dt *ISO_DateTimeRecord ) error {

  // Comprova si està buit
  dt.Empty= true
  for _,v := range data {
    if v != 0 {
      dt.Empty= false
      break
    }
  }
  if dt.Empty { return nil }

  // Ompli
  dt.Year= 1900 + int(uint32(uint8(data[0])))
  dt.Month= uint8(data[1])
  if dt.Month < 1 || dt.Month > 12 {
    return fmt.Errorf ( "error while reading data and time record: "+
      "wrong month value (%d)", dt.Month )
  }
  dt.Day= uint8(data[2])
  if dt.Day < 1 || dt.Day > 31 {
    return fmt.Errorf ( "error while reading data and time record: "+
      "wrong day value (%d)", dt.Day )
  }
  dt.Hour= uint8(data[3])
  if dt.Hour > 23 {
    return fmt.Errorf ( "error while reading data and time record: "+
      "wrong hour value (%d)", dt.Hour )
  }
  dt.Minute= uint8(data[4])
  if dt.Minute > 59 {
    return fmt.Errorf ( "error while reading data and time record: "+
      "wrong minute value (%d)", dt.Minute )
  }
  dt.Second= uint8(data[5])
  if dt.Second > 59 {
    return fmt.Errorf ( "error while reading data and time record: "+
      "wrong second value (%d)", dt.Second )
  }
  dt.GMT= int(data[6])-48
  
  return nil
  
} // end parse_date_time_record


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


// ISO FILE ENTRY //////////////////////////////////////////////////////////////

type _ISO_FileEntry struct {
  
  recording_date_time ISO_DateTimeRecord
  offset              uint32 // Logical block on comença
  size                uint32 // Grandària
  flags               uint8
  file_unit_size      uint8 // Grandària File Unit en LB. 0 -> no interleave
  gap_size            uint8 // Si interleave, grandària del gap
  volume              uint16
  id                  string
  
}

func (self *_ISO_FileEntry) read( data []byte ) error {

  // Longitut
  if len(data) == 0 {
    return errors.New ( "trying to load an empty file entry record" )
  }
  len_dr:= uint8(data[0])
  if len_dr<34 {
    return fmt.Errorf ( "wrong directory entry format: LEN-DR = %d", len_dr )
  }

  // Extended attribute record length
  eattr_len:= uint8(data[1])
  if eattr_len != 0 {
    return errors.New ( "TODO - Extended Attribute Record Length" )
  }

  // Posició i grandària del Extent
  self.offset= parse_int32_LSB_MSB ( data[2:10] )
  self.size= parse_int32_LSB_MSB ( data[10:18] )
  
  // Recording Date and Time
  if err:= parse_date_time_record ( data[18:25],
    &self.recording_date_time ); err != nil {
    return err
  }
  
  // Flags
  self.flags= uint8(data[25])
  if (self.flags&FILE_FLAGS_MULTIEXTENT)!=0 {
    return errors.New ( "TODO - Multiextent" )
  }
  
  // Unit size i interlave
  self.file_unit_size= uint8(data[26])
  self.gap_size= uint8(data[27])

  // Volume
  self.volume= parse_int16_LSB_MSB ( data[28:32] )

  // Identificador
  file_size:= uint8(data[32])
  if file_size == 0 {
    return errors.New (
      "trying to load a file entry record without identifier" )
  } else if file_size == 1 &&
    (self.flags&FILE_FLAGS_DIRECTORY)!= 0 &&
    (data[33] == 0 || data[33] == 1 ) {
    if data[33] == 0 {
      self.id= "."
    } else {
      self.id= ".."
    }
  } else {
    self.id= string(data[33:33+file_size])
  }
  
  return nil
  
} // end read


// ISO FILE READER /////////////////////////////////////////////////////////////

type _ISO_FileReader struct {

  iso             *ISO
  current_lb      uint32
  offset          uint32 // Offset én bytes inicial, típicament 0
  remain          uint32 // Bytes per llegir
  file_unit_size  uint8
  gap_size        uint8
  current_file_lb uint8
  buf             []byte
  f               TrackReader
  
}


func (self *_ISO_FileReader) loadBufNoInterleave() error {

  var err error
  self.buf,err= self.iso.readLogicalBlock ( self.f, self.current_lb )
  if err != nil { return err }
  self.current_lb++

  return nil
  
} // end loadBufNoInterleave


func (self *_ISO_FileReader) loadBufInterleave() error {
  return errors.New ( "TODO - ISO_FileReader.loadBufInteleave" )
} // end loadBufInterleave


func (self *_ISO_FileReader) loadBuf() error {
  if self.file_unit_size == 0 {
    return self.loadBufNoInterleave ()
  } else {
    return self.loadBufInterleave ()
  }
} // end loadBuf


func (self *_ISO_FileReader) Close() error {
  return self.f.Close ()
} // end Close


func (self *_ISO_FileReader) Read( data []byte) (n int,err error) {

  // Prepara
  if self.remain == 0 { return 0,io.EOF }
  n,err= 0,nil
  
  // Ignora offset
  for self.offset > 0 {
    
    // Obté dades
    if len(self.buf)==0 {
      err= self.loadBuf ()
      if err != nil { return }
    }

    // Ignora bytes
    if self.offset>uint32(len(self.buf)) {
      self.offset-= uint32(len(self.buf))
      self.buf= nil
    } else {
      self.buf= self.buf[self.offset:]
      self.offset= 0
    }
    
  }

  // Llig.
  var nbytes uint32
  var buf_size uint32
  var want_size uint32
  for len(data)>0 && self.remain>0 {

    // Obté dades
    if len(self.buf)==0 {
      err= self.loadBuf ()
      if err != nil { return }
    }
    
    // Bytes a llegir
    buf_size= uint32(len(self.buf))
    if self.remain>buf_size {
      nbytes= buf_size
    } else {
      nbytes= self.remain
    }
    want_size= uint32(len(data))
    if nbytes>want_size {
      nbytes= want_size
    }

    // Llig
    copy(data[:nbytes],self.buf[:nbytes])
    data= data[nbytes:]
    self.buf= self.buf[nbytes:]
    self.remain-= nbytes
    n+= int(nbytes)
    
  }

  return
  
} // end Read




/****************/
/* PART PÚBLICA */
/****************/

// ISO ////////////////////////////////////////////////////////////////////////

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

type ISO_DateTimeRecord struct {

  Year   int
  Month  uint8
  Day    uint8
  Hour   uint8
  Minute uint8
  Second uint8
  GMT    int
  Empty  bool
  
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
  blocks_per_sec  int
  
}

type ISO_SupplementaryVolume struct {
  ISO_PrimaryVolume
  
  Flags uint8 // Bit 0: 0 -> only escape sequence ISO 2375; 1 -> at
              // least one escape sequence not ISO 2375
  
}

const (

  FILE_FLAGS_EXISTENCE       = 0x01
  FILE_FLAGS_DIRECTORY       = 0x02
  FILE_FLAGS_ASSOCIATED_FILE = 0x04
  FILE_FLAGS_RECORD          = 0x08
  FILE_FLAGS_PROTECTED       = 0x10
  FILE_FLAGS_MULTIEXTENT     = 0x80
  
)

// Segueix una aproximació greedy.
type ISO struct {

  // Públic
  PrimaryVolume ISO_PrimaryVolume
  Supplementary *ISO_SupplementaryVolume // Pot ser nil
  
  // Privat
  cd          CD
  session     int
  track       int
  current_sec int64
  buffer      [LOGICAL_SECTOR_SIZE]byte
  
}


func ReadISO( cd CD, session int, track int ) (*ISO,error) {

  ret:= ISO{
    Supplementary : nil,
    cd : cd,
    session : session,
    track : track,
    current_sec : -1,
  }
  
  // Parse volume descriptors.
  f,err:= cd.TrackReader ( session, track, 0 )
  if err != nil { return nil,err }
  defer f.Close ()
  if err:= ret.readVolumeDescriptors ( f ); err != nil {
    return nil,err
  }
  
  return &ret,nil
  
} // end ReadISO


func (self *ISO) readVolumeDescriptors( f TrackReader ) error {

  var buf [LOGICAL_SECTOR_SIZE]byte
  sector,end,num_pv,num_sv:= int64(0x10),false,0,0
  for ; !end; sector++ {

    // Prova a llegir
    if err:= f.Seek ( sector ); err != nil {
      return err
    }
    if nbytes,err:= f.Read ( buf[:] ); err != nil {
      return err
    } else if nbytes != LOGICAL_SECTOR_SIZE {
      return fmt.Errorf ( "failed to read volume descriptor at sector %d",
        sector )
    }

    // Comprova tipus
    switch dtype:= buf[0]; dtype {
    case 0: // Boot record
      return fmt.Errorf ( "TODO - BOOT RECORD!!!" )
    case 1: // Primary volume (NOTA!! Com a mínim cal 1)
      if num_pv == 0 {
        num_pv= 1
        if err:= self.readPrimaryVolume ( buf[:] ); err != nil {
          return err
        }
      }
    case 2: // Supplementary volume
      if num_sv == 0 {
        num_sv= 1
        if err:= self.readSupplementaryVolume ( buf[:] ); err != nil {
          return err
        }
      }
    case 3: // Volume partition
      return errors.New ( "volume partition descriptor not implemented" )
    case 255:
      end= true
    default:
      return fmt.Errorf ( "unknown volume descriptor type: %d", dtype )
      
    }
    
  }
  if num_pv == 0 {
    return errors.New ( "primary volume not found" )
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
  if self.PrimaryVolume.LogicalBlockSize<512 ||
    self.PrimaryVolume.LogicalBlockSize>LOGICAL_SECTOR_SIZE ||
    LOGICAL_SECTOR_SIZE%self.PrimaryVolume.LogicalBlockSize!=0 {
    return fmt.Errorf ( "wrong Logical Block Size: %d",
      self.PrimaryVolume.LogicalBlockSize )
  }
  self.PrimaryVolume.blocks_per_sec= 
    LOGICAL_SECTOR_SIZE/int(uint32(self.PrimaryVolume.LogicalBlockSize))

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
  
} // end readPrimaryVolume


func (self *ISO) readSupplementaryVolume( data []byte ) error {

  // Signatura
  if data[1]!='C' || data[2]!='D' || data[3]!='0' ||
    data[4]!='0' || data[5]!='1' {
    return errors.New ( "Volume descriptor signature 'CD001' not "+
      "found in supplementary descriptor" )
  }

  // Reserva
  sup:= ISO_SupplementaryVolume{}
  
  // Versió
  sup.Version= uint8(data[6])

  // VolumeFlags
  sup.Flags= uint8(data[7])

  // Escape sequence (ECMA-35 15.4 No acabe d'entendre)
  //fmt.Printf("ESCAPE [%v]\n",data[88:120])
  
  // Identificadors
  sup.SystemIdentifier=
    strings.TrimRight ( string(data[8:40]), " " )
  sup.VolumeIdentifier=
    strings.TrimRight ( string(data[40:72]), " " )
  
  // Grandàries
  sup.VolumeSpaceSize= parse_int32_LSB_MSB ( data[80:88] )
  sup.VolumeSetSize= parse_int16_LSB_MSB ( data[120:124] )
  sup.VolumeSequenceNumber= parse_int16_LSB_MSB ( data[124:128] )
  sup.LogicalBlockSize= parse_int16_LSB_MSB ( data[128:132] )
  if sup.LogicalBlockSize<512 ||
    sup.LogicalBlockSize>LOGICAL_SECTOR_SIZE ||
    LOGICAL_SECTOR_SIZE%sup.LogicalBlockSize!=0 {
    return fmt.Errorf ( "wrong Logical Block Size: %d",
      sup.LogicalBlockSize )
  }
  sup.blocks_per_sec= 
    LOGICAL_SECTOR_SIZE/int(uint32(sup.LogicalBlockSize))

  // Root directory record
  copy(sup.root_dir_record[:],data[156:190])
  
  // Més identificadors
  sup.VolumeSetIdentifier=
    strings.TrimRight ( string(data[190:318]), " " )
  sup.PublisherIdentifier=
    strings.TrimRight ( string(data[318:446]), " " )
  sup.DataPreparerIdentifier=
    strings.TrimRight ( string(data[446:574]), " " )
  sup.ApplicationIdentifier=
    strings.TrimRight ( string(data[574:702]), " " )
  sup.CopyrightFileIdentifier=
    strings.TrimRight ( string(data[702:739]), " " )
  sup.AbstractFileIdentifier=
    strings.TrimRight ( string(data[739:776]), " " )
  sup.BiblioFileIdentifier=
    strings.TrimRight ( string(data[776:813]), " " )

  // Dates
  parse_date_time ( data[813:830], &sup.VolumeCreation )
  parse_date_time ( data[830:847], &sup.VolumeModification )
  parse_date_time ( data[847:864], &sup.VolumeExpiration )
  parse_date_time ( data[864:881], &sup.VolumeEffective )

  // FileStructureVersion
  sup.FileStructureVersion= uint8(data[881])

    // Assigna
  self.Supplementary= &sup
  
  return nil
  
} // end readSupplementaryVolume


func (self *ISO) readDirectory( entry *_ISO_FileEntry ) (*ISO_Directory,error) {

  // Inicialitza
  ret:= ISO_Directory{
    iso : self,
  }

  // Comprovacions i carrega dades
  if (entry.flags&FILE_FLAGS_DIRECTORY)==0 {
    return nil,errors.New ( "failed to load directory entry: it "+
      "is marked as not directory" )
  }
  if entry.size == 0 {
    return nil,errors.New ( "failed to load directory entry: empty content" )
  }

  // Llig contingut
  ret.content= make([]byte,entry.size)
  fr,err:= self.getFileReader ( entry.offset, entry.size, 0,
    entry.file_unit_size, entry.gap_size )
  if err != nil { return nil,err }
  defer fr.Close ()
  if nb,err:= fr.Read ( ret.content ); err != nil {
    return nil,err
  } else if uint32(nb) != entry.size {
    return nil,fmt.Errorf ( "failed to load directory content: expected "+
      "%d bytes but instead %d bytes were read", entry.size, nb )
  }
  
  return &ret,nil
  
} // end readDirectory


func (self *ISO) getFileReader(

  logical_block  uint32,
  nbytes         uint32,
  offset_bytes   uint32,
  file_unit_size uint8,
  gap_size       uint8,

) (*_ISO_FileReader,error) {

  ret:= _ISO_FileReader{
    iso : self,
    current_lb : logical_block,
    offset : offset_bytes,
    remain : nbytes,
    file_unit_size : file_unit_size,
    gap_size : gap_size,
    current_file_lb : 0,
    buf : nil,
  }

  var err error
  ret.f,err= self.cd.TrackReader ( self.session, self.track, 0 )
  if err != nil { return nil,err }
  
  return &ret,nil
  
} // end getFileReader


// Torna un punter al logical block llegit
func (self *ISO) readLogicalBlock(

  f             TrackReader,
  logical_block uint32,

) ([]byte,error) {

  
  // Llig sector
  sector:= int64(logical_block/uint32(self.PrimaryVolume.blocks_per_sec))
  if sector != self.current_sec {
    if err:= f.Seek ( sector ); err != nil {
      return nil,err
    }
    if nb,err:= f.Read ( self.buffer[:] ); err != nil {
      return nil,err
    } else if nb != LOGICAL_SECTOR_SIZE {
      return nil,fmt.Errorf ( "failed to read sector %d", sector )
    }
    self.current_sec= sector
  }

  // Selecciona bloc
  block:= logical_block%uint32(self.PrimaryVolume.blocks_per_sec)
  ret:= self.buffer[uint32(self.PrimaryVolume.LogicalBlockSize)*block:
    uint32(self.PrimaryVolume.LogicalBlockSize)*(block+1)]

  return ret,nil
  
} // end readLogicalBlock


// Torna el directori arrel.
func (self *ISO) Root() (*ISO_Directory,error) {

  var entry _ISO_FileEntry
  if err:= entry.read ( self.PrimaryVolume.root_dir_record[:] ); err != nil {
    return nil,err
  }
  
  return self.readDirectory ( &entry )
  
} // end Root


// ISO DIRECTORY //////////////////////////////////////////////////////////////

type ISO_Directory struct {

  iso     *ISO
  content []byte
  
}

func (self *ISO_Directory) Begin() (*ISO_DirectoryIter,error) {

  ret:= ISO_DirectoryIter{
    dir : self,
    p : self.content,
  }
  if err:= ret.e.read ( ret.p ); err != nil {
    return nil,err
  }
  
  return &ret,nil
  
} // end Begin


// ISO DIRECTORY ITER //////////////////////////////////////////////////////////

type ISO_DirectoryIter struct {

  // PRIVAT!!!
  dir *ISO_Directory
  e   _ISO_FileEntry
  p   []byte
  
}


func (self *ISO_DirectoryIter) DateTime() *ISO_DateTimeRecord {
  return &self.e.recording_date_time
} // end DateTime


func (self *ISO_DirectoryIter) End() (end bool) {
  
  if len(self.p)==0 || self.p[0]==0 {
    end= true
  } else {
    end= false
  }

  return
  
} // end End


func (self *ISO_DirectoryIter) Flags() uint8 { return self.e.flags }

func (self *ISO_DirectoryIter) GetDirectory() (*ISO_Directory,error) {


  if (self.Flags()&FILE_FLAGS_DIRECTORY)==0 {
    return nil,fmt.Errorf ( "trying to access regular file '%s' as directory",
      self.Id () )
  }
  
  return self.dir.iso.readDirectory ( &self.e )
  
} // end GetDirectory


func (self *ISO_DirectoryIter) GetFileReader() (*_ISO_FileReader,error) {
  return self.dir.iso.getFileReader ( self.e.offset, self.e.size, 0,
    self.e.file_unit_size, self.e.gap_size )
} // end GetFileReader


func (self *ISO_DirectoryIter) Id() string { return self.e.id }


func (self *ISO_DirectoryIter) Next() error {
  
  if self.End () {
    return errors.New ( "reached end of directory entries" )
  }
  
  // Mou al següent
  self.p= self.p[uint8(self.p[0]):]
  if !self.End () {
    if err:= self.e.read ( self.p ); err != nil {
      return err
    }
  }
  
  return nil
  
} // end Next


func (self *ISO_DirectoryIter) Size() uint32 { return self.e.size }
