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
 *  cd.go - CD-Rom image.
 *
 */

package imgs

import (
  "errors"
  "fmt"
  "io"
  "strconv"
  "strings"
  
  "github.com/adriagipas/imgcp/cdread"
)




/******/
/* CD */
/******/

// Segueix una aproximació lazzy
type _CD struct {

  file_name string
  cd        cdread.CD
  
}


func newCD( file_name string ) (*_CD,error) {

  ret:= _CD{
    file_name : file_name,
  }
  var err error
  if ret.cd,err= cdread.Open ( file_name ); err != nil {
    return nil,err
  }

  return &ret,nil
  
} // newCD


func (self *_CD) PrintInfo( file io.Writer, prefix string ) error {

  // Preparació impressió
  P:= fmt.Fprintln
  F:= fmt.Fprintf

  // Obté informació
  info:= self.cd.Info ()
  
  // Imprimeix
  F(file,"%sCD-Rom Image (%s)\n", prefix, self.cd.Format () )
  P(file,prefix,"")
  P(file,prefix, "Sessions:")
  
  var sess *cdread.SessionInfo
  var track *cdread.TrackInfo
  for s:= 0; s < len(info.Sessions); s++ {
    sess= &info.Sessions[s]
    P(file,"")
    F(file,"%s  %d) Tracks:\n",prefix,s)
    P(file,"")
    for t:= 0; t < len(sess.Tracks); t++ {
      track= &sess.Tracks[t]

      // Identificador
      F(file,"%s    Id: %02X",prefix,track.Id)

      // Posició inicial
      // --> Busca índex inicial
      var i int
      for i= 0; i < len(track.Indexes) && track.Indexes[i].Id != 1; i++ {
      }
      if i==len(track.Indexes) {
        F(file, "  Start: ??:??:??")
      } else {
        F(file, "  Start: %02x:%02x:%02x",
          track.Indexes[i].Pos.Minutes,
          track.Indexes[i].Pos.Seconds,
          track.Indexes[i].Pos.Sector )
      }
      
      // Tipus
      F(file,"  Type: ")
      switch track.Type {
      case cdread.TRACK_TYPE_AUDIO:
        F(file,"Audio")
      case cdread.TRACK_TYPE_MODE1_RAW:
        F(file,"Mode1 (Raw sectors)")
      case cdread.TRACK_TYPE_MODE2_RAW:
        F(file,"Mode2 (Raw sectors)")
      case cdread.TRACK_TYPE_MODE2_CDXA_RAW:
        F(file,"CD-XA/Mode2 (Raw sectors)")
      case cdread.TRACK_TYPE_ISO:
        F(file,"Data")
      default:
        F(file,"Unknown")
      }

      // Salt de línia
      P(file,"")

      // Si és ISO imprimeix la info
      if track.Type != cdread.TRACK_TYPE_AUDIO &&
        track.Type != cdread.TRACK_TYPE_UNK {
        if iso,err:= newISO_9660 ( self.cd, s, t ); err == nil {
          P(file,"")
          iso.PrintInfo ( file, prefix+"        " )
        }
        
      }
      
    }
  }
  
  return nil
  
} // end PrintInfo


func (self *_CD) GetRootDirectory() (Directory,error) {

  info:= self.cd.Info ()
  var ret Directory
  if len(info.Sessions) == 1 {
    ret= &_CD_TracksDir{
      cd : self.cd,
      cd_info : info,
      sess : 0,
    }
  } else {
    ret= &_CD_SessionsDir{
      cd : self.cd,
      cd_info : info,
    }
  }
  
  return ret,nil
  
} // end GetRootDirectory




/****************/
/* SESSIONS DIR */
/****************/

type _CD_SessionsDir struct {
  
  cd      cdread.CD
  cd_info *cdread.Info
  
}


func (self *_CD_SessionsDir) Begin() (DirectoryIter,error) {
  
  ret:= _CD_SessionsDirIter{
    dir : self,
    current_sess : 0,
  }

  return &ret,nil
  
} // end Begin


