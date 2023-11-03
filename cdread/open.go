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
 *
 */

package cdread

import (
  "errors"
  "os"
)




/****************/
/* PART PÚBLICA */
/****************/

// Obté una estructura CD que serveix per a llegir la imatge del
// fitxer proporcionat. Si s'utilitza aquest mètodo s'opta per una
// estratègia baix demanda. És a dir, s'obri el fitxer per a generar
// l'estructura però després es tanca, i cada vegada que es demana un
// lector es torna a obrir el fitxer.
func Open( file_name string ) (CD,error) {
  return nil,errors.New("TODO - Open!!!")
} // end Open


// La diferència amb Open és que en aquesta aproximació el fitxer
// sempre es manté obert. Això sí, s'assumeix que la imatge comença al
// principi del fitxer, no en la posició actual.
func OpenFromFile( fd *os.File ) (CD,error) {
  return nil,errors.New("TODO - OpenFromFile!!!")
} // end OpenFromFile
