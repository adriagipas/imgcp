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
 *  image.go - Manipulació dels fitxers imatge.
 *
 */

package imgs;

import (
  "fmt"
  "io"
)


/*********/
/* IMAGE */
/*********/

type Image interface {

  // Imprimeix la informació de la imatge en el fitxer especificat la
  // informació de la imatge. Cada línia s'imprimeix amb el prefix
  // indicat.
  PrintInfo(file io.Writer, prefix string) error

  // Torna el directori arrel del dispositiu.
  GetRootDirectory() (Directory,error)
  
}

// Retorna la Imatge associada al fitxer expecificat. Si tot va bé
// error és nil.
func NewImage(file_name string) (Image,error) {

  // Obté tipus
  ftype,err := Detect ( file_name )
  if err != nil { return nil,err }

  // Crea imatge
  switch ftype {

  case TYPE_LOCAL_FOLDER:
    return newLocalFolder ( file_name )
    
  case TYPE_FAT12:
    return newFAT12 ( file_name )
    
  case TYPE_FAT16:
    return newFAT16 ( file_name )
    
  case TYPE_MBR:
    return newMBR ( file_name ),nil

  case TYPE_IFF:
    return newIFF ( file_name )

  case TYPE_CD:
    return newCD ( file_name )

  case TYPE_ISO9660:
    return newISO_9660_from_filename ( file_name )
    
  case TYPE_CCI:
    return newCCI ( file_name )
    
  case TYPE_NCCH:
    return newNCCH_from_filename ( file_name )

  case TYPE_STFS:
    return newSTFS ( file_name )
    
  default:
    return nil,fmt.Errorf ( "Unable to detect the image type for file '%s'",
      file_name)
  }
  
} // end NewImage


/***************/
/* FILE READER */
/***************/

type FileReader interface {
  
  // Llig en el buffer. Torna el nombre de bytes llegits. Quan aplega
  // al final torna 0 i io.EOF.
  Read(buf []byte) (int,error)

  // Tanca el fitxer.
  Close() error
  
}


/***************/
/* FILE WRITER */
/***************/

type FileWriter interface {

  // Escriu el buffer. Torna el nombre de bytes escrits .
  Write(buf []byte) (int,error)

  // Tanca el fitxer
  Close() error
  
}


/*************/
/* DIRECTORY */
/*************/

type Directory interface {

  // Torna un iterador a la primera entrada en l'ordre intern
  Begin() (DirectoryIter,error)

  // Crea directory. Si ja existeix no el torna a crear.
  MakeDir(name string) (Directory,error)

  // Torna un un FileWriter. Si el fitxer no existeix intenta crear-lo.
  GetFileWriter(name string) (FileWriter,error)
  
}


/******************/
/* DIRECTORY ITER */
/******************/

const DIRECTORY_ITER_TYPE_FILE        = 0
const DIRECTORY_ITER_TYPE_DIR         = 1
const DIRECTORY_ITER_TYPE_DIR_SPECIAL = 2
const DIRECTORY_ITER_TYPE_SPECIAL     = 3

type DirectoryIter interface {

  // Torna cert si el nom proporcionat és compatible amb el nom que
  // busquem.
  CompareToName(name string) bool
  
  // Indica final de fitxer.
  End() bool

  // Torna un directori amb el contingut del directori
  // actual. Intentar cridar aquest mètodes en entrades que no són
  // directoris torna un error.
  GetDirectory() (Directory,error)

  // Torna un FileReader del fitxer actual. Intentar cridar a aquest
  // mètode quan no és un fitxer torna un error.
  GetFileReader() (FileReader,error)

  // Torna el nom de l'entrada.
  GetName() string
  
  // Imprimeix én el fitxer indicat la línea que s'ha de veure per
  // pantalla d'eixe fitxer quan s'executa el comandament ls
  List(file io.Writer) error
  
  // Avança a la següent entrada
  Next() error

  // Elimina el fitxer o directori. En cas dels fitxers especials dona
  // error. ATENCIÓ!!!! Si s'intenta esborrar un directori no
  // s'assegura que s'esborren els fitxers apuntats per aquest. És
  // responsabilitat de l'usuari esborrar abans tots els fitxers abans
  // d'esborrar el directori.
  Remove() error
  
  // Retorna el tipus
  Type() int
  
}


/*************/
/* FIND PATH */
/*************/

type FindPathResult struct {
  IsDir  bool
  Dir    Directory
  FileIt DirectoryIter
}


// Busca el path en el directori especificat. Tornant un punter al
// directori si ho és o a l'iterador del fitxer si és un fitxer.
// NOTA!!! Quan és un directori FileIt conté l'iterador que apunta al
// directori, sempre i quan no siga l'arrel
func FindPath(
  
  dir         Directory,
  path        []string,
  path_is_dir bool,
  
) (FindPathResult,error) {

  // Prepara
  ret := FindPathResult {
    FileIt: nil, // Per si és un directori
    }
  tmp_path := path

  // Cerca
  for ; len(tmp_path)>0 ; {

    // Obté nom actual
    name := tmp_path[0]
    tmp_path= tmp_path[1:]

    // Actualitza path
    i,err := dir.Begin()
    for ; !i.End() && err == nil; err= i.Next () {
      if i.CompareToName ( name ) {
        break
      }
    }

    // Comprova resultat cerca
    if err != nil {    
      return ret,err
      
    } else if i.End() {
      return ret,fmt.Errorf ( "Path not found: %v", path )
      
    } else if i.Type () == DIRECTORY_ITER_TYPE_DIR ||
      i.Type () == DIRECTORY_ITER_TYPE_DIR_SPECIAL {
      ret.FileIt= i
      dir,err= i.GetDirectory ()
      if err != nil { return ret,err }
      
    } else { // Tipus fitxer
      if len(tmp_path) > 0 { // Encara queden més fitxers
        return ret,fmt.Errorf ( "Accessing a file as directory: %v",
          path )
        
      } else if path_is_dir { // Volíem accedir a un directori
        return ret,fmt.Errorf ( "Path (%v) is a file not a directory", path )
        
      } else { // El nostre fitxer
        ret.IsDir= false
        ret.FileIt= i
        return ret,nil
      }
      
    }
    
  }
  
  // Si aplega ací és un directori i l'hem trobat. Si no és l'arrel
  // FileIt tindrà el punter a l'iterador.
  ret.IsDir= true
  ret.Dir= dir
  
  return ret,nil
  
} // end FindPath
