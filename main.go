package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/polisgo2020/search-Arkronzxc/index"
	"github.com/polisgo2020/search-Arkronzxc/util"
	"github.com/urfave/cli/v2"
)

func main() {
	app := cli.NewApp()
	app.Name = "Search index"
	app.Usage = "generate index from text files and search over them"

	indexFileFlag := &cli.StringFlag{
		Aliases: []string{"i"},
		Name:    "index",
		Usage:   "Index file",
	}

	sourcesFlag := &cli.StringFlag{
		Aliases: []string{"s"},
		Name:    "sources, s",
		Usage:   "Files to index",
	}

	searchFlag := &cli.StringFlag{
		Aliases: []string{"sw"},
		Name:    "search-word, sw",
		Usage:   "Search words",
	}

	app.Commands = []*cli.Command{
		{
			Name:    "build",
			Aliases: []string{"b"},
			Usage:   "Build search index",
			Flags: []cli.Flag{
				indexFileFlag,
				sourcesFlag,
			},
			Action: build,
		},
		{
			Name:    "search",
			Aliases: []string{"s"},
			Usage:   "Search over the index",
			Flags: []cli.Flag{
				indexFileFlag,
				searchFlag,
			},
			Action: search,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func build(ctx *cli.Context) error {
	if nameSlice, err := readFileNames(ctx.String("sources")); err != nil {
		log.Print(err)
		return err
	} else {
		invertedIndex, err := index.CreateInvertedIndex(nameSlice)
		if err != nil {
			log.Print(err)
			return err
		}
		if err = createOutputJSON(invertedIndex, ctx.String("index")); err != nil {
			log.Print(err)
			return err
		}
	}
	return nil
}

func search(ctx *cli.Context) error {
	file, err := unmarshalFile(ctx.String("index"))
	if err != nil {
		log.Print(err)
		return err
	}
	rawUserInput := strings.ToLower(ctx.String("search"))
	parsedUserInput := strings.Split(rawUserInput, ",")
	cleanedUserInput := make([]string, 0)
	for i := range parsedUserInput {
		w, err := util.CleanUserData(parsedUserInput[i])
		if err != nil {
			log.Print(err)
			return nil
		}
		if w != "" {
			parsedUserInput = append(parsedUserInput, w)
		}
	}
	ans, err := index.BuildSearchIndex(cleanedUserInput, file)
	if err != nil {
		log.Print(err)
		return err
	}
	for s := range ans {
		log.Printf("filename: %s, frequency : %d", s, ans[s])
	}
	return nil
}

//Returns slice of file names from dir
func readFileNames(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func createOutputJSON(m *index.Index, outputFileName string) error {
	recordFile, err := os.Create(outputFileName)
	if err != nil {
		log.Print(err)
		return err
	}
	data, err := json.Marshal(m)
	if err != nil {
		log.Print(err)
		return err
	}
	_, err = recordFile.Write(data)

	return err
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
