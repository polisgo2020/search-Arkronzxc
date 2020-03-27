package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

func readAndParseFile(filename string) (map[string][]string, error) {
	var m map[string][]string
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
