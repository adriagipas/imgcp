/*
 * Copyright 2022-2023 Adrià Giménez Pastor.
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
 *  detect.go - Funció per a detectar el tipus d'una image.
 *
 */

package imgs

import (
  "fmt"
  "os"
)

/*********/
/* TIPUS */
/*********/

const TYPE_UNK          = 0
const TYPE_MBR          = 1
const TYPE_FAT12        = 2
const TYPE_FAT16        = 3
const TYPE_LOCAL_FOLDER = 4
const TYPE_IFF          = 5


/************/
/* FUNCIONS */
/************/

const HEADER_SIZE = 512

func Detect(file_name string) (int,error) {

  // Primer prova si és una carpeta local
  ret,err := detect_local_folder ( file_name )
  if err!=nil { return -1,err }
  if ret!=TYPE_UNK { return ret,nil }
  
  // Primer prova capçaleres 4 bytes.
  ret,err= detect_h4 ( file_name )
  if err!=nil { return -1,err }
  if ret!=TYPE_UNK { return ret,nil }
  
  // Prova fitxers de blocs 512
  return detect_h512 ( file_name )
  
} // end Detect


func detect_local_folder(file_name string) (int,error) {

  // Obté informació del fitxer.
  f,err := os.Open ( file_name )
  if err != nil { return -1,err }
  info,err := f.Stat ()
  if err != nil { return -1,err }

  // Comprova si és una carpeta
  if info.IsDir () {
    return TYPE_LOCAL_FOLDER,nil
  } else {
    return TYPE_UNK,nil
  }
  
} // end detect_local_folder


// Sols empra els primers 4 bytes per prendre la decisió. Torna
// TYPE_UNK si no es sap.
func detect_h4(file_name string) (int,error) {

  // Obté informació del fitxer
  f,err := os.Open ( file_name )
  if err != nil { return -1,err }
  info,err := f.Stat ()
  if err != nil { return -1,err }

  // Llig primers 4 bytes
  nbytes := info.Size ()
  if nbytes < 4 { return TYPE_UNK,nil }
  var mem [4]byte
  head := mem[:]
  n,err := f.Read ( head )
  if err != nil { return -1,err }
  if n != 4 {
    return -1,fmt.Errorf ( "Unexpected error while reading header from '%s'",
      file_name )
  }
  f.Close ()

  // Comprova
  if head[0]=='F' && head[1]=='O' && head[2]=='R' && head[3]=='M' {
    return TYPE_IFF,nil
  } else if head[0]=='C' && head[1]=='A' && head[2]=='T' && head[3]==' ' {
    return TYPE_IFF,nil
  } else if head[0]=='L' && head[1]=='I' && head[2]=='S' && head[3]=='T' {
    return TYPE_IFF,nil
  } else {
    return TYPE_UNK,nil
  }
  
} // end detect_h4


// Per a detectar el tipus es proven totes les posibiltats que tornen
// puntuacions. El que tinga més puntuació guanya.
func detect_h512(file_name string) (int,error) {

  // Obté informació del fitxer.
  f,err := os.Open ( file_name )
  if err != nil { return -1,err }
  info,err := f.Stat ()
  if err != nil { return -1,err }
  // --> Grandària
  nbytes := info.Size()
  if nbytes < HEADER_SIZE {
    return -1,fmt.Errorf ( "'%s' is too small: %d B", file_name, nbytes )
  }
  // --> Capçalera
  var mem [HEADER_SIZE]byte
  header := mem[:]
  n,err := f.Read ( header )
  if err != nil { return -1,err }
  if n != HEADER_SIZE {
    return -1,fmt.Errorf ( "Unexpected error while reading header from '%s'",
      file_name )
  }
  // --> Tanca
  f.Close ()

  // Concurs
  ret,points := TYPE_UNK,0

  // --> MBR
  if tmp := detect_MBR ( header, nbytes ); tmp > points {
    ret,points= TYPE_MBR,tmp
  }

  // --> FAT12
  if tmp := detect_FAT12 ( header, nbytes ); tmp > points {
    ret,points= TYPE_FAT12,tmp
  }

  // --> FAT16
  if tmp := detect_FAT16 ( header, nbytes ); tmp > points {
    ret,points= TYPE_FAT16,tmp
  }
  
  return ret,nil
  
} // end detect_h512


