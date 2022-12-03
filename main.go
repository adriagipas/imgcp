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
 *  main.go - Utilitat per manipular imatges de dispositius
 *            d'emmagatzemament.
 */

package main;

import (
  "log"
  
  "github.com/adriagipas/imgcp/ops"
  "github.com/adriagipas/imgcp/utils"
)

func main() {

  // Inicialitza log
  log.SetPrefix ( "[imgcp] " )
  log.SetFlags ( 0 )

  // Executa operació
  if args,err := utils.NewArgs(); err == nil {
    if len(args.Files) > 0 {
      switch args.Op {
      case utils.OP_SHOW:
        err= ops.Show ( args )
      case utils.OP_LIST:
        err= ops.List ( args )
      case utils.OP_CAT:
        err= ops.Cat ( args )
      case utils.OP_MKDIR:
        err= ops.Mkdir ( args )
      case utils.OP_COPY:
        err= ops.Copy ( args )
      default:
        err= ops.Show ( args )
      }
      if err != nil {
        log.Fatal ( err )
      }
    }
  } else {
    log.Fatal ( err )
  }
  
}
