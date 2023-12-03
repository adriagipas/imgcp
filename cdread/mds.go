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
 *  mds.go - Format  MDS/MDF (Alcohol 120%) 
 */

package cdread

import (
  "errors"
  "fmt"
  "io"
  "os"
  "path"
  "strings"
)




/****************/
/* TRACK READER */
/****************/

type _Mds_TrackReader struct {

  mode int
  cd   *_CD_Mds
  db   *_CD_Mds_DataBlock

  // Situació sectors
  next_sector int64
  eof         bool

  // Sector actual
  sec_data  []byte
  data      []byte // Slice de sec_data
  data_size int
  pos       int
  file      *os.File
  
}


func (self *_Mds_TrackReader) loadNextSector() error {

  // Eof
  if self.eof ||
    self.next_sector >= int64(uint64(self.db.index.index1_sectors)) {
    self.eof= true
    return nil
  }

  // Mou a la posició del sector
  offset:= self.db.offset + self.next_sector*int64(len(self.sec_data))
  if new_off,err:= self.file.Seek ( offset, 0 ); err != nil {
    return err
  } else if new_off != offset {
    return fmt.Errorf ( "failed to read sector %d", self.next_sector )
  }

  // Llig el sector.
  if n,err:= self.file.Read ( self.sec_data[:] ); err != nil {
    return err
  } else if n != len(self.sec_data) {
    return fmt.Errorf ( "failed to read sector %d", self.next_sector )
  }

  // Actualitza estat.
  self.pos= 0
  switch self.db.trackmode {
  case _CD_MDS_TRACKMODE_AUDIO:
    self.data= self.sec_data[:SECTOR_SIZE]
    self.data_size= SECTOR_SIZE
  case _CD_MDS_TRACKMODE_MODE1:
    self.data= self.sec_data[16:2064]
    self.data_size= 2048
  default:
    return fmt.Errorf ( "load sectors of type %d not implemented",
      self.db.trackmode )
  }
  self.next_sector++
  
  return nil
  
} // end loadNextSector


func (self *_Mds_TrackReader) Close() error {
  return self.file.Close ()
} // end Close


func (self *_Mds_TrackReader) Read( b []byte ) (n int,err error) {
  
  // EOF
  if self.eof { return 0,io.EOF }
  
  // Llig
  pos,remain:= 0,len(b)
  for remain > 0 && !self.eof {
    
    // Recarrega si cal
    // ATENCIÓ!! Si els sectors no són d'aquesta grandària podria
    // fallar. Ara sols suporte RAW sectors.
    for self.pos >= self.data_size && !self.eof {
      if err:= self.loadNextSector (); err != nil {
        return 0,err
      }
    }
    
    // Llig
    if !self.eof {
      // --> Bytes a llegir
      avail:= self.data_size-self.pos
      var nbytes int
      if remain > avail {
        nbytes= avail
      } else {
        nbytes= remain
      }
      // --> Còpia
      copy ( b[pos:pos+nbytes], self.data[self.pos:self.pos+nbytes])
      // --> Actualitza
      pos+= nbytes
      remain-= nbytes
      self.pos+= nbytes
    }
    
  }
  
  return pos,nil
  
} // end Read


func (self *_Mds_TrackReader) Seek( sector int64 ) error {

  // Actualitza estat
  self.eof= false
  self.pos= len(self.sec_data)
  self.next_sector= sector
  
  // Intenta carregar
  if err:= self.loadNextSector (); err != nil {
    return err
  }
  
  return nil
  
} // end Seek




/******/
/* CD */
/******/

type _CD_Mds_Index struct {
  
  index0_sectors uint32
  index1_sectors uint32
  
}

const (
  _CD_MDS_TRACKMODE_NONE             = 0
  _CD_MDS_TRACKMODE_AUDIO            = 1
  _CD_MDS_TRACKMODE_MODE1            = 2
  _CD_MDS_TRACKMODE_MODE2            = 3
  _CD_MDS_TRACKMODE_MODE2_SUBCHANNEL = 4
)

type _CD_Mds_DataBlock struct {

  // Comú
  trackmode    int
  subchannel   bool // Si és cert el sector inclou 0x60 bytes
  addr_control uint8
  point        uint8
  minute       uint8
  second       uint8
  frame        uint8

  // Informació sobre dades
  index       _CD_Mds_Index
  sector_size uint16 // Inclou extra
  start       uint32 // Primer sector (si no hi ha índex 0 el primer
                     // sector és el 00:02:00 ???)
  offset      int64 // Offset dins del fitxer
  file_names  []string // Fitxers on llegir per ordre
  
}

