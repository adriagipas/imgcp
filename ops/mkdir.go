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
 *  mkdir.go - Implementa l'operació MKDIR. Crea directoris.
 *
 */

package ops

import (
  "errors"
  "fmt"
  
  "github.com/adriagipas/imgcp/imgs"
  "github.com/adriagipas/imgcp/utils"
)


/************/
/* OPERACIÓ */
/************/

func Mkdir ( args *utils.Args ) error {

  // Comprova que hi han PATHs
  if len(args.OpArgs) == 0 {
    return errors.New ( "no file paths provided to cat command" )
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

    // Crea directori
    if err := MakeDirPath ( dir, path.Paths ); err != nil {
      return err
    }
    
  }
  
  return nil
  
} // end Mkdir


// Partint del directori arrel proporcionat crea tot els subdirectoris
// (si és necessari) que es proporcionen.
func MakeDirPath(
  
  root imgs.Directory,
  path []string,
  
) error {

  tmp_path := path
  dir := root
  for ; len(tmp_path) > 0; {

    // Obté subdir
    subdir := tmp_path[0]
    tmp_path= tmp_path[1:]
    
    // Cerca el subdirectori
    i,err := dir.Begin()
    for ; !i.End() && err == nil; err= i.Next () {
      if i.CompareToName ( subdir ) {
        break
      }
    }

    // Comprova resultat
    if err != nil {
      return err
      
    } else if i.End() { // Crea el directori
      dir,err= dir.MakeDir ( subdir )
      if err != nil { return err }
      
    } else if i.Type () == imgs.DIRECTORY_ITER_TYPE_DIR ||
      i.Type () == imgs.DIRECTORY_ITER_TYPE_DIR_SPECIAL { // S'ha trobat
      dir,err= i.GetDirectory ()
      if err != nil { return err }

    } else {  // És un fitxer
      return fmt.Errorf ( "Path (%v) includes a regular file path", path )
      
    }
    
  }

  return nil
  
} // end MakeDirPath
