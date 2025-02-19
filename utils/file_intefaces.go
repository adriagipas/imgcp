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
 * along with adriagipas/imgcp.  If not, see <https://www.gnu.org/licenses/>.
 */
/*
 *  file_interfaces.go - Interfícies per manipular fitxers.
 *
 */

package utils

type FileReader interface {
  
  // Llig en el buffer. Torna el nombre de bytes llegits. Quan aplega
  // al final torna 0 i io.EOF.
  Read(buf []byte) (int,error)

  // Tanca el fitxer.
  Close() error
  
}


type FileWriter interface {

  // Escriu el buffer. Torna el nombre de bytes escrits .
  Write(buf []byte) (int,error)

  // Tanca el fitxer
  Close() error
  
}
