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
 *  cue.go - Format CUE/BIN.
 */

package cdread

import (
  "bufio"
  "errors"
  "fmt"
  "io"
  "os"
  "path"
  "strconv"
  "strings"
  "unicode"
)




/*********/
/* UTILS */
/*********/

func newCueTokenizer( line string ) *bufio.Scanner {

  // Funció que processa text
  split_func:= func(
    
    data  []byte,
    atEOF bool,
    
  ) (advance int, token []byte, err error) {

    if atEOF && len(data) == 0 {
      return 0,nil,nil
    }
    
    // Ignora espais
    pos:= 0
    for pos= 0; pos < len(data) && unicode.IsSpace ( rune(data[pos]) ); pos++ {
    }
    if pos == len(data) {
      return pos,nil,nil
    }
    
    // Processa String
    if data[pos] == '"' {
      i:= pos + 1
      for ; i < len(data) && data[i]!='"'; i++ {
      }
      if i == len(data) { // No s'ha trobat
        if atEOF {
          advance,token,err= 0,nil,fmt.Errorf ( "string not closed: [%s]",
            string(data[pos:]) )
        } else {
          advance,token,err= pos,nil,nil // Demana més dades
        }
      } else if i == pos+1 { // String buit
        advance,token,err= 0,nil,fmt.Errorf ( "empty string" )
      } else { // Torna string
        advance,token,err= i+1,data[pos+1:i],nil
      }

      // Token normal
    } else { 
      i:= pos + 1
      for ; i < len(data) && !unicode.IsSpace ( rune(data[i]) ); i++ {
      }
      if i==len(data) { // Hem parat sense espai
        if atEOF { // És un token
          advance,token,err= len(data),data[pos:],nil
        } else { // Demanem més per si de cas
          advance,token,err= pos,nil,nil
        }
      } else { // Hem trobat espai
        advance,token,err= i,data[pos:i],nil
      }
    }
    
    return
    
  } // end split_func
  
  // Crea scanner
  ret:= bufio.NewScanner ( strings.NewReader ( line ) )
  ret.Split ( split_func )
  
  return ret
  
} // end newCueTokenizer


func processTimeCue( str string ) (int64,error) {

  if len(str) != 8 ||
    str[0] < '0' || str[0] > '9' ||
    str[1] < '0' || str[1] > '9' ||
    str[2] != ':' ||
    str[3] < '0' || str[3] > '9' ||
    str[4] < '0' || str[4] > '9' ||
    str[5] != ':' ||
    str[6] < '0' || str[6] > '9' ||
    str[7] < '0' || str[7] > '9' {
    return -1,fmt.Errorf ( "wrong time format: %s", str )
  }

  // Transforma.
  var ret int64= 0
  if tmp,err:= strconv.ParseInt ( str[:2], 10, 64 ); err != nil {
    return -1,err
  } else {
    ret+= tmp*60*75 // minuts
  }
  if tmp,err:= strconv.ParseInt ( str[3:5], 10, 64 ); err != nil {
    return -1,err
  } else {
    ret+= tmp*75 // segons
  }
  if tmp,err:= strconv.ParseInt ( str[6:], 10, 64 ); err != nil {
    return -1,err
  } else {
    ret+= tmp // sectors
  }

  return ret,nil
  
} // end processTimeCue




/****************/
/* TRACK READER */
/****************/

type _Cue_TrackReader struct {

  mode     int
  cd       *_CD_Cue
  track    *_CD_Cue_Track
  track_id int
  
  // Situació sectors
  next_sector int64
  eof         bool
  
  // Sector actual.
  // NOTA!!! SECTOR_SIZE té trellat en modes RAW.
  sec_data  [SECTOR_SIZE]byte
  data      []byte // Slice de sec_data
  data_size int
  pos       int
  bin_file  *_CD_Cue_BinFile
  file      *os.File
  
}

