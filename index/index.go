package index

import (
	"github.com/polisgo2020/search-Arkronzxc/files"
	"github.com/polisgo2020/search-Arkronzxc/util"
	"log"
	"sync"
)

type Index map[string][]string

// returns map where key is a word in file, value is filename
func CreateInvertedIndex(files []string) (*Index, error) {
	m := new(Index)

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
		select {
		//reads file data if channel is not closed and it has unread data
		case data, ok := <-fileChan:
			if !ok {
				break ReadLoop
			}
			for j := range data {
				if (*m)[j] == nil {
					(*m)[j] = []string{data[j]}
				} else {
					(*m)[j] = append((*m)[j], data[j])
				}
			}
		}
	}

	return m, nil
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

	cleanData := make([]string, 0)
	for i := range searchArgs {
		util.CleanUserData(searchArgs[i], func(word string) {
			cleanData = append(cleanData, word)
		})
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