type _CD_Mds_Session struct {

  id                 uint16
  start_sector       int32
  end_sector         int32
  total_data_blocks  uint8 // Tots els data blocs
  data_blocks_leadin uint8 // Sols els que són >=A0
  first_track_num    uint16
  last_track_num     uint16
  data               []_CD_Mds_DataBlock
  
}

const (
  _CD_MDS_MEDIA_TYPE_CDROM  = 0
  _CD_MDS_MEDIA_TYPE_CDR    = 1
  _CD_MDS_MEDIA_TYPE_CDRW   = 2
  _CD_MDS_MEDIA_TYPE_DVDROM = 3
  _CD_MDS_MEDIA_TYPE_DCDR   = 4
)

type _CD_Mds struct {

  file_name   string
  media_type  int
  sessions    []_CD_Mds_Session
  default_mdf string // MDF per defecte

}


func (self *_CD_Mds_Index) read( f *os.File, offset uint32 ) error {

  // Llig block.
  var buf [0x8]byte
  if _,err:= f.Seek ( int64(uint64(offset)), 0 ); err != nil {
    return err
  }
  if nr,err:= f.Read ( buf[:] ); err != nil {
    return err
  } else if nr!=len(buf) {
    return fmt.Errorf ( "failed to read index block in %X", offset )
  }

  // Valors
  self.index0_sectors= uint32(buf[0]) |
    (uint32(buf[1])<<8) |
    (uint32(buf[2])<<16) |
    (uint32(buf[3])<<24)
  self.index1_sectors= uint32(buf[4]) |
    (uint32(buf[5])<<8) |
    (uint32(buf[6])<<16) |
    (uint32(buf[7])<<24)

  return nil
  
} // end read


func (self *_CD_Mds_DataBlock) readFileName(
  
  f      *os.File,
  offset uint32,
  
) error {
  
  // Llig block.
  var buf_file [0x10]byte
  if _,err:= f.Seek ( int64(uint64(offset)), 0 ); err != nil {
    return err
  }
  if nr,err:= f.Read ( buf_file[:] ); err != nil {
    return err
  } else if nr!=len(buf_file) {
    return fmt.Errorf ( "failed to read file block in %X", offset )
  }

  // Parseja valors
  offset_name:= uint32(buf_file[0]) |
    (uint32(buf_file[1])<<8) |
    (uint32(buf_file[2])<<16) |
    (uint32(buf_file[3])<<24)
  var format8b bool
  switch buf_file[4] {
  case 0:
    format8b= true
  case 1:
    format8b= false
  default:
    return fmt.Errorf ( "failed to read file block in %X: unsuported"+
      " filename format %X", offset, buf_file[4] )
  }
  
  // Read name
  if _,err:= f.Seek ( int64(uint64(offset_name)), 0 ); err != nil {
    return err
  }
  var file_name string
  if format8b {
    var buf [1]byte
    var tmp []byte= nil
    buf[0]= 0xff
    for ; buf[0] != 0; {
      if nr,err:= f.Read ( buf[:] ); err != nil {
        return err
      } else if nr!=len(buf) {
        return fmt.Errorf ( "failed to read file name in %X", offset_name )
      }
      if buf[0] != 0 { tmp= append(tmp,buf[0]) }
    }
    file_name= string(tmp)
  } else {
    var buf [2]byte
    var tmp []rune= nil
    var c rune= rune(-1)
    for ; c != 0; {
      if nr,err:= f.Read ( buf[:] ); err != nil {
        return err
      } else if nr!=len(buf) {
        return fmt.Errorf ( "failed to read file name in %X", offset_name )
      }
      c= rune(uint32(buf[0]) | (uint32(buf[1])<<8))
      if c != 0 { tmp= append(tmp,c) }
    }
    file_name= string(tmp)
  }
  if file_name == "*.mdf" {
    file_name= ""
  }
  self.file_names= append(self.file_names,file_name)
  
  return nil
  
} // end readFileName


