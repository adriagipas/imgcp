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
 *  copy.go - Implementa l'operació COPY. Copia fitxers.
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

func Copy ( args *utils.Args ) error {

  // Comprova que hi han PATHs
  if len(args.OpArgs) <= 1 {
    return errors.New ( "at least two paths must be provided" )
  }

  // Obté destí
  dst_path,err := args.GetPath( args.OpArgs[len(args.OpArgs)-1] )
  if err != nil { return err }
  dst,err := getDstFileCopy ( dst_path)
  if err != nil { return err }
  args.OpArgs= args.OpArgs[:len(args.OpArgs)-1]

  // Opera
  if dst.is_dir {
    if err := copyArgsToDir ( args, dst ); err != nil {
      return err
    }
  } else {
    if err := copyArgsToFile ( args, dst ); err != nil {
      return err
    }
  }
  
  return nil
  
} // end Copy


func copyArgsToDir(args *utils.Args, dst DstFile) error {

  verbose := len(args.OpArgs)>1
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

    // Cerca Path
    res,err := imgs.FindPath ( dir, path.Paths, path.IsDir )
    if err != nil { return err }

    // Copia
    if res.IsDir {

      // Crea nom destí si l'origen no és l'arrel
      var dst_dir imgs.Directory= dst.dir
      if len(path.Paths) > 1 {
        dir,err := dst_dir.MakeDir ( path.Paths[len(path.Paths)-1] )
        if err != nil { return err }
        dst_dir= dir
      }
      
      // Còpia
      if err := copyDirToDir ( path.Path, res.Dir, dst_dir ); err != nil {
        return err
      }
      
    } else { // Fitxer a directori
      err := copyFileToDir ( path.Path, path.Paths[len(path.Paths)-1],
        res.FileIt, dst.dir, verbose )
      if err != nil { return err }
    }
    
  }
  
  return nil
  
} // end copyArgsToDir


func copyDirToDir(
  prefix  string,
  src_dir imgs.Directory,
  dst_dir imgs.Directory,
) error {

  // Recorre directori origen
  i,err := src_dir.Begin ()
  for ; !i.End () && err == nil; err= i.Next () {
    if i.Type () == imgs.DIRECTORY_ITER_TYPE_DIR {
      
      // Directori origen
      new_src_dir,err := i.GetDirectory ()
      if err != nil { return err }
      
      // Directori destí
      new_dst_dir,err := dst_dir.MakeDir ( i.GetName () )
      if err != nil { return err }
      
      // Copia directoris
      new_prefix := prefix + "/" + i.GetName ()
      err= copyDirToDir ( new_prefix, new_src_dir, new_dst_dir )
      if err != nil { return err }
      
    } else if i.Type () == imgs.DIRECTORY_ITER_TYPE_FILE {
      file_name := i.GetName ()
      path := prefix + "/" + file_name
      err := copyFileToDir ( path, file_name, i, dst_dir, true )
      if err != nil { return err }
    }
    
  }
  if err != nil { return err }
  
  return nil
  
} // end CopyDirToDir


func copyFileToDir(
  path      string,
  file_name string,
  file      imgs.DirectoryIter,
  dst_dir   imgs.Directory,
  verbose   bool,
) error {

  if verbose {
    fmt.Printf ( "Copying %s ...\n", path )
  }
  
  // Crea fitxer src
  src_f,err := file.GetFileReader ()
  if err != nil { return err }
  
  // Crea fitxer dst
  dst_f,err := dst_dir.GetFileWriter ( file_name )
  if err != nil { return err }

  // Còpia
  if err := copyFiles ( src_f, dst_f ); err != nil {
    return fmt.Errorf ( "An error occurred while copying '%s': %s",
      path, err )
  }
  src_f.Close ()
  dst_f.Close ()
  
  return nil
  
} // end copyFileToDir