// Intenta carregar el següent sector. Si no es pot perquè s'ha
// aplegat al final es fica EOF a true.
func (self *_Cue_TrackReader) loadNextSector() error {
  
  // Eof
  if self.eof ||
    self.next_sector >= int64(len(self.cd.maps)) ||
    self.cd.maps[self.next_sector].track_id != self.track_id {
    self.eof= true
    return nil
  }

  // Prepara fitxer.
  bin_file:= self.cd.maps[self.next_sector].file
  if bin_file != self.bin_file {
    
    self.bin_file= bin_file

    // Tanca si tenim algun fitxer obert
    if self.file != nil {
      if err:= self.file.Close (); err != nil { return err }
    }
    
    // Obri nou fitxer
    var err error
    self.file,err= os.Open ( bin_file.file_name )
    if err != nil { return err }
    
  }

  // Mou a la posició del sector
  new_off,err:= self.file.Seek ( self.cd.maps[self.next_sector].offset, 0 )
  if err != nil { return err }
  if new_off != self.cd.maps[self.next_sector].offset {
    return fmt.Errorf ( "failed to read sector %d", self.next_sector )
  }

  // Llig el sector.
  n,err:= self.file.Read ( self.sec_data[:] )
  if err != nil { return err }
  if n != SECTOR_SIZE {
    return fmt.Errorf ( "failed to read sector %d", self.next_sector )
  }

  // Actualitza estat.
  self.pos= 0
  switch self.track.track_type {
  case TRACK_TYPE_AUDIO:
    self.data= self.sec_data[:]
    self.data_size= SECTOR_SIZE
  case TRACK_TYPE_MODE1_RAW:
    self.data= self.sec_data[16:2064]
    self.data_size= 2048
  case TRACK_TYPE_MODE2_RAW:
    self.data= self.sec_data[16:]
    self.data_size= 2336
  case TRACK_TYPE_MODE2_CDXA_RAW:
    if self.sec_data[0x12]&0x20 == 0 { // Form1
      self.data= self.sec_data[0x18:0x818]
      self.data_size= 2048
      if self.mode == MODE_CDXA_MEDIA_ONLY {
        self.pos= self.data_size+1
      }
    } else { // Form2
      self.data= self.sec_data[0x18:0x18+2324]
      self.data_size= 2324
      if self.mode == MODE_DATA {
        self.pos= self.data_size+1
      }
    }
  default:
    return fmt.Errorf ( "load sectors of type %d not implemented",
      self.track.track_type )
  }
  self.next_sector++

  return nil
  
} // end loadNextSector


func (self *_Cue_TrackReader) Close() (err error) {

  if self.file != nil {
    err= self.file.Close ()
  } else {
    err= nil
  }

  return
  
} // end Close


