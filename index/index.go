package index

import (
	"sync"

	"github.com/polisgo2020/search-Arkronzxc/util"
	"github.com/rs/zerolog/log"

	"github.com/polisgo2020/search-Arkronzxc/files"
)

type Index map[string][]string

// CreateInvertedIndex returns map where key is a word in file, value is filename
func CreateInvertedIndex(files []string) (*Index, error) {
	log.Debug().Strs("files", files)
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
	log.Debug().Interface("inverted index", m).Msg("inverted index created")
	return &m, nil
}

// ConcurrentBuildFileMap concurrently writes words into the word array and iterates over it applying filename as value
func ConcurrentBuildFileMap(wg *sync.WaitGroup, filename string, mapChan chan<- map[string]string) {
	log.Debug().Interface("wg", wg).Str("filename", filename)
	defer wg.Done()

	m := make(map[string]string)
	wordArr, err := files.ConcurrentReadFile(filename)
	if err != nil {
		log.Err(err).Msg("error while reading file concurrently")
		return
	}
	log.Debug().Strs("word array", wordArr)
	for i := range wordArr {
		m[wordArr[i]] = filename
	}
	log.Debug().Interface("map", m)
	mapChan <- m
}

// buildSearchIndex searches by index and returns the structure where the key is the file name, and the value is the
// number of words from the search query that were found in this file
func (m *Index) BuildSearchIndex(searchArgs []string) (map[string]int, error) {
	log.Debug().Interface("index", m).Strs("search args", searchArgs)

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
	log.Debug().Strs("clean data", cleanData)

	for _, v := range cleanData {
		if filesArray, ok := (*m)[v]; ok {
			for _, fileName := range filesArray {
				ans[fileName]++
			}
		}
	}

	log.Debug().Interface("answer", ans).Msg("search index successfully filled")
	return ans, nil
}
