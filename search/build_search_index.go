package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("too few program arguments")
		return
	}
	t, err := BuildSearchIndex(os.Args[2:], os.Args[1])
	if err != nil {
		log.Print(err)
		return
	}
	fmt.Printf("%+v", t)
}

func BuildSearchIndex(searchArgs []string, file string) (map[string]int, error) {
	ans := make(map[string]int)

	m, err := readAndParseFile(file)
	if err != nil {
		log.Print(err)
	}

	for _, v := range cleanUserData(searchArgs) {
		if filesArray, ok := m[v]; ok {
			for _, fileName := range filesArray {
				ans[fileName]++
			}
		}
	}
	return ans, nil
}
