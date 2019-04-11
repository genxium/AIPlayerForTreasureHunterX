package main

import (
	"flag"
	"fmt"
)

func main() {
	flag.Parse()
	args := flag.Args()
	fmt.Println(args)
}

func test2() {

	defer fmt.Println("kobako")
	defer fmt.Println("1111")

}

func test() {

	done1 := make(chan bool)
	done2 := make(chan bool)

	go func() {
		defer close(done1)
		fmt.Println("11111")
	}()

	go func() {
		defer close(done2)
		fmt.Println("22222")
	}()

	for {
		select {
		case <-done1:
			fmt.Println("done1")
			return
		case <-done2:
			fmt.Println("done2")
			return
		}
	}
}
