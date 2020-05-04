package util

import (
	"os"

	"github.com/kljensen/snowball"
	"github.com/rs/zerolog/log"
)

// CleanUserData receives a processing word and after processing applies function operation for processed word
func CleanUserData(word string) (string, error) {
	if !EnglishStopWordChecker(word) && len(word) > 0 {
		stemmedWord, err := snowball.Stem(word, "english", false)
		if err != nil {
			log.Err(err).Msg("error while stemming the word")
			return "", err
		}
		return stemmedWord, nil
	}
	return "", nil
}

func FileSize(path string) int64 {
	log.Debug().Str("path", path)

	fi, err := os.Stat(path)
	if err != nil {
		log.Fatal().Msg("error while calculating file size")
	}

	log.Debug().Int64("file size", fi.Size()).Msg("file size calculated")

	return fi.Size()
}
