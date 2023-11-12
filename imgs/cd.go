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

      // NOTA!!! TODO!!! Faltaria provar a imprimir informació ISO.
      
    }
  }
  
  return nil
  
} // end PrintInfo


func (self *_CD) GetRootDirectory() (Directory,error) {
  return nil,errors.New("TODO - CD.GetRootDirectory")
} // end GetRootDirectory
