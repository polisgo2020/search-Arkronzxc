package index

import (
	"log"
	"sync"

	"github.com/polisgo2020/search-Arkronzxc/util"

	"github.com/polisgo2020/search-Arkronzxc/files"
)

type Index map[string][]string

// returns map where key is a word in file, value is filename
func CreateInvertedIndex(files []string) (*Index, error) {
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
ReadLoop:
	for {
		//reads file data if channel is not closed and it has unread data
		data, ok := <-fileChan
		if !ok {
			break ReadLoop
		}
		for j := range data {
			if m[j] == nil {
				m[j] = []string{data[j]}
			} else {
				m[j] = append(m[j], data[j])
			}
		}
	}

	return &m, nil
}

//ConcurrentBuildFileMap concurrently writes words into the word array and iterates over it applying filename as value
func ConcurrentBuildFileMap(wg *sync.WaitGroup, filename string, mapCah chan<- map[string]string) {
	defer wg.Done()

	m := make(map[string]string)
	wordArr, err := files.ConcurrentReadFile(filename)
	if err != nil {
		log.Print(err)
		return
	}
	for i := range wordArr {
		m[wordArr[i]] = filename
	}
	mapCah <- m
}

func BuildSearchIndex(searchArgs []string, m *Index) (map[string]int, error) {
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

	for _, v := range cleanData {
		if filesArray, ok := (*m)[v]; ok {
			for _, fileName := range filesArray {
				ans[fileName]++
			}
		}
	}
	return ans, nil
}