func (self *_CD_Mds_Session) readDataBlocks( f *os.File, offset uint32 ) error {

  var buf [0x50]byte
  var d _CD_Mds_DataBlock
  var index_offset uint32
  var num_file_names uint32
  var offset_file uint32
  for i:= 0; i < int(self.total_data_blocks); i++ {

    // Llig
    if _,err:= f.Seek ( int64(uint64(offset)), 0 ); err != nil { return err }
    if nr,err:= f.Read ( buf[:] ); err != nil {
      return err
    } else if nr!=0x50 {
      return fmt.Errorf ( "failed to read data block %d for session %d", i+1,
        self.id )
    }
    offset+= 0x50

    // Trackmode
    switch buf[0] {
    case 0x00:
      d.trackmode= _CD_MDS_TRACKMODE_NONE
    case 0xa9:
      d.trackmode= _CD_MDS_TRACKMODE_AUDIO
    case 0xaa:
      d.trackmode= _CD_MDS_TRACKMODE_MODE1
    case 0xab:
      d.trackmode= _CD_MDS_TRACKMODE_MODE2
    case 0xec:
      d.trackmode= _CD_MDS_TRACKMODE_MODE2_SUBCHANNEL
    default:
      return fmt.Errorf ( "failed to read data block %d for session %d:"+
        " unknown trackmode %X", i+1, self.id, buf[0] )
    }

    // Altres comuns
    d.subchannel= buf[1]==8
    d.addr_control= uint8(buf[2])
    d.point= uint8(buf[4])
    d.minute= uint8(buf[0x9])
    d.second= uint8(buf[0xa])
    d.frame= uint8(buf[0xb])

    // Informació sobre dades
    d.index.index0_sectors= 0
    d.index.index1_sectors= 0
    d.sector_size= 0
    d.start= 0
    d.offset= 0
    d.file_names= nil
    if d.point < 0xa0 {

      // Index offset
      index_offset= uint32(buf[0xc]) |
        (uint32(buf[0xd])<<8) |
        (uint32(buf[0xe])<<16) |
        (uint32(buf[0xf])<<24)
      if err:= d.index.read ( f, index_offset ); err != nil {
        return err
      }
      
      // Sector size
      d.sector_size= uint16(buf[0x10]) | (uint16(buf[0x11])<<8)
      if d.sector_size < 0x800 || d.sector_size > 0x990 {
        return fmt.Errorf ( "failed to read data block %d for session %d:"+
          " wrong sector size", i+1, self.id, d.sector_size )
      }

      // Start i offset
      d.start= uint32(buf[0x24]) |
        (uint32(buf[0x25])<<8) |
        (uint32(buf[0x26])<<16) |
        (uint32(buf[0x27])<<24)
      d.offset= int64(uint64(buf[0x28]) |
        (uint64(buf[0x29])<<8) |
        (uint64(buf[0x2a])<<16) |
        (uint64(buf[0x2b])<<24) |
        (uint64(buf[0x2c])<<32) |
        (uint64(buf[0x2d])<<40) |
        (uint64(buf[0x2e])<<48) |
        (uint64(buf[0x2f])<<56))
      if d.offset < 0 {
        return fmt.Errorf ( "failed to read data block %d for session %d:"+
          " negative offset %d", i+1, self.id, d.offset )
      }

      // Llig filenames
      num_file_names= uint32(buf[0x30]) |
        (uint32(buf[0x31])<<8) |
        (uint32(buf[0x32])<<16) |
        (uint32(buf[0x33])<<24)
      offset_file= uint32(buf[0x34]) |
        (uint32(buf[0x35])<<8) |
        (uint32(buf[0x36])<<16) |
        (uint32(buf[0x37])<<24)
      for j:= uint32(0); j < num_file_names; j++ {
        if err:= d.readFileName ( f, offset_file ); err != nil {
          return err
        }
        offset_file+= 0x10
      }
      
    }
    
    self.data= append(self.data,d)
    
  }
  
  return nil
  
} // end readDataBlocks


