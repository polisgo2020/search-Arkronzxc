package util

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/polisgo2020/search-Arkronzxc/index"

	"github.com/kljensen/snowball"
	"github.com/rs/zerolog/log"
)

// CleanUserData receives a processing word and after processing applies function operation for processed word
func CleanUserData(word string) (string, error) {
	log.Debug().Str("word", word)

	if !EnglishStopWordChecker(word) && len(word) > 0 {
		stemmedWord, err := snowball.Stem(word, "english", false)
		if err != nil {
			log.Err(err).Msg("Error while stemming the word")
			return "", err
		}
		log.Debug().Str("stemmed word", stemmedWord)
		return stemmedWord, nil
	}
	return "", nil
}

func FileSize(path string) int64 {
	log.Debug().Str("path", path)

	fi, err := os.Stat(path)
	if err != nil {
		log.Fatal().Msg("Error while calculating file size")
	}

	log.Debug().Int64("file size", fi.Size()).Msg("file size calculated")
	return fi.Size()
}

func UnmarshalFile(filename string) (*index.Index, error) {
	log.Debug().Str("Filename", filename)

	var m *index.Index
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Err(err).Msg("Error while extracting content from file")
		return nil, err
	}
	log.Debug().Str("Content", string(content)).Msg("Content successfully extracted")

	if json.Unmarshal(content, &m) != nil {
		log.Err(err).Msg("Error while serializing file from JSON")
		return nil, err
	}
	log.Debug().Str("Content", string(content)).Interface("m", m).
		Msg("JSON successfully serialized into index")
	return m, nil
}
