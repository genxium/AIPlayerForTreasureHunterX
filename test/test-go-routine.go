package main
import(
  "fmt"
)

func main(){

  go func(){
    panic("Oh shit")
  }()

  go func(){
    fmt.Println("DDDDD")
  }()

  fmt.Scanln()
  fmt.Println("done")

}
