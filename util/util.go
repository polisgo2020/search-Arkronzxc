package util

import (
	"log"
	"os"

	"github.com/kljensen/snowball"
)

//CleanUserData receives a processing word and after processing applies function operation for processed word
func CleanUserData(word string, operation func(word string)) {
	if !EnglishStopWordChecker(word) && len(word) > 0 {
		stemmedWord, err := snowball.Stem(word, "english", false)
		if err != nil {
			log.Print(err)
			return
		}
		operation(stemmedWord)
	}
}

func FileSize(path string) int64 {
	fi, err := os.Stat(path)
	if err != nil {
		log.Fatal(err)
	}

	return fi.Size()
}