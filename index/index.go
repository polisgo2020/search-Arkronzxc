package index

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/polisgo2020/search-Arkronzxc/util"

	"github.com/polisgo2020/search-Arkronzxc/files"
)

type Index map[string][]string

type searchResponse struct {
	Filename    string `json:"filename"`
	WordCounter int    `json:"wordCounter"`
}

// CreateInvertedIndex returns map where key is a word in file, value is filename
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

	for data := range fileChan {
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

// ConcurrentBuildFileMap concurrently writes words into the word array and iterates over it applying filename as value
func ConcurrentBuildFileMap(wg *sync.WaitGroup, filename string, mapChan chan<- map[string]string) {
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
	mapChan <- m
}

// buildSearchIndex searches by index and returns the structure where the key is the file name, and the value is the
// number of words from the search query that were found in this file
func (m *Index) buildSearchIndex(searchArgs []string) (map[string]int, error) {
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

func SearchHandler(input string) http.HandlerFunc {

	searchIndex, err := unmarshalFile(input)
	if err != nil {
		panic(err)
	}

	return func(writer http.ResponseWriter, request *http.Request) {

		writer.Header().Set("Content-Type", "application/json")

		parsedSearchPhrase, err, errCode := parseSearchPhrase(request)
		if err != nil {
			log.Print(err)
			http.Error(writer, http.StatusText(errCode), errCode)
			return
		}

		resp, err, errCode := answerFormation(searchIndex, parsedSearchPhrase)
		if err != nil {
			log.Print(err)
			http.Error(writer, http.StatusText(errCode), errCode)
			return
		}

		finalJson, err := json.Marshal(resp)
		if err != nil {
			log.Print(err)
			http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if _, err := fmt.Fprint(writer, string(finalJson)); err != nil {
			log.Print(err)
			http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}

func parseSearchPhrase(request *http.Request) ([]string, error, int) {

	searchPhrase := request.FormValue("search")

	rawUserInput := strings.ToLower(searchPhrase)

	parsedUserInput := strings.Split(rawUserInput, " ")

	cleanedUserInput := make([]string, 0)
	for i := range parsedUserInput {
		w, err := util.CleanUserData(parsedUserInput[i])
		if err != nil {
			log.Print(err)
			return nil, err, http.StatusBadRequest
		}
		if w != "" {
			cleanedUserInput = append(cleanedUserInput, w)
		}
	}
	return cleanedUserInput, nil, -1
}

func answerFormation(index *Index, cleanedUserInput []string) ([]*searchResponse, error, int) {

	ans, err := index.buildSearchIndex(cleanedUserInput)
	if err != nil {
		log.Print(err)
		return nil, err, http.StatusInternalServerError
	}

	var resp []*searchResponse
	for s := range ans {
		resp = append(resp, &searchResponse{
			Filename:    s,
			WordCounter: ans[s],
		})
		log.Printf("filename: %s, words encountered : %d", s, ans[s])
	}
	return resp, nil, -1
}

func unmarshalFile(filename string) (*Index, error) {
	var m *Index
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	if json.Unmarshal(content, &m) != nil {
		log.Print(err)
		return nil, err
	}
	return m, nil
}
