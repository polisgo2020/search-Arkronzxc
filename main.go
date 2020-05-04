package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/polisgo2020/search-Arkronzxc/config"
	"github.com/polisgo2020/search-Arkronzxc/web"

	"github.com/rs/zerolog"

	"github.com/polisgo2020/search-Arkronzxc/index"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

func main() {

	var err error

	if err = initLogger(config.Load()); err != nil {
		log.Err(err).Msg("can not init logger")
		return
	}

	app := cli.NewApp()
	app.Name = "Search index"
	app.Usage = "generate index from text files and search over them"

	indexFileFlag := &cli.StringFlag{
		Aliases:  []string{"i"},
		Name:     "index",
		Usage:    "Index file",
		Required: true,
	}

	sourcesFlag := &cli.StringFlag{
		Aliases:  []string{"s"},
		Name:     "sources, s",
		Usage:    "Files to index",
		Required: true,
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
			},
			Action: search,
		},
	}

	err = app.Run(os.Args)
	if err != nil {
		log.Err(err)
	}
}

func initLogger(c *config.Config) error {
	logLvl, err := zerolog.ParseLevel(c.LogLevel)
	if err != nil {
		return err
	}
	zerolog.SetGlobalLevel(logLvl)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	return nil
}

func build(ctx *cli.Context) error {

	log.Info().Msg("Build option chosen")

	log.Debug().
		Str("files to index in dir", ctx.String("sources")).
		Str("output file with index", ctx.String("index")).
		Msg("build option")

	if nameSlice, err := readFileNames(ctx.String("sources")); err != nil {
		return fmt.Errorf("error while reading file names: %w", err)
	} else {
		invertedIndex, err := index.CreateInvertedIndex(nameSlice)
		if err != nil {
			return fmt.Errorf("error while creating inverted index: %w", err)
		}
		if err = createOutputJSON(invertedIndex, ctx.String("index")); err != nil {
			return fmt.Errorf("error while creating output json: %w", err)
		}
	}

	log.Debug().Msg("build successfully completed")
	return nil
}

func search(ctx *cli.Context) error {
	c := config.Load()

	input := ctx.String("index")
	log.Info().Msg("starting searching")

	log.Debug().Str("input", input)

	searchIndex, err := index.UnmarshalFile(input)
	if err != nil {
		return err
	}
	log.Debug().Interface("index", searchIndex)

	log.Info().Msg("handler is complete")

	return web.StartingWeb(searchIndex, c)

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
		log.Err(err).Msg("Error while recording file")
		return err
	}
	data, err := json.Marshal(m)
	if err != nil {
		log.Err(err).Msg("Error when initializing output JSON")
		return err
	}

	_, err = recordFile.Write(data)

	log.Debug().Msg("File is read")
	return err
}
