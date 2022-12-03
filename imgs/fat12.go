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
 *  fat12.go - Implementa el sistema de fitxers FAT12.
 *
 */

package imgs

import (
  "fmt"
  "io"
  "os"
  
  "github.com/adriagipas/imgcp/utils"
)


/*********/
/* FAT16 */
/*********/

func newSubimgFAT12(file_name string, offset int64, length uint64,
) (*_FAT1216,error) {
  return newFAT1216 ( file_name, offset, length, false )  
}


func newFAT12(file_name string) (*_FAT1216,error) {

  // Obté grandària
  f,err := os.Open ( file_name )
  if err != nil { return nil,err }
  info,err := f.Stat ()
  if err != nil { return nil,err }
  f.Close ()

  // Obté image
  return newSubimgFAT12 ( file_name, 0, uint64(info.Size ()) )
  
}

/***************/
/* FAT12 TABLE */
/***************/

type _FAT12_Table []byte


func (self _FAT12_Table) badCluster() uint16 {
  return 0xFF7
}


func (self _FAT12_Table) length() uint16 {
  return uint16((len(self)*2)/3)
}


func (self _FAT12_Table) get(ind int) uint16 {
  
  var ret uint16

  pos := (3*ind)/2
  if (ind&0x1) == 0 { // Parell
    ret= uint16(self[pos]) | (uint16(self[pos+1]&0xF)<<8)
  } else { // Imparell
    ret= uint16(self[pos]>>4) | (uint16(self[pos+1])<<4)
  }

  return ret
  
} // end get


func (self _FAT12_Table) chain(ind uint16) uint16 {
  return self.get ( int(ind) )
} // end chain


func (self _FAT12_Table) fPrintInfo(
  
  f      *os.File,
  file   io.Writer,
  prefix string,
  br     *_FAT1216_BR,
  
)  error {

  // Conta clusters disponibles i fitxers
  var bad, free, nfiles int
  total := self.length () - 2
  for i := 2; i < int(self.length()); i++ {
    e := self.get ( i )
    if e == 0 {
      free++
    } else if e == 0xFF7 {
      bad++
    } else if e >= 0xFF8 {
      nfiles++
    }
  }

  // Imprimeix informació
  cluster_size := uint64(br.bpb.bytes_per_sec) * uint64(br.bpb.secs_per_clu)
  fmt.Fprintln ( file, "" )
  fmt.Fprintf ( file, "%sUsage\n", prefix )
  fmt.Fprintf ( file, "%s-----\n\n", prefix )
  fmt.Fprintf ( file, "" )
  fmt.Fprintf ( file, "%s  * NUM. FILES:    %d\n", prefix, nfiles )
  fmt.Fprintf ( file, "%s  * FREE CLUSTERS: %d (%.1f%% [%s])\n",
    prefix, free, 100*(float32(free)/float32(total)),
    utils.NumBytesToStr ( uint64(free)*cluster_size ) )
  fmt.Fprintf ( file, "%s  * BAD CLUSTERS:  %d (%.1f%% [%s])\n",
    prefix, bad, 100*(float32(bad)/float32(total)),
    utils.NumBytesToStr ( uint64(bad)*cluster_size ) )
  
  return nil
  
} // end fPrintInfo


func (self _FAT12_Table) getData() []byte {
  return self
}


func (self _FAT12_Table) write(ind uint16, val uint16) {
  
  pos := (3*int(ind))/2
  if (ind&0x1) == 0 { // Parell
    self[pos]= uint8(val)
    self[pos+1]= (self[pos+1]&0xF0) | (uint8(val>>8)&0x0F)
  } else { // Imparell
    self[pos]= (self[pos]&0x0F) | (uint8(val&0x0F)<<4)
    self[pos+1]= uint8(val>>4)
  }

} // end write
