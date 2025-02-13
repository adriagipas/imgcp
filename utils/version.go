/*
 * Copyright 2022-2025 Adrià Giménez Pastor.
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
 *  version.go - Versió del programa.
 *
 */

package utils;

import "fmt"

const VERSION = "1.3.2"

func PrintVersion() {

  P := fmt.Println

  P("imgcp "+VERSION)
  P("Copyright (C) 2022-2025 Adrià Giménez Pastor")
  P("License GPLv3+: GNU GPL version 3 or later "+
    "<https://gnu.org/licenses/gpl.html>.")
  P("This is free software: you are free to change and redistribute it.")
  P("There is NO WARRANTY, to the extent permitted by law.")
  P("")
  
} // end PrintVersion