func detect_FAT12(header []byte, nbytes int64) int {

  ret := detect_FAT1216 ( header, nbytes )
  if ret == -1 { return -1 }

  // Calcula nombre de clusters approx
  sectors := uint16(header[0x13]) | (uint16(header[0x14])<<8)
  sectors_clu := uint16(header[0x13])
  if sectors_clu == 0 {
    return -1
  }
  clusters := sectors/sectors_clu
  if clusters >= 4085 {
    return -1
  } else { ret++ }

  // Identificador FAT (no sempre està)
  if header[0x36]=='F' && header[0x37]=='A' &&
    header[0x38]=='T' && header[0x39]=='1' &&
    header[0x3a]=='2' {
    ret+= 10
  }
  
  return ret
  
} // end detect_FAT12


func detect_FAT16(header []byte, nbytes int64) int {

  ret := detect_FAT1216 ( header, nbytes )
  if ret == -1 { return -1 }

  // Calcula nombre de clusters approx
  sectors := uint16(header[0x13]) | (uint16(header[0x14])<<8)
  sectors_clu := uint16(header[0x13])
  if sectors_clu == 0 {
    return -1
  }
  clusters := sectors/sectors_clu
  if clusters < 4085 || clusters >= 65525 {
    return -1
  } else { ret++ }

  // Identificador FAT (no sempre està)
  if header[0x36]=='F' && header[0x37]=='A' &&
    header[0x38]=='T' && header[0x39]=='1' &&
    header[0x3a]=='6' {
    ret+= 10
  }
  
  return ret
  
} // end detect_FAT16


func detect_FAT1216(header []byte, nbytes int64) int {
  
  ret := 0
  
  // Signature
  signature := uint16(header[0x1fe]) | (uint16(header[0x1ff])<<8)
  if signature != 0xaa55 {
    return -1
  } else { ret++ }

  // Signature 2
  if tmp := header[0x26]; tmp == 0x28 || tmp == 0x29 {
    ret++
  }

  // Grandària sector
  sec_size := int64(uint64(uint16(header[0xb]) | (uint16(header[0xc])<<8)))
  if sec_size == 0 || nbytes%sec_size != 0 {
    return -1
  } else { ret++ }

  // Sectors en el volum
  sectors := uint16(header[0x13]) | (uint16(header[0x14])<<8)
  if sectors == 0 {
    return -1
  } 
  sectors_bytes := int64(uint64(sectors))*sec_size
  if sectors_bytes != nbytes {
    return -1
  } else { ret++ }
  
  return ret
  
} // detect_FAT1216


func detect_MBR(header []byte, nbytes int64) int {
  
  ret := 0
  
  // Signature
  signature := uint16(header[0x1fe]) | (uint16(header[0x1ff])<<8)
  if signature != 0xaa55 {
    return -1
  } else { ret++ }

  // Grandària
  if nbytes%512 != 0 || nbytes == 512 {
    return -1
  } else { ret++ }

  // Grandària particions actives
  var sectors uint64 = 1
  if pos := 0x1be; (header[pos]&0x80) != 0 {
    sectors+= uint64(header[pos+0xc]) |
      (uint64(header[pos+0xd])<<8) |
      (uint64(header[pos+0xe])<<16) |
      (uint64(header[pos+0xf])<<24)
  }
  if size := int64(sectors*512); size > nbytes {
    return -1
  } else if size > 512 { ret++ }
  
  return ret
  
} // detect_MBR
