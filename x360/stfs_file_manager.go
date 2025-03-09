/*
 * Copyright 2025 Adrià Giménez Pastor.
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
 *  stfs_file_manager.go - File manager per a volums de tipus STFS.
 */

package x360

import (
  "errors"
  "fmt"
  "io"
  "os"

  "github.com/adriagipas/imgcp/utils"
)


/****************/
/* FILE MANAGER */
/****************/

type _StfsFileManager struct {

  file_name         string
  blocks_per_hash   int32
  base_offset       int64
  read_only_format  bool
  root_active_index bool
  
}


func newStfsFileManager(
  
  file_name string,
  stfs      *STFS,
  
) (*_StfsFileManager,error) {

  ret:= _StfsFileManager{
    file_name: file_name,
  }

  // Bites volume descriptor
  ret.read_only_format= (stfs.Metadata.Volume[1]&0x1)!=0
  ret.root_active_index= (stfs.Metadata.Volume[1]&0x2)!=0
  
  // Si és read_only_format el hash table no estan repetits
  if ret.read_only_format {
    ret.blocks_per_hash= 1
  } else {
    ret.blocks_per_hash= 2
  }

  // El base offset és el HeaderSize redondejat al que ocupa un block
  // (0x1000)
  ret.base_offset= int64((stfs.Metadata.HeaderSize+0xfff)&0xf000)

   
  return &ret,nil
  
} // end newStfsFileManager


func (self *_StfsFileManager) Open(
  
  block       int32,
  num_blocks  int32,
  consecutive bool,
  size        int32, // <= 0 vol dir que no es sap (tots els blocs)
  
) (utils.FileReader,error) {
  return newStfsFile ( self, block, num_blocks, consecutive, size )
} // end _StfsFileManager.Open


func (self *_StfsFileManager) BlockToOffset( block int32 ) int64 {

  // Hi ha 3 nivells de taules hash. El primer és de blocks, el segon
  // de taules hash, i el tercer de taules hash del nivell
  // anterior. Cada hash ocupa 1 o 2 blocks (depen de metadades) i va
  // al principi. Totes les hash fan referència a 170 valors.
  //
  // El tema és que els blocks de Hash estan intercalats amb els de
  // dades i no afecta a l'índex dels blocks.

  // Reajusta block
  var extra_blocks int32
  extra_blocks= (block/170 + 1)*self.blocks_per_hash
  if block >= 170 {
    extra_blocks+= (block/(170*170) + 1)*self.blocks_per_hash
    if block >= 170*170 {
      extra_blocks+= (block/(170*170*170) + 1)*self.blocks_per_hash
    }
  }
  block+= extra_blocks
  
  // Calcula offset
  ret:= self.base_offset + int64(block)<<12
  
  return ret
  
} // end BlockToOffset


func (self *_StfsFileManager) BlockToHashBlock( block int32, level int ) int32 {

  // Level 0
  if level == 0 {

    // Primera hash
    if block < 170 { return 0 }

    // --> Posició pel nivell 0
    block_step:= 170 + self.blocks_per_hash // Distància entre blocks
    ret:= (block/170)*block_step
    // --> Extra pel nivell 1
    ret+= ((block/(170*170)) + 1)*self.blocks_per_hash

    if block < 170*170 {
      return ret
    } else {
      return ret + self.blocks_per_hash
    }

    // Level 1
  } else if level == 1 {
    
    // El primer block Level 1 està al final de la primera taula
    if block < 170*170 { return 170 + self.blocks_per_hash }

    // Posició
    block_step:= 170*170 + 171*self.blocks_per_hash
    ret:= (block/(170*170))*block_step

    return ret + self.blocks_per_hash

    // Level 2
  } else {
    return 170*170 + 171*self.blocks_per_hash
  }
  
} // end BlockToHashBlock


func (self *_StfsFileManager) BlockToHashOffset(

  block int32,
  level int,
  
) int64 {

  hash_block:= self.BlockToHashBlock ( block, level )
  ret:= self.base_offset + int64(hash_block)<<12

  return ret
  
} // end BlockToHashOffset


/********/
/* FILE */
/********/

const _STFS_BLOCK_SIZE = 0x1000
const _STFS_HASH_ENTRY_SIZE = 0x18

type _StfsFile struct {

  mng           *_StfsFileManager
  fd            *os.File
  v             [_STFS_BLOCK_SIZE]byte
  pv            []byte // Punter al buffer
  remain        int
  current_block int32
  block_count   int32
  consecutive   bool
  
}

