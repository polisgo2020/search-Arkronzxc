package main

import (
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
		DefaultText: "8888",
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

	err = app.Run(os.Args)
	if err != nil {
		log.Err(err).Msg("can't initialize console application")
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
		log.Err(err).Str("file names", ctx.String("sources")).Msg("error while reading files")
		return err
	} else {
		invertedIndex, err := index.CreateInvertedIndex(nameSlice)
		if err != nil {
			log.Err(err).Interface("inverted index", invertedIndex).Msg("error while creating inverted index")
			return err
		}
		err = repo.SaveIndex(*invertedIndex)
		if err != nil {
			log.Err(err).Msg("error while saving to database")
			return err
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
