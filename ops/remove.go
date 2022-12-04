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
 *  remove.go - Implementa l'operació REMOVE. Elimina fitxers o directoris.
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

func Remove ( args *utils.Args ) error {

  // Comprova que hi han PATHs
  if len(args.OpArgs) == 0 {
    return errors.New ( "no file paths provided to remove command" )
  }

  verbose := false
  if len(args.OpArgs)>1 {
    verbose= true
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
    
    // Elimina
    if res.IsDir {
      if err := RemoveDir ( path.Path, res.Dir ); err != nil {
        return err
      }
      if res.FileIt != nil {
        fmt.Printf ( "Removing %s ...\n", path.Path )
        if err := res.FileIt.Remove (); err != nil {
          return err
        }
      }
    } else {
      if verbose {
        fmt.Printf ( "Removing %s ...\n", path.Path )
      }
      if err := res.FileIt.Remove (); err != nil {
        return err
      }
    }
    
  }
  
  return nil
  
} // end Remove


func RemoveDir(prefix string, dir imgs.Directory) error {

  i,err := dir.Begin ()
  for ; err == nil && !i.End (); err= i.Next () {
    switch typ := i.Type (); typ {
      
    case imgs.DIRECTORY_ITER_TYPE_FILE:
      fmt.Printf ( "Removing %s/%s ...\n", prefix, i.GetName () )
      if err := i.Remove (); err != nil {
        return err
      }

    case imgs.DIRECTORY_ITER_TYPE_DIR:
      new_prefix := prefix + "/" + i.GetName ()
      new_dir,err := i.GetDirectory ()
      if err != nil { return err }
      if err := RemoveDir ( new_prefix, new_dir ); err != nil {
        return err
      }
      fmt.Printf ( "Removing %s/%s ...\n", prefix, i.GetName () )
      if err := i.Remove (); err != nil {
        return err
      }
      
    case imgs.DIRECTORY_ITER_TYPE_DIR_SPECIAL:
      name := i.GetName ()
      if name != "." && name != ".." {
        return fmt.Errorf ( "Special directory '%s/%s' cannot be removed",
          prefix, name )
      }

    case imgs.DIRECTORY_ITER_TYPE_SPECIAL:
      return fmt.Errorf ( "Special file '%s/%s' cannot be removed",
        prefix, i.GetName () )

    default:
      
    }
  }
  
  return nil
  
} // end RemoveDir
