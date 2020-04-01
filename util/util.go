package util

import (
	"fmt"
	"log"
	"os"

	"github.com/kljensen/snowball"
)

type StopWordError struct {
	What string
}

func (e StopWordError) Error() string {
	return fmt.Sprintf("%s", e.What)
}

//CleanUserData receives a processing word and after processing applies function operation for processed word
func CleanUserData(word string) (string, error) {
	if !EnglishStopWordChecker(word) && len(word) > 0 {
		stemmedWord, err := snowball.Stem(word, "english", false)
		if err != nil {
			log.Print(err)
			return "", err
		}
		return stemmedWord, nil
	}
	return "", nil
}

func FileSize(path string) int64 {
	fi, err := os.Stat(path)
	if err != nil {
		log.Fatal(err)
	}

	return fi.Size()
}
