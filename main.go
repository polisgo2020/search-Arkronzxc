package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/polisgo2020/search-Arkronzxc/index"

	//"github.com/rs/zerolog/log"
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
		DefaultText: "8080",
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

func search(ctx *cli.Context) error {

	log.Println("search starting")

	r := chi.NewRouter()

	r.Use(middleware.DefaultLogger)

	r.Get("/", index.SearchHandler(ctx.String("index")))

	if err := http.ListenAndServe(":"+ctx.String("port"), r); err != nil {
		log.Print("error", err)
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
