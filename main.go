package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/middleware"

	"github.com/go-chi/chi"
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

	portFlag := &cli.StringFlag{
		Aliases:     []string{"p"},
		Name:        "port",
		Usage:       "Network interface",
		DefaultText: "80",
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
				portFlag,
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

type searchResponse struct {
	Filename    string `json:"filename"`
	WordCounter int    `json:"wordCounter"`
}

func search(ctx *cli.Context) error {

	log.Println("starting")
	r := chi.NewRouter()

	file, err := unmarshalFile(ctx.String("index"))
	if err != nil {
		log.Print(err)
		return err
	}

	r.Use(middleware.DefaultLogger)

	r.Get("/", func(writer http.ResponseWriter, request *http.Request) {
		log.Println("request")
		searchPhrase := request.FormValue("search")

		rawUserInput := strings.ToLower(searchPhrase)
		parsedUserInput := strings.Split(rawUserInput, " ")
		cleanedUserInput := make([]string, 0)
		for i := range parsedUserInput {
			w, err := util.CleanUserData(parsedUserInput[i])
			if err != nil {
				log.Print(err)
				http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			if w != "" {
				cleanedUserInput = append(cleanedUserInput, w)
			}
		}
		log.Println("cleanUserInput: ", cleanedUserInput)
		ans, err := index.BuildSearchIndex(cleanedUserInput, file)
		if err != nil {
			log.Print(err)
			http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		var resp []*searchResponse
		for s := range ans {
			resp = append(resp, &searchResponse{
				Filename:    s,
				WordCounter: ans[s],
			})
			log.Printf("filename: %s, frequency : %d", s, ans[s])
		}
		finalJson, err := json.Marshal(resp)
		if err != nil {
			log.Print(err)
			http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

		writer.Header().Set("Content-Type", "application/json")

		if _, err := fmt.Fprint(writer, string(finalJson)); err != nil {

		}

	})

	if err = http.ListenAndServe(":"+ctx.String("port"), r); err != nil {
		log.Println("error", err)
	}

	return nil
}

// Returns slice of file names from dir
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
