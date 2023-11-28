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
  "os"
)




/******/
/* CD */
/******/

type _CD_Mds struct {

  file_name string
  
}


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
  
  fmt.Println ( buf )
  fmt.Println ( buf[0x12:0x16])
  
  return errors.New ( "TODO - Mds.init" )
  
} // end init


func (self *_CD_Mds) Format() string { return "MDS/MDF (Alcohol 120%)" }


func (self *_CD_Mds) Info() *Info {
  fmt.Println("TODO Mds.Info")
  return nil
}


func (self *_CD_Mds) TrackReader(
  
  session_id int,
  track_id   int,
  mode       int,
  
) (TrackReader,error) {
  return nil,errors.New("TODO Mds.TrackReader")
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
