package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	finalDataFile        = "output\\final.csv"
	finalOutputDirectory = "output"
)

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
			_, ok := m[k]
			if !ok {
				m[k] = make([]string, 0)
			}
			m[k] = append(m[k], v)
		}
	}
	return CreateOutputCSV(m)
}

//IDK why that func doesn't work in win10
func CreateOutputCSV(m map[string][]string) error {
	if err := os.MkdirAll(finalOutputDirectory, 0777); err != nil {
		return err
	}
	recordFile, _ := os.Create(finalDataFile)
	w := csv.NewWriter(recordFile)
	for k, v := range m {
		t, err := json.Marshal(v)
		if err != nil {
			fmt.Printf("error %e while creating json from obj %+v \n", err, &v)
		}
		fmt.Printf("key: %s, value: %s \n", k, t)
		err = w.Write([]string{fmt.Sprintf("%s", k), fmt.Sprintf("%s", t)})
		if err != nil {
			fmt.Printf("error %e while saving record %s,%s \n", err, k, t)
		}
	}
	return nil
}
