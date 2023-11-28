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
 *  open.go - Obri imatges de CDs.
 */

package cdread




// Obté una estructura CD que serveix per a llegir els tracks de ls la
// imatge del fitxer proporcionat.
func Open( file_name string ) (CD,error) {

  // NOTA! Tenint en compte com és cada format crec que l'òptim és
  // directament intentar llegir-los seguint el següent ordre.
  cd,err:= OpenMds ( file_name )
  if err == nil { return cd,nil }
  cd,err= OpenIso ( file_name )
  if err == nil { return cd,nil }
  cd,err= OpenCue ( file_name )
  
  return cd,err
  
} // end Open