func newStfsFile(

  mng         *_StfsFileManager,
  block       int32,
  num_blocks  int32,
  consecutive bool,
  size        int32, // <= 0 vol dir que no es sap (tots els blocs)
  
) (*_StfsFile,error) {
  
  // Comprovacions inicials
  if block < 0 {
    return nil,fmt.Errorf (
      "unable to open STFS File Reader starting in block %d",
      block )
  }
  if num_blocks <= 0 {
    return nil,fmt.Errorf (
      "unable to open STFS File Reader with block_count %d",
      num_blocks )
  }

  // Inicialitza.
  var err error
  ret:= _StfsFile{
    mng: mng,
    current_block: block,
    block_count: num_blocks-1, // Els blocks que falten
    consecutive: consecutive,
  }
  if size <= 0 {
    ret.remain= int(num_blocks*_STFS_BLOCK_SIZE)
  } else {
    ret.remain= int(size)
  }
  if ret.fd,err= os.Open ( mng.file_name ); err != nil {
    return nil,err
  }
  if err= ret.loadCurrentBlock (); err != nil {
    return nil,err
  }
  
  return &ret,nil
  
} // end newStfsFile


// Carrega en memòria el current_block
func (self *_StfsFile) loadCurrentBlock() error {

  self.pv= self.v[:] // Apunta al principi
  offset:= self.mng.BlockToOffset ( self.current_block )
  if nbytes,err:= self.fd.ReadAt ( self.pv, offset ); err != nil {
    return fmt.Errorf ( "Error while reading block %d: %s",
      self.current_block, err )
  } else if nbytes != len(self.pv) {
    return fmt.Errorf (
          "Error while reading block %d: failed to read current block",
      self.current_block )
  }

  return nil
  
} // end loadCurrentBlock


func (self *_StfsFile) blockToHashOffset( block int32 ) (int64,error) {

  // Offset bàsic.
  off:= self.mng.BlockToHashOffset ( block, 0 )
  if self.mng.read_only_format { return off,nil }

  // Cal decidir si emprar els blocks secundaris i per tant reajustar
  // l'offset???
  return -1,errors.New("TODO - _StfsFile.blockToHashOffset")
  
} // end blockToHashOffset


func (self *_StfsFile) nextBlock( block int32 ) (int32,error) {

  if off,err:= self.blockToHashOffset ( block ); err == nil {

    // Llig entrada hash 
    var buf [_STFS_HASH_ENTRY_SIZE]byte
    offset:= off + int64((block%170)*_STFS_HASH_ENTRY_SIZE)
    if nbytes,err:= self.fd.ReadAt ( buf[:], offset ); err != nil {
      return -1,fmt.Errorf ( "Error while reading hash block for block %d: %s",
        block, err )
    } else if nbytes != len(buf) {
      return -1,fmt.Errorf (
        "Error while reading block %d: failed to read current block",
        self.current_block )
    }

    // Llig block
    tmp:= (uint32(buf[_STFS_HASH_ENTRY_SIZE-3])<<16) |
      (uint32(buf[_STFS_HASH_ENTRY_SIZE-2])<<8) |
      uint32(buf[_STFS_HASH_ENTRY_SIZE-1])
    if tmp&0x800000 != 0 {
      tmp|= 0xFF000000
    }
    ret:= int32(tmp)
    
    return ret,nil
    
  } else {
    return -1,err
  }
  
} // end nextBlock


func (self *_StfsFile) loadNextBlock() error {
  
  // Comprovacions.
  if self.block_count == 0 {
    return errors.New (
      "Error while loading next block: no more blocks remaining" )
  }
  
  // Obté el següent block
  self.block_count--
  if self.consecutive {
    self.current_block++
  } else {
    var err error
    self.current_block,err= self.nextBlock ( self.current_block )
    if err != nil { return err }
  }
  
  // Carrega el block.
  if err:= self.loadCurrentBlock (); err != nil {
    return err
  }
  
  return nil
  
} // end loadNextBlock


func (self *_StfsFile) Read( buf []byte ) (int,error) {

  // Cas especial EOF
  if self.remain == 0 {
    return 0,io.EOF
  }

  // Llig
  ret:= len(buf)
  if ret > self.remain { ret= self.remain }
  for toread:= ret; toread > 0; {

    // Si el buffer està buit avança al següent
    if len(self.pv) == 0 {
      if err:= self.loadNextBlock (); err != nil {
        return -1,err
      }
    }

    // Llig del buffer.
    copy ( buf, self.pv )
    remain_pv:= len(self.pv)
    if toread >= remain_pv { // Consumisc el block senzer
      buf= buf[remain_pv:]
      self.pv= self.pv[remain_pv:]
      self.remain-= remain_pv
      toread-= remain_pv
    } else { // Consumisc el tros
      self.pv= self.pv[toread:]
      self.remain-= toread
      toread= 0
    }
    
  }
  
  return ret,nil
  
} // end _StfsFile.Read


func (self *_StfsFile) Close() error {
  
  if self.fd != nil {
    self.fd.Close ()
    self.fd= nil
  }

  return nil
  
} // end _StfsFile.Close