func (self *_Cue_TrackReader) Read( b []byte ) (n int,err error) {

  // EOF
  if self.eof { return 0,io.EOF }

  // Llig
  pos,remain:= 0,len(b)
  for remain > 0 && !self.eof {

    // Recarrega si cal
    // ATENCIÓ!! Si els sectors no són d'aquesta grandària podria
    // fallar. Ara sols suporte RAW sectors.
    for self.pos >= self.data_size {
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


func (self *_Cue_TrackReader) Seek( sector int64 ) error {

  // Tanca fitxers
  if self.file != nil {
    if err:= self.file.Close (); err != nil {
      return err
    }
    self.file= nil
  }
  self.bin_file= nil

  // Actualitza estat
  self.eof= false
  self.pos= SECTOR_SIZE
  self.next_sector= self.track.sector_index01 + sector

  // Intenta carregar
  if err:= self.loadNextSector (); err != nil {
    return err
  }

  return nil
  
} // end Seek




/******/
/* CD */
/******/

type _CD_Cue_BinFile struct {
  file_name string
  size      int64 // Grandària en sectors
  asize     int64 // Sectors acumulats dels fitxers anteriors sense
                  // incloure l'actual.
  next      *_CD_Cue_BinFile
}

type _CD_Cue_Track struct {
  track_type     int
  p              int // Posició de la primera entrada en entries
  N              int // Nombre d'entrades
  sector_index01 int64 // Primer sector de l'índex 01.
}

const (
  CD_CUE_ENTRY_TYPE_PREGAP = 0
  CD_CUE_ENTRY_TYPE_INDEX  = 1
)

type _CD_Cue_Entry struct {
  entry_type int
  id         int // Identificador
  time       int64 // Posició del primer sector, en un pregap és el
                   // número de segments a ignorar. Aquest camp es
                   // reaprofita i acaba siguent el offset de cada
                   // entrada en valor absolut.
  file       *_CD_Cue_BinFile // Fitxer on està desada esta entrada.
}

type _CD_Cue_SectorMap struct {
  offset   int64
  track_id int
  index_id uint8
  file     *_CD_Cue_BinFile
  subq_ptr int
}

type _CD_Cue struct {
  
  // Fitxer
  file_name string
  
  // Binary
  files *_CD_Cue_BinFile
  
  // Cue
  tracks  []_CD_Cue_Track
  entries []_CD_Cue_Entry
  
  // Mapa sectors
  maps []_CD_Cue_SectorMap
  
  // Sectors subcanal_q erronis.
  // El format és literalment el del fitxer LSD:
  // 3B MIN SEC FRA || 12B QSUb (inclou CRC)
  subq [][]uint8
  
}


func (self *_CD_Cue) addBinary( file_name string ) error {

  // Intenta obrir
  f,err:= os.Open ( file_name )
  if err != nil { return err }
  info,err:= f.Stat ()
  if err != nil { return err }
  f.Close ()

  // Comprova grandària
  if info.Size ()%SECTOR_SIZE != 0 {
    return fmt.Errorf ( "binary file '%s' has a wrong size", file_name )
  }

  // Afegeix
  bf:= &_CD_Cue_BinFile{
    file_name : file_name,
    size      : info.Size ()/SECTOR_SIZE,
  }
  if self.files != nil {
    bf.next= self.files
    bf.asize= bf.next.asize + bf.next.size
  } 
  self.files= bf

  return nil
  
} // end addBinary


func (self *_CD_Cue) readFile( tok *bufio.Scanner ) error {

  // Obté nom del fitxer
  if !tok.Scan () {
    err:= tok.Err ()
    if err == nil {
      return errors.New ( "wrong file format: unable to read file name" )
    } else {
      return fmt.Errorf ( "wrong file format: %s", err )
    }
  }
  file_name:= tok.Text ()

  // Comprova que el tipus és BINARY
  if !tok.Scan () {
    err:= tok.Err ()
    if err == nil {
      return errors.New ( "wrong file format: unable to read token BINARY" )
    } else {
      return fmt.Errorf ( "wrong file format: %s", err )
    }
  }
  if aux:= tok.Text (); aux != "BINARY" {
    return fmt.Errorf ( "wrong file format: expected token BINARY, "+
      "instead '%s' was read", aux )
  }
  
  // Obté el nom del fitxer binary
  if !path.IsAbs ( file_name ) {
    dir_name:= path.Dir ( self.file_name )
    file_name= path.Join ( dir_name, file_name )
  }
  
  return self.addBinary ( file_name )
  
} // end readFile


func (self *_CD_Cue) readTrack( tok *bufio.Scanner ) error {
  
  // Llig track id
  if !tok.Scan () {
    err:= tok.Err ()
    if err == nil {
      return errors.New (
        "wrong track format: unable to read track identifier" )
    } else {
      return fmt.Errorf ( "wrong track format: %s", err )
    }
  }
  aux:= tok.Text ()
  _track_id,err:= strconv.ParseInt ( aux, 10, 32 )
  if err != nil {
    return fmt.Errorf ( "unable to parse track index (%s)", aux )
  }
  track_id:= int(_track_id)
  if track_id != len(self.tracks)+1 {
    return fmt.Errorf ( "expecting track %d, instead track %d was read",
      len(self.tracks)+1, track_id )
  }

  // Inicialitza track
  track:= _CD_Cue_Track{
    p : len(self.entries),
    N : 0,
  }

  // Obté mode
  if !tok.Scan () {
    err:= tok.Err ()
    if err == nil {
      return errors.New (
        "wrong track format: unable to read mode" )
    } else {
      return fmt.Errorf ( "wrong track format: %s", err )
    }
  }
  switch mode:= tok.Text (); mode {
  case "AUDIO":
    track.track_type= TRACK_TYPE_AUDIO
  case "MODE1/2352":
    track.track_type= TRACK_TYPE_MODE1_RAW
  case "MODE2/2352":
    track.track_type= TRACK_TYPE_MODE2_RAW
  default:
    return fmt.Errorf ( "TRACK format unknown: %s", mode )
  }
  
  // Afegeix.
  self.tracks= append(self.tracks,track)
  
  return nil
  
} // end readTrack


func (self *_CD_Cue) readIndex( tok *bufio.Scanner ) error {

  // Comprova que existeix fitxer.
  if self.files == nil {
    return errors.New ( "index defined before specifying a file" )
  }

  // Current track
  track:= &self.tracks[len(self.tracks)-1]

  // Get index id.
  if !tok.Scan () {
    err:= tok.Err ()
    if err == nil {
      return errors.New (
        "wrong index format: unable to read index identifier" )
    } else {
      return fmt.Errorf ( "wrong index format: %s", err )
    }
  }
  aux:= tok.Text ()
  _index_id,err:= strconv.ParseInt ( aux, 10, 32 )
  if err != nil {
    return fmt.Errorf ( "unable to parse index identifier (%s)", aux )
  }
  index_id:= int(_index_id)
  if index_id < 0 {
    return fmt.Errorf ( "wrong index identifier %d", index_id )
  }

  // Inicialitza entrada
  entry:= _CD_Cue_Entry{
    entry_type : CD_CUE_ENTRY_TYPE_INDEX,
    id         : index_id,
    file       : self.files,
  }

  // Obté time.
  if !tok.Scan () {
    err:= tok.Err ()
    if err == nil {
      return errors.New (
        "wrong index format: unable to read time" )
    } else {
      return fmt.Errorf ( "wrong index format: %s", err )
    }
  }
  entry.time,err= processTimeCue ( tok.Text () )
  if err != nil { return err }

  // Afegeix
  self.entries= append(self.entries,entry)
  track.N++
  
  return nil
  
} // end readIndex


func (self *_CD_Cue) readPregap( tok *bufio.Scanner ) error {

  // Prepara
  track:= &self.tracks[len(self.tracks)-1]
  entry:= _CD_Cue_Entry{
    entry_type : CD_CUE_ENTRY_TYPE_PREGAP,
    id         : -1,
    file       : nil,
  }

  // Obté time.
  if !tok.Scan () {
    err:= tok.Err ()
    if err == nil {
      return errors.New (
        "wrong index format: unable to read time" )
    } else {
      return fmt.Errorf ( "wrong index format: %s", err )
    }
  }
  var err error= nil
  entry.time,err= processTimeCue ( tok.Text () )
  if err != nil { return err }

  // Afegeix
  self.entries= append(self.entries,entry)
  track.N++
  
  return nil
  
} // end readPregap


func (self *_CD_Cue) readCommand( s *bufio.Scanner ) (cont bool,err error) {

  // Busca la primera línia no buida
  var line string
  var ok bool
  for ok= s.Scan (); ok; ok= s.Scan () {
    line= strings.TrimSpace ( s.Text () )
    if len(line) != 0 { break }
  }
  if !ok { return false,s.Err () }

  // Crea tokenizer i obté commandament
  cont,err= true,nil
  tok:= newCueTokenizer( line )
  if !tok.Scan () {
    return false,errors.New ( "wrong command format: command not found" )
  }
  switch cmd:= tok.Text (); cmd {
  case "FILE":
    err= self.readFile ( tok )
  case "TRACK":
    err= self.readTrack ( tok )
  case "INDEX":
    err= self.readIndex ( tok )
  case "PREGAP":
    err= self.readPregap ( tok )
  default:
    return false,fmt.Errorf ( "unknown command: %s", cmd )
  }
  
  return
  
} // end readContent


func readCDCue( cd *_CD_Cue, f *os.File ) error {

  var cont bool
  var err error
  s:= bufio.NewScanner ( f )
  for cont,err= true,nil; cont && err == nil; cont,err= cd.readCommand ( s ) {
  }

  return err
  
} // readCDCue


func (self *_CD_Cue) calcTotalGap() int64 {

  var gap int64= 2*75 // 2 segons inicials
  for n:= 0; n < len(self.entries); n++ {
    if self.entries[n].entry_type == CD_CUE_ENTRY_TYPE_PREGAP {
      gap+= self.entries[n].time
    }
  }
  
  return gap
  
} // end calcTotalGap


func (self *_CD_Cue) calcFilesSize() int64 {

  var ret int64= 0
  for p:= self.files; p != nil; p= p.next {
    ret+= p.size
  }
  
  return ret
  
} // end calcFilesSize


func (self *_CD_Cue) checkIndexesInRange() error {

  for n:= 0; n < len(self.entries); n++ {
    if self.entries[n].entry_type == CD_CUE_ENTRY_TYPE_INDEX &&
      self.entries[n].time >= self.entries[n].file.size {
      return fmt.Errorf ( "index reference out of range %d (binary"+
        " file has %d sectors)", n, self.entries[n].file.size )
    }
  }

  return nil
  
} // end checkIndexesInRange


func (self *_CD_Cue) createMapSectors() error {

  // Obté número total de sectors mapejats i comprova.
  gap:= self.calcTotalGap ()
  bin_size:= self.calcFilesSize ()
  if err:= self.checkIndexesInRange (); err != nil {
    return err
  }

  // Reserva memòria
  self.maps= make([]_CD_Cue_SectorMap,gap+bin_size)

  // Ompli i reajusta 'entries[n].time'.
  // --> Index 00 Track1 (Pregap 2s)
  var n int64= 0
  for ; n < 2*75; n++ {
    self.maps[n].offset= -1
    self.maps[n].track_id= 0
    self.maps[n].index_id= 0x00
    self.maps[n].file= nil
    self.maps[n].subq_ptr= -1
  }
  // Tracks
  err:= errors.New ( "invalid PREGAP/INDEX commands" )
  gap= 2*75
  var offset int64= 0
  var track *_CD_Cue_Track
  var entry *_CD_Cue_Entry
  var prev_file *_CD_Cue_BinFile= nil
  var end int64
  for t:= 0; t < len(self.tracks); t++ {
    track= &self.tracks[t]
    prev_ind:= 0
    for e:= track.p; e != track.p+track.N; e++ {
      entry= &self.entries[e]
      switch entry.entry_type {
      case CD_CUE_ENTRY_TYPE_PREGAP: // Pregap
        // Calc end
        if e == len(self.entries)-1 ||
          self.entries[e+1].entry_type != CD_CUE_ENTRY_TYPE_INDEX {
          return err
        }
        gap+= entry.time
        end= self.entries[e+1].file.asize + self.entries[e+1].time + gap
        if end <= n { return err }
        // Recalculate time and fill
        entry.time= n
        for ; n != end; n++ {
          self.maps[n].offset= -1
          self.maps[n].track_id= t
          self.maps[n].index_id= 0x00
          self.maps[n].file= nil
          self.maps[n].subq_ptr= -1
        }
      case CD_CUE_ENTRY_TYPE_INDEX: // Índex
        // Comprovacions d'index, el 0 és opcional
        if (entry.id == 0 && prev_ind != 0) ||
          (entry.id > 0 && entry.id != prev_ind+1) {
          return err
        }
        if entry.id != 0 { prev_ind++ }
        // Fixa sector_index01
        if entry.id == 1 { track.sector_index01= n }
        // Inicialitza offset si canvia de fitxer.
        if entry.file != prev_file {
          offset= 0
          prev_file= entry.file
        }
        // Calc end
        if e == len(self.entries)-1 {
          end= int64(len(self.maps))
        } else if self.entries[e+1].entry_type == CD_CUE_ENTRY_TYPE_INDEX {
          end= self.entries[e+1].file.asize + self.entries[e+1].time + gap
        } else {
          if e+1 == len(self.entries)-1 ||
            self.entries[e+2].entry_type != CD_CUE_ENTRY_TYPE_INDEX {
            return err
          }
          end= self.entries[e+2].file.asize + self.entries[e+2].time + gap
        }
        if end <= n { return err }
        // Recalculate time and fill
        entry.time= n
        for ; n != end; n++ {
          self.maps[n].offset= offset
          self.maps[n].track_id= t
          self.maps[n].index_id= BCD ( entry.id )
          self.maps[n].file= entry.file
          self.maps[n].subq_ptr= -1
          offset+= SECTOR_SIZE
        }
      }
    }
  }

  return nil
  
} // end createMapSectors


func (self *_CD_Cue) relabelCDXATracks() error {

  var track *_CD_Cue_Track
  var tr TrackReader
  var err error
  for i:= 0; i < len(self.tracks); i++ {
    track= &self.tracks[i]
    if track.track_type == TRACK_TYPE_MODE2_RAW {
      tr,err= self.TrackReader ( 0, i, 0 )
      if err != nil { return err }
      defer tr.Close ()
      if is_cdxa,err:= CheckTrackIsMode2CDXA ( tr ); err != nil {
        return err
      } else if is_cdxa {
        track.track_type= TRACK_TYPE_MODE2_CDXA_RAW
      }
      tr.Close ()
    }
  }
  
  return nil
  
} // end relabelCDXATracks


func (self *_CD_Cue) Format() string { return "CUE/BIN" }

func (self *_CD_Cue) Info() *Info {

  // Inicialitza
  ret:= Info{}
  ret.Sessions= make([]SessionInfo,1)
  tracks:= make([]TrackInfo,len(self.tracks))
  indexes:= make([]IndexInfo,len(self.entries))

  // Sessions.
  ret.Sessions[0].Tracks= tracks

  // Tracks i entries
  var tp *_CD_Cue_Track
  var ep *_CD_Cue_Entry
  ret.Tracks= tracks
  for t:= 0; t < len(self.tracks); t++ {
    tp= &self.tracks[t]
    tracks[t].Type= tp.track_type
    tracks[t].Id= BCD ( t+1 )
    tracks[t].Indexes= indexes[tp.p:tp.p+tp.N]
    if t > 0 {
      tracks[t-1].PosLastSector= GetPosition ( self.entries[tp.p].time - 1 )
    }
    for e:= tp.p; e != tp.p+tp.N; e++ {
      ep= &self.entries[e]
      if ep.entry_type == CD_CUE_ENTRY_TYPE_INDEX {
        indexes[e].Id= BCD ( ep.id )
      } else {
        indexes[e].Id= 0
      }
      indexes[e].Pos= GetPosition ( ep.time )
    }
  }
  tracks[len(self.tracks)-1].PosLastSector=
    GetPosition ( int64(len(self.maps))-1 )

  return &ret
  
} // end Info


func (self *_CD_Cue) TrackReader(

  session_id int,
  track_id   int,
  mode       int,
  
) (TrackReader,error) {
  
  // Selecciona sessió
  if session_id != 0 {
    return nil,fmt.Errorf ( "session (%d) out of range", session_id )
  }
  
  // Selecciona track
  if track_id < 0 || track_id >= len(self.tracks) {
    return nil,fmt.Errorf ( "track (%d) out of range", track_id )
  }
  track:= &self.tracks[track_id]

  // Crea trackreader
  ret:= _Cue_TrackReader{
    mode        : mode,
    cd          : self,
    track       : track,
    track_id    : self.maps[track.sector_index01].track_id,
    next_sector : track.sector_index01,
    eof         : false,
    bin_file    : nil,
    file        : nil,
    pos         : SECTOR_SIZE,
  }
  if err:= ret.loadNextSector (); err != nil {
    return nil,err
  }
  
  return &ret,nil
  
} // end TrackReader




/**********************/
/* FUNCIONS PÚBLIQUES */
/**********************/

func OpenCue( file_name string ) (CD,error) {

  // Intenta obrir el fitxer
  f,err:= os.Open ( file_name )
  if err != nil { return nil,err }
  defer f.Close ()
  
  // Crea i llig
  ret:= _CD_Cue{
    file_name : file_name,
  }
  if err:= readCDCue ( &ret, f ); err != nil {
    return nil,err
  }

  // Crea el mapa de sectors
  if err:= ret.createMapSectors (); err != nil {
    return nil,err
  }

  // Identifica tracks CDXA.
  if err:= ret.relabelCDXATracks(); err != nil {
    return nil,err
  }
  
  return &ret,nil
  
} // end OpenCue
