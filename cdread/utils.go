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


func BCD( num int ) uint8 {
  return uint8((num/10)*0x10 + num%10)
} // end BCD


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
