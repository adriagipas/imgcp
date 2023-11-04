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

const (
  CD_CUE_TRACK_TYPE_AUDIO = 0
  CD_CUE_TRACK_TYPE_MODE1 = 1
  CD_CUE_TRACK_TYPE_MODE2 = 2
)

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
  index_id int
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

  // Posició actual
  current_sec int64

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
    track.track_type= CD_CUE_TRACK_TYPE_AUDIO
  case "MODE1/2352":
    track.track_type= CD_CUE_TRACK_TYPE_MODE1
  case "MODE2/2352":
    track.track_type= CD_CUE_TRACK_TYPE_MODE2
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
  fmt.Println ( "Gap", gap )
  fmt.Println ( "BinSize", bin_size )
  
  return errors.New ( "TODO - readMapSectors !!!!" )
  
} // end createMapSectors


func (self *_CD_Cue) Info() *Info {
  fmt.Println("TODO - _CD_Cue.Info!!!!")
  return nil
} // end Info


func (self *_CD_Cue) Reader() (Reader,error) {
  return nil,errors.New("TODO - _CD_Cue.Reader!!!!")
} // end Reader




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
  
  return &ret,nil
  
} // end OpenCue