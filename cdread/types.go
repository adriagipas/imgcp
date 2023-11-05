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

const (
  DISK_TYPE_AUDIO       = 0
  DISK_TYPE_MODE1       = 1
  DISK_TYPE_MODE1_AUDIO = 2
  DISK_TYPE_MODE2       = 3
  DISK_TYPE_MODE2_AUDIO = 4
  DISK_TYPE_UNK         = -1
)

type Position struct {
  Minutes uint8 // BCD, 74, (00h..73h)
  Seconds uint8 // BCD, 60, (00h..59h)
  Sector  uint8 // BCD, 75, (00h..74h)
}

type IndexInfo struct {
  Id  uint8    // Identificador en BCD 99 (01h..99h) Pot existir una 0.
  Pos Position
}

type TrackInfo struct {
  Id            uint8    // Identificador en BCD 99 (01h..99h)
  Indexes       []IndexInfo
  PosLastSector Position // Posició absoluta de l'últim sector del
                         // track
}

type SessionInfo struct {
  Tracks []TrackInfo
}

type Info struct {
  Sessions []SessionInfo
  Tracks   []TrackInfo
  Type     int
}

type CD interface {

  // Torna una estructura amb informació sobre l'estructura del CD.
  Info() *Info
  
  /* LOW_LEVEL??
  // Torna un lector.
  Reader() (Reader,error)
  */
  
}

/* LOW_LEVEL???
type Reader interface {

  // Tanca el lector.
  Close() error
  
  // Torna l'identificador de l'índex actual en BCD.
  CurrentIndex() uint8
  
  // Torna el número de la sessió actual. Començant per 0.
  CurrentSession() int

  // Torna el número (en sencer 1..99) (global) del 'track' actual.
  CurrentTrack() int

  // Mou la posició de lectura al principi de l'àrea 'Lead-in' de la
  // sessió actual.
  MoveToLeadIn() error
  
  // Mou la posició de lectura al principi de la SESS indicat
  // (1..?). Torna nil si s'ha pogut moure sense cap
  // problema. Internament llig el TOC.
  MoveToSession(session int) error

  // Mou la posició de lectura al principi del TRACK indicat (1..99)
  // (índex global) (No és BCD!!!). Torna nil si s'ha pogut moure
  // sense cap problema.
  MoveToTrack(track int) error

  // Torna el nombre de sessions.
  NumSessions() int

  // Llig en BUF el sector actual (grandària SECTOR_SIZE bytes) i
  // avança al següent sector si MOVE és cert. El contingut del sector
  // és tot (no sóls el datafield) i de fet pot ser que no siga un
  // sector de dades. IS_AUDIO indica si el sector és d'audio.
  Read(buf []byte,move bool) (is_audio bool,err error)
  
  // Fica el disc en l'estat inicial, com si haguerem reinsertat el
  // disc en la unitat.
  Reset() error

  // Fa un seek a la posició indicada (Minut.Segon.Sector). Torna nil
  // si tot ha anat bé. Els valors estan en decimal.
  Seek(minut int,segon int,sector int) error

  // Torna la posició actual (és en BCD).
  Tell() Position
  
}
*/
