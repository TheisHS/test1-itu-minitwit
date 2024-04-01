package main

import (
	"fmt"
)


func devLog(str string) {
  if env == "dev" {
    fmt.Println(str)
  }
}
