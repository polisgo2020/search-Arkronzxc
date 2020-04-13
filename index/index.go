package index

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	"github.com/polisgo2020/search-Arkronzxc/util"
	"github.com/rs/zerolog/log"

	"github.com/polisgo2020/search-Arkronzxc/files"
)

type Index map[string][]string

// CreateInvertedIndex returns map where key is a word in file, value is filename
func CreateInvertedIndex(files []string) (*Index, error) {
	log.Debug().Strs("Files", files)
	m := make(Index)

	wg := sync.WaitGroup{}
	fileChan := make(chan map[string]string, 1000)

	for i := range files {
		wg.Add(1)
		go ConcurrentBuildFileMap(&wg, files[i], fileChan)
	}

	go func(wg *sync.WaitGroup, readChan chan map[string]string) {
		wg.Wait()
		close(readChan)
	}(&wg, fileChan)

	for data := range fileChan {
		for j := range data {
			if m[j] == nil {
				m[j] = []string{data[j]}
			} else {
				m[j] = append(m[j], data[j])
			}
		}
	}
	log.Debug().Interface("Inverted index", m).Msg("Inverted index created")
	return &m, nil
}

// ConcurrentBuildFileMap concurrently writes words into the word array and iterates over it applying filename as value
func ConcurrentBuildFileMap(wg *sync.WaitGroup, filename string, mapChan chan<- map[string]string) {
	log.Debug().Interface("Wg", wg).Str("Filename", filename)
	defer wg.Done()

	m := make(map[string]string)
	wordArr, err := files.ConcurrentReadFile(filename)
	if err != nil {
		log.Err(err).Msg("Error while reading file concurrently")
		return
	}
	log.Debug().Strs("Word array", wordArr)
	for i := range wordArr {
		m[wordArr[i]] = filename
	}
	log.Debug().Interface("Map", m)
	mapChan <- m
}

// buildSearchIndex searches by index and returns the structure where the key is the file name, and the value is the
// number of words from the search query that were found in this file
func (m *Index) BuildSearchIndex(searchArgs []string) (map[string]int, error) {
	log.Debug().Interface("Index", m).Strs("Search args", searchArgs)

	ans := make(map[string]int)

	var cleanData []string
	for i := range searchArgs {
		w, err := util.CleanUserData(searchArgs[i])
		if err != nil {
			return nil, err
		}
		if w != "" {
			cleanData = append(cleanData, w)
		}
	}
	log.Debug().Strs("Clean data", cleanData)

	for _, v := range cleanData {
		if filesArray, ok := (*m)[v]; ok {
			for _, fileName := range filesArray {
				ans[fileName]++
			}
		}
	}

	log.Debug().Interface("Ans", ans).Msg("Search index successfully filled")
	return ans, nil
}

func UnmarshalFile(filename string) (*Index, error) {
	log.Debug().Str("Filename", filename)

	var m *Index
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
