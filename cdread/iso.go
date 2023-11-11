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
 *  iso.go - Format imatge ISO.
 */

package cdread

import (
  "fmt"
  "os"
)




/*************/
/* CONSTANTS */
/*************/

const _ISO_SECTOR_SIZE = 2048




/****************/
/* TRACK READER */
/****************/

type _Iso_TrackReader struct {
  *os.File
}


func (self *_Iso_TrackReader) Seek( sector int64 ) error {

  offset:= sector*_ISO_SECTOR_SIZE
  n,err:= self.File.Seek ( offset, 0 )
  if err != nil { return err }
  if n != offset {
    return fmt.Errorf ( "unable to move to sector (%d)", sector )
  }
  
  return nil
  
} // end Seek




/******/
/* CD */
/******/

type _CD_Iso struct {

  file_name   string
  num_sectors int64
  
}


func (self *_CD_Iso) Info() *Info {

  // Inicialitza.
  ret:= Info{}
  ret.Sessions= make([]SessionInfo,1)
  tracks:= make([]TrackInfo,1)
  indexes:= make([]IndexInfo,1)


  // Sessions
  ret.Sessions[0].Tracks= tracks

  // Tracks i indexes
  ret.Tracks= tracks
  tracks[0].Type= TRACK_TYPE_ISO
  tracks[0].Id= BCD ( 1 )
  tracks[0].Indexes= indexes
  indexes[0].Pos= GetPosition ( 0 )
  tracks[0].PosLastSector= GetPosition ( self.num_sectors - 1 )
  
  return &ret
  
} // end Info


func (self *_CD_Iso) TrackReader(

  session_id int,
  track_id   int,
  mode       int,

) (TrackReader,error) {

  // Selecciona sessió
  if session_id != 0 {
    return nil,fmt.Errorf ( "session (%d) out of range", session_id )
  }

  // Selecciona track
  if track_id != 0 {
    return nil,fmt.Errorf ( "track (%d) out of range", track_id )
  }

  // Crea TrackReader
  var err error
  ret:= _Iso_TrackReader{}
  if ret.File,err= os.Open ( self.file_name ); err != nil {
    return nil,err
  }
  
  return &ret,nil
  
} // end TrackReader




/**********************/
/* FUNCIONS PÚBLIQUES */
/**********************/

func OpenIso( file_name string ) (CD,error) {

  // Intenta obrir el fitxer
  f,err:= os.Open ( file_name )
  if err != nil { return nil,err }
  defer f.Close ()

  // Obté nombre de sectors.
  info,err:= f.Stat ()
  if err != nil { return nil,err }
  if info.Size()%_ISO_SECTOR_SIZE != 0 || info.Size()<17*_ISO_SECTOR_SIZE {
    return nil,fmt.Errorf ( "'%s' size (%d) is not a valid size for a ISO file",
      file_name, info.Size () )
  }

  // Llig la signatura del primer decriptor de volum.
  var data [5]byte
  if _,err:= f.Seek ( 16*_ISO_SECTOR_SIZE + 1, 0 ); err != nil {
    return nil,err
  }
  nread,err:= f.Read ( data[:] )
  if err != nil { return nil,err }

  // Comprova
  if nread != 5 || data[0]!='C' || data[1]!='D' || data[2]!='0' ||
    data[3]!='0' || data[4]!='1' {
    return nil,fmt.Errorf ( "'%s' is not a ISO file" )
  }

  // Crea CD
  ret:= _CD_Iso{
    file_name   : file_name,
    num_sectors : info.Size()/_ISO_SECTOR_SIZE,
  }

  return &ret,nil
  
} // end OpenIso
