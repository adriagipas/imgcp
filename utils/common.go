/*
 * Copyright 2022 Adrià Giménez Pastor.
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
 *  common.go - Funcions bàsiques.
 *
 */

package utils;

import (
  "errors"
  "fmt"
  "os"
  "strconv"
)

/************/
/* FUNCIONS */
/************/

func NumBytesToStr(num_bytes uint64) string {
  if num_bytes > 1024*1024*1024 { // G
    val := float64(num_bytes)/(1024*1024*1024)
    return strconv.FormatFloat ( val, 'f', 1, 32 ) + "G"
  } else if num_bytes > 1024*1024 { // M
    val := float64(num_bytes)/(1024*1024)
    return strconv.FormatFloat ( val, 'f', 1, 32 ) + "M"
  } else if num_bytes > 1024 { // K
    val := float64(num_bytes)/1024
    return strconv.FormatFloat ( val, 'f', 1, 32 ) + "K"
  } else {
    return strconv.FormatUint ( num_bytes, 10 )
  }
} // end NumBytesToStr


// Llig bytes d'un fitxer fent comprovacions
func ReadBytes(

  f        *os.File,
  f_begin  int64,
  f_length int64,
  buf      []byte,
  offset   int64,
  
) error {

  length := int64(len(buf))
  
  end := f_begin + f_length
  if offset < f_begin || offset >= end {
    return fmt.Errorf ( "error while reading bytes: offset (%d) is out"+
      " of bounds (offset:%d, length:%d)",
      offset, f_begin, f_length )
  }
  my_end := offset + length
  if my_end > end {
    return fmt.Errorf ( "error while reading bytes: segment "+
      "(offset:%d, length:%d) is out of bounds (offset:%d, length:%d)",
    offset, length, f_begin, f_length )
  }

  // Llig bytes
  nbytes,err := f.ReadAt ( buf, offset )
  if err != nil { return err }
  if nbytes != len(buf) {
    return errors.New("Unexpected error occurred while reading bytes")
  }
  
  return nil
  
} // ReadBytes


// Llig bytes d'un fitxer fent comprovacions
func WriteBytes(

  f        *os.File,
  f_begin  int64,
  f_length int64,
  buf      []byte,
  offset   int64,
  
) error {

  length := int64(len(buf))
  
  end := f_begin + f_length
  if offset < f_begin || offset >= end {
    return fmt.Errorf ( "error while writing bytes: offset (%d) is out"+
      " of bounds (offset:%d, length:%d)",
      offset, f_begin, f_length )
  }
  my_end := offset + length
  if my_end > end {
    return fmt.Errorf ( "error while writing bytes: segment "+
      "(offset:%d, length:%d) is out of bounds (offset:%d, length:%d)",
    offset, length, f_begin, f_length )
  }

  // Escriu bytes
  nbytes,err := f.WriteAt ( buf, offset )
  if err != nil { return err }
  if nbytes != len(buf) {
    return errors.New("Unexpected error occurred while writing bytes")
  }
  
  return nil
  
} // WriteBytes


func Warning(format string, args ...any) {
  fmt.Fprintf ( os.Stderr, "[WW] " )
  fmt.Fprintf ( os.Stderr, format, args... )
  fmt.Fprintf ( os.Stderr, "\n" )
}
