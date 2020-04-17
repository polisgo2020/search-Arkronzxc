package web

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/polisgo2020/search-Arkronzxc/index"
	"github.com/polisgo2020/search-Arkronzxc/util"
)

type searchResponse struct {
	Filename    string `json:"filename"`
	WordCounter int    `json:"wordCounter"`
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

func answerFormation(index *index.Index, cleanedUserInput []string) ([]*searchResponse, error, int) {

	ans, err := index.BuildSearchIndex(cleanedUserInput)
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

func unmarshalFile(filename string) (*index.Index, error) {
	var m *index.Index
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