func (self *_CD_SessionsDir) MakeDir( name string ) (Directory,error) {
  return nil,errors.New ( "Make directory not implemented for CD images" )
} // end MakeDir


func (self *_CD_SessionsDir) GetFileWriter(name string) (FileWriter,error) {
  return nil,errors.New ( "Writing a file not implemented for CD images" )
} // end GetFileWriter


type _CD_SessionsDirIter struct {

  dir          *_CD_SessionsDir
  current_sess int
  
}


func (self *_CD_SessionsDirIter) CompareToName(name string) bool {
  return strings.ToLower ( name ) == self.GetName ()
} // end CompareToName


func (self *_CD_SessionsDirIter) End() bool {
  return self.current_sess >= len(self.dir.cd_info.Sessions)
} // end End


func (self *_CD_SessionsDirIter) GetDirectory() (Directory,error) {
  
  ret:= _CD_TracksDir{
    cd : self.dir.cd,
    cd_info : self.dir.cd_info,
    sess : self.current_sess,
  }
  
  return &ret,nil
  
} // end GetDirectory


func (self *_CD_SessionsDirIter) GetFileReader() (FileReader,error) {
  return nil,errors.New ( "_CD_SessionsDirIter.GetFileReader: WTF!!" )
} // end GetFileReader


func (self *_CD_SessionsDirIter) GetName() string {
  return strconv.FormatInt ( int64(self.current_sess), 10 )
} // end GetName


func (self *_CD_SessionsDirIter) List( file io.Writer ) error {
  
  fmt.Fprintf ( file, "session  %d\n", self.current_sess )

  return nil
  
} // end List


func (self *_CD_SessionsDirIter) Next() error {
  
  self.current_sess++
  
  return nil
  
} // end Next


func (self *_CD_SessionsDirIter) Remove() error {
  return errors.New ( "Remove file not implemented for CD images" )
} // end Remove


func (self *_CD_SessionsDirIter) Type() int {
  return DIRECTORY_ITER_TYPE_DIR
} // end Type




/**************/
/* TRACKS DIR */
/**************/

type _CD_TracksDir struct {
  
  cd      cdread.CD
  cd_info *cdread.Info
  sess    int
  
}


func (self *_CD_TracksDir) Begin() (DirectoryIter,error) {
  return newCDTracksDirIter( self )
} // end Begin


func (self *_CD_TracksDir) MakeDir( name string ) (Directory,error) {
  return nil,errors.New ( "Make directory not implemented for CD images" )
} // end MakeDir


func (self *_CD_TracksDir) GetFileWriter(name string) (FileWriter,error) {
  return nil,errors.New ( "Make directory not implemented for CD images" )
} // end GetFileWriter


type _CD_TracksDirIter struct {

  dir          *_CD_TracksDir
  current_track int
  is_iso        bool
  
}


func (self *_CD_TracksDirIter) checkIsIso() error {

  self.is_iso= false
  
  // Si estem al final no faces res.
  if self.End() { return nil}
  
  ttype:= self.getTrackType ()
  if ttype == cdread.TRACK_TYPE_AUDIO || ttype == cdread.TRACK_TYPE_UNK {
    return nil
  }

  // Obté un reader
  tr,err:= self.dir.cd.TrackReader ( self.dir.sess, self.current_track, 0 )
  if err != nil { return err }
  defer tr.Close ()
  
  // Comprova signatura ISO
  var data [6]byte
  if err:= tr.Seek ( 0x10 ); err == nil { // Podria ser que fora més menut
    if _,err:= tr.Read ( data[:] ); err != nil { return err }
    if data[1]=='C' && data[2]=='D' && data[3]=='0' &&
      data[4]=='0' && data[5]=='1' {
      self.is_iso= true
    }
  }

  return nil
  
} // end checkIsIso


func (self *_CD_TracksDirIter) getTrackType() int {
    return self.dir.cd_info.Sessions[self.dir.sess].
      Tracks[self.current_track].Type
} // end getTrackType


