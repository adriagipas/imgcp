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
 *  fat16.go - Implementa el sistema de fitxers FAT16.
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

func newSubimgFAT16(file_name string, offset int64, length uint64,
) (*_FAT1216,error) {
  return newFAT1216 ( file_name, offset, length, true )  
}


func newFAT16(file_name string) (*_FAT1216,error) {

  // Obté grandària
  f,err := os.Open ( file_name )
  if err != nil { return nil,err }
  info,err := f.Stat ()
  if err != nil { return nil,err }
  f.Close ()

  // Obté image
  return newSubimgFAT16 ( file_name, 0, uint64(info.Size ()) )
  
}

/***************/
/* FAT16 TABLE */
/***************/

type _FAT16_Table []byte


func (self _FAT16_Table) badCluster() uint16 {
  return 0xFFF7
}


func (self _FAT16_Table) length() uint16 {
  return uint16(len(self)/2)
}


func (self _FAT16_Table) get(ind int) uint16 {
  return uint16(self[ind*2]) | (uint16(self[ind*2+1])<<8)
} // end get


func (self _FAT16_Table) chain(ind uint16) uint16 {
  return self.get ( int(ind) )
} // end chain


func (self _FAT16_Table) fPrintInfo(
  
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
    } else if e == 0xFFF7 {
      bad++
    } else if e >= 0xFFF8 {
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


func (self _FAT16_Table) getData() []byte {
  return self
}


func (self _FAT16_Table) write(ind uint16, val uint16) {

  pos := int(ind)*2
  self[pos]= uint8(val)
  self[pos+1]= uint8(val>>8)
  
} // end write
