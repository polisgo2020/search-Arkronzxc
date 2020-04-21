package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/polisgo2020/search-Arkronzxc/config"
	"github.com/polisgo2020/search-Arkronzxc/db"

	"github.com/go-chi/chi"
	"github.com/rs/cors"

	"github.com/polisgo2020/search-Arkronzxc/index"
	"github.com/polisgo2020/search-Arkronzxc/util"
	"github.com/rs/zerolog/log"
)

type searchResponse struct {
	Filename         string `json:"filename"`
	WordsEncountered int    `json:"wordsEncountered"`
}

type service struct {
	idx *index.Index
}

func (s *service) searchHandler(writer http.ResponseWriter, request *http.Request) {

	writer.Header().Set("Content-Type", "application/json")

		log.Info().Str("received", request.FormValue("search")).Msg("got request")

		parsedSearchPhrase, err, errCode := parseSearchPhrase(request)
		if err != nil {
			log.Err(err).Int("status", errCode).Msg("error while parsing search phrase")
			http.Error(writer, http.StatusText(errCode), errCode)
			return
		}
		log.Debug().Strs("parse search phrase", parsedSearchPhrase).Msg("search phrase parsed")

		searchIndex, err := repo.GetIndex(parsedSearchPhrase)

		resp, err, errCode := answerFormation(searchIndex, parsedSearchPhrase)
		if err != nil {
			log.Err(err).Int("status", errCode).Msg("error while creating answer")
			http.Error(writer, http.StatusText(errCode), errCode)
			return
		}
		log.Debug().Interface("response", resp)

		finalJson, err := json.Marshal(resp)
		if err != nil {
			log.Err(err).Msg("error while serializing final JSON")
			http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		log.Debug().Interface("final json", finalJson)

		if _, err := fmt.Fprint(writer, string(finalJson)); err != nil {
			log.Err(err).Msg("error while ")
			http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}

func parseSearchPhrase(request *http.Request) ([]string, error, int) {
	log.Debug().Interface("request", request)

	searchPhrase := request.FormValue("search")
	log.Debug().Str("search phrase", searchPhrase)

	rawUserInput := strings.ToLower(searchPhrase)
	log.Debug().Str("raw user input", rawUserInput)

	parsedUserInput := strings.Split(rawUserInput, " ")
	log.Debug().Strs("parsed user input", parsedUserInput)

	cleanedUserInput := make([]string, 0, len(parsedUserInput))
	for i := range parsedUserInput {
		w, err := util.CleanUserData(parsedUserInput[i])
		if err != nil {
			err = fmt.Errorf("error while cleaning each word in query: %w", err)
			return nil, err, http.StatusBadRequest
		}
		if w != "" {
			cleanedUserInput = append(cleanedUserInput, w)
		}
	}
	log.Debug().Strs("clean user input: ", cleanedUserInput).Msg("user input parsed")
	return cleanedUserInput, nil, -1
}

func answerFormation(index *index.Index, cleanedUserInput []string) ([]*searchResponse, error, int) {
	log.Debug().Interface("index", index).Strs("cleaned user input", cleanedUserInput)

	ans, err := index.BuildSearchIndex(cleanedUserInput)
	if err != nil {
		err = fmt.Errorf("error while building search index with cleaned user input: %w", err)
		return nil, err, http.StatusInternalServerError
	}
	log.Debug().Interface("answer", ans)

	var resp []*searchResponse
	for s := range ans {
		resp = append(resp, &searchResponse{
			Filename:         s,
			WordsEncountered: ans[s],
		})
	}

	log.Debug().Interface("search response", resp).Msg("search response created")
	return resp, nil, -1
}

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)

		log.Debug().
			Str("method", r.Method).
			Str("remote", r.RemoteAddr).
			Str("path", r.URL.Path).
			Int("duration", int(time.Since(start))).
			Msgf("Called url %s", r.URL.Path)
	})
}

func StartingWeb(c *config.Config) error {
	log.Debug().Msg("initialize web application")

	r := chi.NewRouter()
	filesDir := http.Dir("./static")
	s := &service{idx: searchIndex}
	corsPolicy := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	})

	repo, err := db.NewIndexRepository(c)
	if err != nil {
		log.Err(err).Msg("error while initializing db in search handler")
		return nil
	}

	r.Use(corsPolicy.Handler)

	r.Use(logMiddleware)

	r.Get("/api", s.searchHandler)
	err := fileServer(r, "/", filesDir)
	if err != nil {
		return err
	}

	if err := http.ListenAndServe(":"+c.Listen, r); err != nil {
		return err
	}
	return nil
}

func fileServer(r chi.Router, path string, root http.FileSystem) error {
	if strings.ContainsAny(path, "{}*") {
		return fmt.Errorf("fileServer does not permit any URL parameters")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		log.Debug().Interface("request", r).Msg("getting static content")
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
	return nil
}
