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
 *  cat.go - Implementa l'operació CAT. Concatena fitxers i els
 *           imprimeix per pantalla.
 *
 */

package ops

import (
  "errors"
  "fmt"
  "os"
  
  "github.com/adriagipas/imgcp/imgs"
  "github.com/adriagipas/imgcp/utils"
)


/************/
/* OPERACIÓ */
/************/

const CAT_BUF_SIZE = 1024

func Cat ( args *utils.Args ) error {

  // Comprova que hi han PATHs
  if len(args.OpArgs) == 0 {
    return errors.New ( "no file paths provided to cat command" )
  }

  // Buffer
  var mem [CAT_BUF_SIZE]byte
  buf := mem[:]
  
  // Processa args
  for _,arg := range args.OpArgs {

    // Obté path
    path,err := args.GetPath ( arg )
    if err != nil { return err }

    // Comprova que no és un directori
    if path.IsDir {
      return errors.New ( "cat command cannot be applied over directories" )
    }

    // Crea imatge
    img,err := imgs.NewImage ( path.FileName )
    if err != nil { return err }

    // Obté directory root
    dir,err := img.GetRootDirectory ()
    if err != nil { return err }

    // Processa path
    res,err := imgs.FindPath ( dir, path.Paths, path.IsDir )
    if err != nil { return err }
    if res.IsDir {
      return fmt.Errorf ( "'%v' is a directory not a file", path.Paths )
    }
    
    // Obri fitxer
    f,err := res.FileIt.GetFileReader ()
    if err != nil { return err }
    
    // Llig i imprimeix
    nbytes,err := f.Read ( buf )
    if err != nil { return err }
    for ; nbytes > 0; {
      n,err := os.Stdout.Write ( buf[:nbytes] )
      if err != nil { return err }
      if n != nbytes {
        return errors.New ( "Unexpected error while writing to"+
          " standard output" )
      }
      nbytes,err= f.Read ( buf )
      if err != nil { return err }
    }
    
    // Tanca el fitxer
    f.Close ()
    
  }
  
  return nil
  
} // end List