func (self *_CD_Mds) readSessions( f *os.File, offset uint32 ) error {

  var buf [0x18]byte
  var s *_CD_Mds_Session
  var data_offset uint32
  for i:= 0; i < len(self.sessions); i++ {

    s= &self.sessions[i]

    // Llig
    if _,err:= f.Seek ( int64(uint64(offset)), 0 ); err != nil { return err }
    if nr,err:= f.Read ( buf[:] ); err != nil {
      return err
    } else if nr!=0x18 {
      return fmt.Errorf ( "failed to read session block %d", i+1 )
    }
    offset+= 0x18
    
    // Sectors
    s.start_sector= int32(uint32(buf[0]) |
      (uint32(buf[1])<<8) |
      (uint32(buf[2])<<16) |
      (uint32(buf[3])<<24))
    s.end_sector= int32(uint32(buf[4]) |
      (uint32(buf[5])<<8) |
      (uint32(buf[6])<<16) |
      (uint32(buf[7])<<24))

    // Session number
    s.id= uint16(buf[8]) | (uint16(buf[9])<<8)
    if s.id != uint16(i+1) {
      return fmt.Errorf ( "failed to read session block %d:"+
        " session number is %d", i+1, s.id )
    }
    
    // Data blocks
    s.total_data_blocks= uint8(buf[0xa])
    s.data_blocks_leadin= uint8(buf[0xb])

    // Tracks
    s.first_track_num= uint16(buf[0xc]) | (uint16(buf[0xd])<<8)
    if s.first_track_num < 1 || s.first_track_num > 0x63 {
      return fmt.Errorf ( "failed to read session block %d: wrong first track"+
        " number %X", i+1, s.first_track_num )
    }
    s.last_track_num= uint16(buf[0xe]) | (uint16(buf[0xf])<<8)
    if s.last_track_num < 1 || s.last_track_num > 0x63 {
      return fmt.Errorf ( "failed to read session block %d: wrong last track"+
        " number %X", i+1, s.last_track_num )
    }
    if s.first_track_num > s.last_track_num {
      return fmt.Errorf ( "failed to read session block %d: %X > %X",
        i+1, s.first_track_num, s.last_track_num )
    }

    // Llig dades
    data_offset= uint32(buf[0x14]) |
      (uint32(buf[0x15])<<8) |
      (uint32(buf[0x16])<<16) |
      (uint32(buf[0x17])<<24)
    if err:= s.readDataBlocks ( f, data_offset ); err != nil {
      return err
    }
    
  }
  
  return nil
  
} // end readSessions


func (self *_CD_Mds) init( f *os.File ) error {

  // Llig capçalera
  var buf [0x58]byte
  if nb,err:= f.Read ( buf[:] ); err != nil {
    return err
  } else if nb != 0x58 {
    return fmt.Errorf ( "unable to read MDS header from file %s",
      self.file_name )
  }

  // Comprova capçalera
  id:= string(buf[:16])
  if id != "MEDIA DESCRIPTOR" && buf[11]!=1 &&
    (buf[12]!=3 || buf[12]!=4 || buf[12]!=5) {
    return errors.New ( "%s is not a MDS/MDF file" )
  }

  // Llig valors capçalera
  media_type:= uint16(buf[0x12]) | (uint16(buf[0x13])<<8)
  switch media_type {
  case 0:
    self.media_type= _CD_MDS_MEDIA_TYPE_CDROM
  case 1:
    self.media_type= _CD_MDS_MEDIA_TYPE_CDR
  case 2:
    self.media_type= _CD_MDS_MEDIA_TYPE_CDRW
  case 0x10:
    self.media_type= _CD_MDS_MEDIA_TYPE_DVDROM
  case 0x12:
    self.media_type= _CD_MDS_MEDIA_TYPE_DCDR
  default:
    return fmt.Errorf ( "unknown media type: %02", buf[0x12] )
  }
  num_sessions:= uint16(buf[0x14]) | (uint16(buf[0x15])<<8)
  if num_sessions == 0 {
    return errors.New ( "number of sessions is 0" )
  }

  // Llig sessions
  offset:= uint32(buf[0x50]) |
    (uint32(buf[0x51])<<8) |
    (uint32(buf[0x52])<<16) |
    (uint32(buf[0x53])<<24)
  self.sessions= make([]_CD_Mds_Session,num_sessions)
  if err:= self.readSessions ( f, offset ); err != nil { return err }

  // Default mdf
  ext:= path.Ext(self.file_name)
  if strings.ToLower(ext) == ".mds" {
    self.default_mdf,_= strings.CutSuffix(self.file_name,ext)
    self.default_mdf+= ".mdf"
  } else {
    self.default_mdf= self.file_name+".mdf"
  }
  
  return nil
  
} // end init


func (self *_CD_Mds) Format() string { return "MDS/MDF (Alcohol 120%)" }


