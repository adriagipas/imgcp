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
 * show.go - Implementa l'operació SHOW. Mostra per pantala la
 *           informació de la imatge.
 */

package ops

import (
  "fmt"
  "os"

  "github.com/adriagipas/imgcp/imgs"
  "github.com/adriagipas/imgcp/utils"
)


/**********************/
/* FUNCIONS PÚBLIQUES */
/**********************/

func Show ( args *utils.Args ) error {

  // No es suporten arguments
  if len(args.OpArgs) != 0 {
    return fmt.Errorf ( "(SHOW) invalid arguments: %v", args.OpArgs )
  }

  // Executa operació
  print_name := len(args.Files)>1
  for name,file := range args.Files {
    fmt.Println("")
    if print_name {
      fmt.Printf("  %s) \"%s\"\n",name,file)
      fmt.Println("")
    }
    img,err := imgs.NewImage ( file )
    if err != nil {
      return err
    }
    if err = img.PrintInfo ( os.Stdout, "    " ); err != nil {
      return err
    }
    fmt.Println("")
  }
  
  return nil
  
} // end Show
