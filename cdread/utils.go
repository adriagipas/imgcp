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
 *  utils.go - Utilitats.
 */

package cdread

import (
  "io"
)

func BCD( num int ) uint8 {
  return uint8((num/10)*0x10 + num%10)
} // end BCD


// Aquesta funció comprova si un track en mode2 és CDXA o no. No té
// sentit cridar a questa funció si el track no és mode2.
func CheckTrackIsMode2CDXA( tr TrackReader ) (bool,error) {

  // Prepara
  var buf [2336]byte
  if err:= tr.Seek ( 0x10 ); err != nil {
    return false,err
  }
  
  // Localitza el primary volum descriptor
  var vd_type uint8= 0 // Busque el 1 o 0xff
  var data []byte
  for vd_type != 1 && vd_type != 0xff {

    // Intenta llegir
    if n,err:= tr.Read ( buf[:] ); err != nil && err != io.EOF {
      return false,err
    } else if n != len(buf) {
      return false,nil
    }

    // Comprova que és sector de dades (FORM1) i selecciona dades
    if uint8(buf[2]) != uint8(buf[6]) || (uint8(buf[2])&0x20)!=0 {
      return false,nil
    }
    data= buf[8:0x808]

    // Comprova que és un Volume Descriptor
    if data[1]!='C' || data[2]!='D' || data[3]!='0' ||
      data[4]!='0' || data[5]!='1' {
      return false,nil
    }

    // Tipus
    vd_type= uint8(data[0])
    
  }
  // No s'ha trobat un volum primari.
  if vd_type == 0xff { return false,nil }

  // Comprova signatura
  s:= data[0x400:0x408]
  if s[0]=='C' && s[1]=='D' && s[2]=='-' && s[3]=='X' &&
    s[4]=='A' && s[5]=='0' && s[6]=='0' && s[7]=='1' {
    return true,nil
  } else {
    return false,nil
  }
  
} // end CheckTrackIsMode2CDXA


// Tradueix un índex de sector en una estructura de tipus Position.
func GetPosition( sec_ind int64 ) Position {

  // Obté minuts, segons i sectors
  mm:= sec_ind/(60*75)
  tmp:= sec_ind%(60*75)
  ss:= tmp/75
  sec:= tmp%75

  // Passa a BCD
  ret:= Position{
    Minutes : BCD ( int(mm) ),
    Seconds : BCD ( int(ss) ),
    Sector  : BCD ( int(sec) ),
  }

  return ret
  
} // end GetPosition
