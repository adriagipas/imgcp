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
 *  types.go - Tipus bàsics (una adaptació de
 *             https://github.com/adriagipas/CD).
 */

package cdread

const SECTOR_SIZE = 0x930

type Position struct {
  Minutes uint8 // BCD, 74, (00h..73h)
  Seconds uint8 // BCD, 60, (00h..59h)
  Sector  uint8 // BCD, 75, (00h..74h)
}

type IndexInfo struct {
  Id  uint8    // Identificador en BCD 99 (01h..99h) Pot existir una 0.
  Pos Position
}

const (
  TRACK_TYPE_AUDIO          = 0
  TRACK_TYPE_MODE1_RAW      = 1
  TRACK_TYPE_MODE2_RAW      = 2
  TRACK_TYPE_MODE2_CDXA_RAW = 3
  TRACK_TYPE_ISO            = 4
  TRACK_TYPE_UNK            = -1
)

type TrackInfo struct {
  Id            uint8    // Identificador en BCD 99 (01h..99h)
  Indexes       []IndexInfo
  PosLastSector Position // Posició absoluta de l'últim sector del
                         // track
  Type          int
}

type SessionInfo struct {
  Tracks []TrackInfo
}

type Info struct {
  Sessions []SessionInfo
  Tracks   []TrackInfo
}

// Açò sols afecta als CD-XA
const (
  MODE_DATA            = 0 // Es pot ficar 0. En CD-XA ignora els
                           // sectors Form2 .
  MODE_CDXA            = 1 // Torna tots els sectors (cadascun en la
                           // grandària que toque).
  MODE_CDXA_MEDIA_ONLY = 2 // Ignora sectors Form1
)

type CD interface {

  // Torna una estructura amb informació sobre l'estructura del CD.
  Info() *Info
  
  // Torna un lector de bytes d'un track.
  TrackReader(session int,track int,mode int) (TrackReader,error)
  
}

type TrackReader interface {

  // Tanca el lector. Deprés de tancat no es pot llegir.
  Close() error

  // Funciona exactament com la interfície Reader.
  Read(b []byte) (n int,err error)

  // Mou el lector al principi del sector (0 és el primer sector del
  // track) indicat.
  Seek(sector int64) error
  
}