func (self *_CD_Mds) Info() *Info {

  var s *_CD_Mds_Session
  
  // Calcula tracks totals
  var num_tracks int= 0
  for i:= 0; i < len(self.sessions); i++ {
    s= &self.sessions[i]
    for j:= 0; j < len(s.data); j++ {
      if s.data[j].trackmode != _CD_MDS_TRACKMODE_NONE {
        num_tracks++
      }
    }
  }

  // Crea
  ret:= Info{}
  ret.Sessions= make([]SessionInfo,len(self.sessions))
  ret.Tracks= make([]TrackInfo,num_tracks)

  // Crea tracks
  var pos int= 0
  var beg_pos int
  var db *_CD_Mds_DataBlock
  var track *TrackInfo
  for i:= 0; i < len(self.sessions); i++ {
    s= &self.sessions[i]
    beg_pos= pos
    for j:= 0; j < len(s.data); j++ {
      db= &s.data[j]
      if db.trackmode != _CD_MDS_TRACKMODE_NONE {

        // Track
        track= &ret.Tracks[pos]
        pos++

        // Id
        track.Id= BCD ( int(uint32(db.point)) )

        // Indexes i PosLastSector
        start_sect:= db.start
        num_indexes:= 0
        if db.index.index0_sectors>0 { num_indexes++ }
        if db.index.index1_sectors>0 { num_indexes++ }
        if num_indexes > 0 {
          track.Indexes= make([]IndexInfo,num_indexes)
          if db.index.index0_sectors>0 {
            track.Indexes[0].Id= 0
            track.Indexes[0].Pos= GetPosition ( int64(uint64(start_sect)) )
            start_sect+= db.index.index0_sectors
          } else {
            start_sect+= 2*75
          }
          if db.index.index1_sectors>0 {
            track.Indexes[num_indexes-1].Id= 1
            track.Indexes[num_indexes-1].Pos=
              GetPosition ( int64(uint64(start_sect)) )
            start_sect+= db.index.index0_sectors
          }
        }
        track.PosLastSector= GetPosition( int64(uint64(start_sect-1)) )

        // Tipus
        switch db.trackmode {
        case _CD_MDS_TRACKMODE_AUDIO:
          track.Type= TRACK_TYPE_AUDIO
        case _CD_MDS_TRACKMODE_MODE1:
          if db.sector_size >= SECTOR_SIZE {
            track.Type= TRACK_TYPE_MODE1_RAW
          } else {
            track.Type= TRACK_TYPE_UNK
          }
        default:
          track.Type= TRACK_TYPE_UNK
        }
        
      }
    }
    ret.Sessions[i].Tracks= ret.Tracks[beg_pos:pos]
  }
  
  return &ret
  
}


func (self *_CD_Mds) TrackReader(
  
  session_id int,
  track_id   int,
  mode       int,
  
) (TrackReader,error) {

  // Selecciona sessió
  if session_id >= len(self.sessions) {
    return nil,fmt.Errorf ( "session (%d) out of range", session_id )
  }
  sess:= &self.sessions[session_id]

  // Selecciona track (bloc de dades). Busquem per si de cas.
  tmp_id:= track_id
  var db *_CD_Mds_DataBlock= nil
  for i:= 0; i < len(sess.data) && db == nil; i++ {
    if sess.data[i].trackmode != _CD_MDS_TRACKMODE_NONE {
      if tmp_id == 0 {
        db= &sess.data[i]
      } else {
        tmp_id--
      }
    }
  }
  if db == nil {
    return nil,fmt.Errorf ( "track (%d) out of range", track_id )
  }
  
  // Comprova que és un bloc de dades suportat
  if len(db.file_names) == 0 || db.index.index1_sectors == 0 {
    return nil,fmt.Errorf ( "track (%d) empty", track_id )
  }
  if len(db.file_names) > 1 {
    return nil,fmt.Errorf ( "multiple files (%d) not supported",
      len(db.file_names) )
  }

  // Crea el TrackReader
  ret:= _Mds_TrackReader{
    mode        : mode,
    cd          : self,
    db          : db,
    next_sector : 0,
    eof         : false,
    sec_data    : make([]byte,int(uint32(db.sector_size))),
    data        : nil,
    data_size   : 0,
    pos         : int(uint32(db.sector_size)),
  }
  file_name:= db.file_names[0]
  if file_name == "" { file_name= self.default_mdf }
  var err error
  ret.file,err= os.Open ( file_name )
  if err != nil { return nil,err }

  return &ret,nil
  
} // end TrackReader




/**********************/
/* FUNCIONS PÚBLIQUES */
/**********************/

func OpenMds( file_name string ) (CD,error) {

  // Intenta obrir el fitxer.
  f,err:= os.Open ( file_name )
  if err != nil { return nil,err }
  defer f.Close ()

  // Crea i inicialitza.
  ret:= _CD_Mds{
    file_name : file_name,
  }
  if err:= ret.init ( f ); err != nil {
    return nil,err
  }
  
  return &ret,nil
  
} // end OpenMds