func newCDTracksDirIter( dir *_CD_TracksDir ) (*_CD_TracksDirIter,error) {

  // Inicialitza
  ret:= _CD_TracksDirIter{
    dir : dir,
    current_track : 0,
    is_iso : false,
  }

  // Comprova si és ISO
  if err:= ret.checkIsIso (); err != nil {
    return nil,err
  }
  
  return &ret,nil
  
} // end newCDTracksDirIter


func (self *_CD_TracksDirIter) CompareToName(name string) bool {
  return strings.ToLower ( name ) == self.GetName ()
} // end CompareToName


func (self *_CD_TracksDirIter) End() bool {

  ntracks:= len(self.dir.cd_info.Sessions[self.dir.sess].Tracks)
  
  return self.current_track >= ntracks
  
} // end End


func (self *_CD_TracksDirIter) GetDirectory() (Directory,error) {

  iso,err:= newISO_9660 ( self.dir.cd, self.dir.sess, self.current_track )
  if err != nil { return nil,err }

  return iso.GetRootDirectory ( )
  
} // end GetDirectory


func (self *_CD_TracksDirIter) GetFileReader() (FileReader,error) {
  
  ttype:= self.getTrackType ()
  if ttype == cdread.TRACK_TYPE_AUDIO {
    f,err:= self.dir.cd.TrackReader ( self.dir.sess, self.current_track, 0 )
    if err != nil { return nil,err }
    return newCDWavReader ( f )
  } else {
    return self.dir.cd.TrackReader ( self.dir.sess, self.current_track, 0 )
  }
  
} // end GetFileReader


func (self *_CD_TracksDirIter) GetName() string {

  ttype:= self.getTrackType ()
  var ret string
  if ttype == cdread.TRACK_TYPE_AUDIO {
    ret= fmt.Sprintf ( "%d.wav", self.current_track )
  } else if self.is_iso {
    ret= strconv.FormatInt ( int64(self.current_track), 10 )
  } else {
    ret= fmt.Sprintf ( "%d.dat", self.current_track )
  }

  return ret
  
} // end GetName


func (self *_CD_TracksDirIter) List( file io.Writer ) error {

  P:= func(args... any) {
    fmt.Fprint ( file, args... )
  }
  F:= func(format string,args... any) {
    fmt.Fprintf ( file, format, args... )
  }

  // És o no directori
  it_type:= self.Type ()
  if it_type==DIRECTORY_ITER_TYPE_DIR { P("d") } else { P("-") }

  P("  ")

  // Posició inicial
  var i int
  track:= &self.dir.cd_info.Sessions[self.dir.sess].Tracks[self.current_track]
  for i= 0; i < len(track.Indexes) && track.Indexes[i].Id != 1; i++ {
  }
  if i==len(track.Indexes) {
    P("??:??:??")
  } else {
    F("%02x:%02x:%02x",
      track.Indexes[i].Pos.Minutes,
      track.Indexes[i].Pos.Seconds,
      track.Indexes[i].Pos.Sector )
  }

  P("  ")

  // Tipus
  ttype:= self.getTrackType ()
  if ttype == cdread.TRACK_TYPE_AUDIO {
    P("[AUDIO]")
  } else if ttype == cdread.TRACK_TYPE_UNK {
    P("[?????]")
  } else if self.is_iso {
    P("[ISO  ]")
  } else {
    P("[DATA ]")
  }

  P("  ")

  // Nom
  P(self.GetName ())

  P("\n")

  return nil
  
} // end List


func (self *_CD_TracksDirIter) Next() error {
  
  self.current_track++
  if err:= self.checkIsIso (); err != nil {
    return err
  }
  
  return nil
  
} // end Next


func (self *_CD_TracksDirIter) Remove() error {
  return errors.New ( "Remove file not implemented for CD images" )
} // end Remove


func (self *_CD_TracksDirIter) Type() int {
  
  if self.is_iso {
    return DIRECTORY_ITER_TYPE_DIR
  } else {
    return DIRECTORY_ITER_TYPE_FILE
  }
  
} // end Type




