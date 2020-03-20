package main

import (
	"github.com/kljensen/snowball"
	"github.com/polisgo2020/search-Arkronzxc/dictionary"
	"log"
)

func cleanUserData(searchArgs []string) []string {
	var cleanSearchArgs []string

	for _, word := range searchArgs {
		if !dictionary.EnglishStopWordChecker(word) && len(word) > 0 {
			stemmedWord, err := snowball.Stem(word, "english", false)
			if err != nil {
				log.Print(err)
				continue
			}
			cleanSearchArgs = append(cleanSearchArgs, stemmedWord)
		}
	}
	return cleanSearchArgs
}
