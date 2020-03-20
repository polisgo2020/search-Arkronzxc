package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("too few program arguments")
		return
	}
	err := CreateInvertedIndex(os.Args[1])
	if err != nil {
		log.Print(err)
	}
}