/**************/
/* WAV READER */
/**************/

type _CD_WavReader struct {

  f          cdread.TrackReader
  header     [44]byte
  header_pos int
  
}


func newCDWavReader( f cdread.TrackReader ) (*_CD_WavReader,error) {

  // Inicialitza
  ret:= _CD_WavReader{
    f : f,
    header_pos : 0,
  }

  // Calcula grandària i rebobina.
  var size int64= 0
  var nread int
  var err error= nil
  var buf [2048]byte
  if err= f.Seek ( 0 ); err != nil { return nil,err }
  for err == nil {
    nread,err= f.Read ( buf[:] )
    size+= int64(nread)
  }
  if err != io.EOF { return nil,err }
  if err= f.Seek ( 0 ); err != nil { return nil,err }

  // Inicialitza capçalera
  h:= ret.header[:]
  // --> RIFF
  h[0]= 'R'; h[1]= 'I'; h[2]= 'F'; h[3]= 'F'
  // --> Grandària total
  total:= size+44-8
  h[4]= byte(uint8(total&0xff))
  h[5]= byte(uint8((total>>8)&0xff))
  h[6]= byte(uint8((total>>16)&0xff))
  h[7]= byte(uint8((total>>24)&0xff))
  // --> WAVE
  h[8]= 'W'; h[9]= 'A'; h[10]= 'V'; h[11]= 'E'
  // --> fmt
  h[12]= 'f'; h[13]= 'm'; h[14]= 't'; h[15]= ' '
  // --> Length of format data as listed above
  h[16]= 16; h[17]= 0; h[18]= 0; h[19]= 0
  // --> Type of format (1 is PCM) - 2 byte integer
  h[20]= 1; h[21]= 0
  // --> Number of Channels - 2 byte integer
  h[22]= 2; h[23]= 0
  // --> Sample Rate
  srate:= 44100
  h[24]= byte(uint8(srate&0xff))
  h[25]= byte(uint8((srate>>8)&0xff))
  h[26]= byte(uint8((srate>>16)&0xff))
  h[27]= byte(uint8((srate>>24)&0xff))
  // --> (Sample Rate * BitsPerSample * Channels) / 8
  tmp1:= 176400
  h[28]= byte(uint8(tmp1&0xff))
  h[29]= byte(uint8((tmp1>>8)&0xff))
  h[30]= byte(uint8((tmp1>>16)&0xff))
  h[31]= byte(uint8((tmp1>>24)&0xff))
  // --> (BitsPerSample * Channels)
  h[32]= 4; h[33]= 0
  // --> Bits per sample
  h[34]= 16; h[35]= 0
  // --> data
  h[36]= 'd'; h[37]= 'a'; h[38]= 't'; h[39]= 'a'
  // --> Data size
  h[40]= byte(uint8(size&0xff))
  h[41]= byte(uint8((size>>8)&0xff))
  h[42]= byte(uint8((size>>16)&0xff))
  h[43]= byte(uint8((size>>24)&0xff))

  return &ret,nil
  
} // end _CD_WavReader


func (self *_CD_WavReader) Read( buf []byte ) (int,error) {

  // Prepara
  if len(buf) == 0 { return 0,nil }
  var ret= 0

  // Llig capçalera
  if self.header_pos<len(self.header) {
    
    // Bytes a llegir
    remain_header:= len(self.header)-self.header_pos
    var nbytes int
    if len(buf) > remain_header {
      nbytes= remain_header
    } else {
      nbytes= len(buf)
    }

    // Copia i actualitza
    copy(buf[:nbytes],self.header[self.header_pos:self.header_pos+nbytes])
    buf= buf[nbytes:]
    self.header_pos+= nbytes
    ret+= nbytes
    
  }

  // Llig del track
  if len(buf)>0 {
    nread,err:= self.f.Read ( buf )
    if err != nil { return ret,err }
    ret+= nread
  }
  
  return ret,nil
  
} // end Read


func (self *_CD_WavReader) Close() error {
  return self.f.Close ()
} // end Close
