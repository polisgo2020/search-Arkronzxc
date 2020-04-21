package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/polisgo2020/search-Arkronzxc/db"

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
		log.Err(err).Msg("can't init logger")
		return
	}

	app := cli.NewApp()
	app.Name = "Search index"
	app.Usage = "generate index from text files and search over them"

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
				sourcesFlag,
			},
			Action: build,
		},
		{
			Name:    "search",
			Aliases: []string{"s"},
			Usage:   "Search over the index",
			Flags:   []cli.Flag{},
			Action:  search,
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
	repo, err := db.NewIndexRepository(config.Load())
	if err != nil {
		return err
	}

	log.Info().Msg("build option chosen")

	log.Debug().
		Str("files to index", ctx.String("sources")).
		Str("index file", ctx.String("index")).
		Msg("build option")

	if nameSlice, err := readFileNames(ctx.String("sources")); err != nil {
		return fmt.Errorf("error while reading file names: %w", err)
	} else {
		invertedIndex, err := index.CreateInvertedIndex(nameSlice)
		if err != nil {
			return fmt.Errorf("error while creating inverted index: %w", err)
		}
		if err = repo.SaveIndex(*invertedIndex); err != nil {
			return fmt.Errorf("error while creating output json: %w", err)
		}
	}

	log.Debug().Msg("build successfully completed")
	return nil
}

func search(ctx *cli.Context) error {
	log.Info().Msg("starting searching")

	err := web.StartingWeb(config.Load())
	if err != nil {
		log.Err(err)
		return err
	}
	return nil
}

// Returns slice of file names from dir
func readFileNames(root string) ([]string, error) {
	log.Debug().Str("root", root)
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	log.Debug().Strs("files", files)
	return files, err
}