func copyArgsToFile(args *utils.Args, dst DstFile) error {

  // Sols es permet copiar des d'un fitxer.
  if len(args.OpArgs) != 1 {
    return errors.New ( "Cannot copy multiple files into a single file" )
  }
  
  // Obté path
  path,err := args.GetPath( args.OpArgs[0] )
  if err != nil { return err }

  // Crea imatge
  img,err := imgs.NewImage ( path.FileName )
  if err != nil { return err }
  
  // Obté root
  dir,err := img.GetRootDirectory ()
  if err != nil { return err }

  // Busca el path
  res,err := imgs.FindPath ( dir, path.Paths, path.IsDir )
  if err != nil { return err }
  if res.IsDir {
    return fmt.Errorf ( "'%v' is a directory not a file", path.Paths )
  }

  // Copia fitxer
  src_f,err := res.FileIt.GetFileReader ()
  if err != nil { return err }
  if err := copyFiles ( src_f, dst.f ); err != nil {
    return fmt.Errorf ( "An error occurred while copying '%s' to '%s': %s",
      path.Path, dst.path.Path, err )
  }
  src_f.Close ()
  dst.f.Close ()

  return nil
  
} // end copyArgsToFile


type DstFile struct {
  path   *utils.Path
  is_dir bool
  dir    imgs.Directory
  f      imgs.FileWriter
}

func getDstFileCopy(path *utils.Path) (DstFile,error) {

  ret := DstFile{
    path: path,
  }
  
  // Crea imatge
  img,err := imgs.NewImage ( path.FileName )
  if err != nil { return ret,err }

  // Directori arrel
  ret.dir,err= img.GetRootDirectory ()
  if err != nil { return ret,err }
  ret.is_dir= true
  
  // Cerca
  tmp_path := path.Paths
  for ; len(tmp_path)>0; {

    // Obté nom actual
    name := tmp_path[0]
    tmp_path= tmp_path[1:]

    // Actualitza path
    i,err := ret.dir.Begin()
    for ; !i.End() && err == nil; err= i.Next () {
      if i.CompareToName ( name ) {
        break
      }
    }

    // Comprova resultat cerca
    if err != nil {
      return ret,err

    } else if i.End() {
      if len(tmp_path)>0 { // Error
        return ret,fmt.Errorf ( "Folder '%s' in path '%v' does not exist",
          name, path.Paths )

      } else if path.IsDir {
        return ret,fmt.Errorf ( "Destination path '%v' not found", path.Paths )
        
      } else { // El path apunta a un fitxer nou.
        ret.is_dir= false
        ret.f,err= ret.dir.GetFileWriter ( name )
        if err != nil { return ret,err }
      }
      
    } else if i.Type () == imgs.DIRECTORY_ITER_TYPE_DIR ||
      i.Type () == imgs.DIRECTORY_ITER_TYPE_DIR_SPECIAL {
      ret.dir,err= i.GetDirectory ()
      if err != nil { return ret,err }
      
    } else if len(tmp_path) > 0 { // Encara queden més fitxer
      return ret,fmt.Errorf ( "Accessing a file as directory: %v",
        path.Paths )
      
    } else if path.IsDir { // Volíem accedir a un directori
      return ret,fmt.Errorf ( "Destination path '%v' is a file not"+
        " a directory", path.Paths )

    } else { // Sobreescriu fitxer
      ret.is_dir= false
      ret.f,err= ret.dir.GetFileWriter ( name )
      if err != nil { return ret,err }
    }
    
  }

  return ret,nil
  
} // end getDstFileCopy


const COPY_BUF_SIZE = 1024

func copyFiles(src imgs.FileReader, dst imgs.FileWriter) error {
  
  // Buffer
  var mem [COPY_BUF_SIZE]byte
  buf := mem[:]

  // Copia
  nbytes,err := src.Read ( buf )
  if err != nil { return err }
  for ; nbytes > 0; {
    n,err := dst.Write ( buf[:nbytes] )
    if err != nil { return err }
    if n != nbytes {
      return errors.New ( "Unexpected error while writing to"+
        " file" )
    }
    nbytes,err= src.Read ( buf )
    if err != nil { return err }
  }
  
  return nil
  
} // end copyFiles
