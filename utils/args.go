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
 *  args.go - Processament de la línia de comandaments.
 *
 */

package utils;

import (
  "errors"
  "fmt"
  "os"
  "strconv"
  "strings"
)


/*********/
/* TIPUS */
/*********/

type Args struct {

  // Diccionari amb els fitxers
  Files map[string]string

  // Operador i arguments
  Op     int
  OpArgs []string

  // Estat ocult
  no_names int // Compta quants fitxers hi han sense nom
  
}


/*************/
/* CONSTANTS */
/*************/

const OP_NONE  = 0
const OP_SHOW  = 1
const OP_LIST  = 2
const OP_CAT   = 3
const OP_MKDIR = 4
const OP_COPY  = 5


/*********************/
/* FUNCIONS PRIVADES */
/*********************/

func print_usage() {
  P := fmt.Println
  P("USAGE:\n")
  P("  imgcp <IMGs> [<OP>]\n")
  P("    <IMGs>: <IMG> [<IMG>]*")
  P("    <IMG>:  <image file name> | <NAME>=<image file name>")
  P("    <NAME>: [A-Z]+")
  P("    <PATH>: <PATH_NONAME> | <NAME>=<PATH_NONAME>")
  P("    <PATH_NONNAME>: A file path separated by '/'")
  P("")
  P("    <OP>: <OP_SHOW> | <OP_LIST> | <OP_CAT> | <OP_MKDIR> | <OP_COPY>")
  P("")
  P("    <OP_CAT> : cat <PATH> [<PATH>]*")
  P("")
  P("    <OP_COPY> : (copy | cp) <PATH> [<PATH>]* <PATH>")
  P("")
  P("    <OP_LIST> : (list | ls) <PATH> [<PATH>]*")
  P("")
  P("    <OP_MKDIR> : mkdir <PATH> [<PATH>]*")
  P("")
  P("    <OP_SHOW>: show | sh")
  P("")
  P("OPERATIONS:\n")
  P("  cat: Similar to the UNIX cat command, concatenate files and print")
  P("       on the standard output")
  P("")

  P("  copy: Copy files from one image (or host) to another image (or host).")
  P("        Destionation path is always the last provided path. Several")
  P("        source paths can be provided. If the source path is a directory")
  P("        then it is copied recursively.")
  P("")
  P("  list: Similar to the UNIX ls command, show the content inside")
  P("        the provided PATH. If the PATH is a file show the properties")
  P("        of the provided PATH")
  P("")
  P("  mkdir: Similar to the UNIX mkdir command, creates a directory for")
  P("         a provided path. All subdirectories in the path are also")
  P("         created.")
  P("")
  P("  show: This is the default operation. Show the information")
  P("        of the current files.")
  P("")
}


func check_name(name string) bool {
  for i := 0; i < len(name); i++ {
    if name[i] < 'A' || name[i] > 'Z' {
      return false
    }
  }
  return true
}


func (self *Args) register_filename(file_name string) error {

  var name string

  // Obté el nom
  ind := strings.Index ( file_name, "=" )
  if ind == -1 {
    name= strconv.FormatInt ( int64(self.no_names+1), 10 )
    self.no_names++
  } else if ind == 0 || ind == len(file_name)-1 {
    return errors.New("wrong file name syntax: "+file_name)
  } else {
    aux := strings.SplitN ( file_name, "=", 2 )
    if len(aux) != 2 || !check_name ( aux[0] ) {
      return errors.New("wrong file name syntax: "+file_name)
    }
    name,file_name= aux[0],aux[1]
  }

  // Intenta registrar
  if _,ok := self.Files[name]; ok {
    return errors.New("repeated file name: "+name)
  }
  self.Files[name]= file_name
  
  return nil
  
} // register_filename


/**********************/
/* FUNCIONS PÚBLIQUES */
/**********************/

func NewArgs() (*Args,error) {

  // Crea arguments
  args := Args {
    Op     : OP_NONE,
    Files  : make(map[string]string),
    OpArgs : os.Args[:0],
  }
  
  // Processa arguments
  for i := 1; i < len(os.Args); i++ {
    if os.Args[i]=="show" || os.Args[i]=="sh" { // Operació show
      args.Op= OP_SHOW
      args.OpArgs= os.Args[i+1:]
      break
    } else if os.Args[i]=="list" || os.Args[i]=="ls" { // Operació list
      args.Op= OP_LIST
      args.OpArgs= os.Args[i+1:]
      break
    } else if os.Args[i]=="cat" { // Operació cat
      args.Op= OP_CAT
      args.OpArgs= os.Args[i+1:]
      break
    } else if os.Args[i]=="mkdir" { // Operació mkdir
      args.Op= OP_MKDIR
      args.OpArgs= os.Args[i+1:]
      break
    } else if os.Args[i]=="copy" || os.Args[i]=="cp" { // Operació list
      args.Op= OP_COPY
      args.OpArgs= os.Args[i+1:]
      break
    } else { // Filename
      if err := args.register_filename ( os.Args[i] ); err != nil {
        return nil,err
      }
    }
  }
  
  // Si no té fitxers mostra usage
  if len(args.Files) == 0 {
    print_usage ()
  }
  
  return &args,nil
  
} // end NewArgs


// Aquesta funció processa un 'string' representant un path a fitxer i
// torna un objecte PATH.
func (self *Args) GetPath (path string) (*Path,error) {

  // Trim string
  path= strings.TrimSpace ( path )
  if len(path) == 0 {
    return nil,errors.New ( "Empty path" )
  }
  opath := path

  // Comprovacions de sintaxis. No es permeten dobles separados.
  if ind := strings.Index ( path, "//" ); ind != -1 {
    return nil,errors.New("wrong syntax for path: "+path)
  }

  // Obté el nom del fitxer
  var name string
  ind := strings.Index ( path, "=" )
  if ind == -1 {
    name= strconv.FormatInt ( 1, 10 )
  } else if ind == 0 || ind == len(path)-1 {
    return nil,errors.New("wrong syntax for path: "+path)
  } else {
    aux := strings.SplitN ( path, "=", 2 )
    if len(aux) != 2 || !check_name ( aux[0] ) {
      return nil,errors.New("wrong syntax for path: "+path)
    }
    name,path= aux[0],strings.TrimSpace ( aux[1] )
  }

  // Obté el nom del fitxer real
  file_name,ok := self.Files[name]
  if !ok {
    return nil,fmt.Errorf ( "Unknown file name: %s", name )
  }

  // Crea Path
  var paths []string
  is_dir := false
  if path != "/" {
    if path[0]=='/' { path= path[1:] }
    if path[len(path)-1]=='/' {
      path= path[:len(path)-1]
      is_dir= true
    }
    paths= strings.Split ( path, "/" )
  } else {
    paths= []string{}
    is_dir= true
  }
  ret := Path{
    FileName: file_name,
    Path: opath,
    Paths: paths,
    IsDir: is_dir,
  }
  
  return &ret,nil
  
} // end GetPath


/********/
/* PATH */
/********/

type Path struct {
  FileName string    // Nom del fitxer on estem buscant
  Path     string
  Paths    []string  // Camí al fitxer que busquem, si està buit vol
                     // dir que busquem en l'arrel
  IsDir    bool      // Si l'últim caràcter és un / s'enten que es vol
                     // accedir a este fitxer com si fora un directori.
}
