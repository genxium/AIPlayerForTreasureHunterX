package main
import (
  "os"
  "fmt"
)

func main(){
  fmt.Println(os.Args[1:])
  fmt.Println(os.Args[3])
}
