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
 *  list.go - Implementa l'operació LIST. Mostra per pantalla el
 *            contingut d'un directori o la informació d'un fitxer.
 *
 */

package ops

import (
  "errors"
  "os"

  "github.com/adriagipas/imgcp/imgs"
  "github.com/adriagipas/imgcp/utils"
)


/************/
/* OPERACIÓ */
/************/

func List ( args *utils.Args ) error {

  // Comprova que hi han PATHs
  if len(args.OpArgs) == 0 {
    return errors.New ( "no file paths provided to list command" )
  }

  // Processa args
  for _,arg := range args.OpArgs {

    // Obté path
    path,err := args.GetPath ( arg )
    if err != nil { return err }
    
    // Crea imatge
    img,err := imgs.NewImage ( path.FileName )
    if err != nil { return err }
    
    // Obté directory root
    dir,err := img.GetRootDirectory ()
    if err != nil { return err }
    
    // Processa path
    res,err := imgs.FindPath ( dir, path.Paths, path.IsDir )
    if err != nil { return err }
    
    // Llista 
    if res.IsDir {
      i,err := res.Dir.Begin()
      if err != nil { return err }
      for ; err == nil && !i.End(); err= i.Next() {
        if err:= i.List ( os.Stdout ); err != nil {
          return err
        }
      }
    } else {
      res.FileIt.List ( os.Stdout )
    }
    
  }
  
  return nil
  
} // end List
