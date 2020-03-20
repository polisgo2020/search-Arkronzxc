package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"
)

var finalDataFile = path.Join(finalOutputDir, "output.json")

const finalOutputDir = "output"

// Returns slice of file names from dir
func ReadFileName(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// returns map where key is a word in file, value is filename
func ReadFiles(files string) (map[string]string, error) {
	m := make(map[string]string)

	content, err := ioutil.ReadFile(files)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	//help me to resolve this kostil'
	t, _ := utf8.DecodeRune([]byte("'"))
	f := func(c rune) bool {
		return !unicode.IsLetter(c) && t != c
	}
	str := strings.FieldsFunc(string(content), f)
	for _, v := range str {
		m[v] = files
	}
	return m, nil
}

func CreateInvertedIndex(filepath string) error {
	m := make(map[string][]string)
	files, err := ReadFileName(filepath)

	if err != nil {
		log.Print(err, "No such file")
		return err
	}
	for _, filename := range files {
		wordMap, err := ReadFiles(filename)
		if err != nil {
			log.Print(err, "Error while parsing file occurred")
			continue
		}
		for k, v := range wordMap {
			if _, ok := m[k]; !ok {
				m[k] = make([]string, 0)
			}
			m[k] = append(m[k], v)
		}
	}
	return CreateOutputJSON(m)
}

func CreateOutputJSON(m map[string][]string) error {
	if err := os.MkdirAll(finalOutputDir, 0777); err != nil {
		return err
	}
	recordFile, _ := os.Create(finalDataFile)
	data, err := json.Marshal(m)
	if err != nil {
		log.Print(err)
		return err
	}
	_, err = recordFile.Write(data)

	return err
}
