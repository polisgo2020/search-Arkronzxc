package web

import (
	"encoding/json"
	"fmt"

	"github.com/go-chi/render"

	"net/http"
	"strings"
	"time"

	"github.com/polisgo2020/search-Arkronzxc/config"

	"github.com/go-chi/chi"
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

	resp, err, errCode := answerFormation(s.idx, parsedSearchPhrase)
	if err != nil {
		log.Err(err).Int("status", errCode).Msg("error while creating answer")
		http.Error(writer, http.StatusText(errCode), errCode)
		return
	}

	finalJson, err := json.Marshal(resp)
	if err != nil {
		log.Err(err).Msg("error while serializing final JSON")
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	log.Debug().
		Strs("parse search phrase", parsedSearchPhrase).
		Interface("resp", resp).
		Interface("final json", finalJson).
		Msg("search phrase parsed")

	if _, err := fmt.Fprint(writer, string(finalJson)); err != nil {
		log.Err(err).Msg("error while writing response")
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return

	}
}

func parseSearchPhrase(request *http.Request) ([]string, error, int) {

	log.Debug().Interface("Request", request)

	searchPhrase := request.FormValue("search")
	log.Debug().Str("Search phrase", searchPhrase)

	rawUserInput := strings.ToLower(searchPhrase)
	log.Debug().Str("Raw user input", rawUserInput)

	parsedUserInput := strings.Split(rawUserInput, " ")
	log.Debug().Strs("Parsed user input", parsedUserInput)

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
	log.Debug().Interface("answer", ans).Msg("answer")

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

func StartingWeb(idx *index.Index, c *config.Config) error {
	s := &service{
		idx: idx,
	}
	r := chi.NewRouter()
	r.Use(logMiddleware)
	r.Route("/api", func(r chi.Router) {
		r.Use(render.SetContentType(render.ContentTypeJSON))
		r.Get("/", s.searchHandler)
	})
	r.Get("/*", func(writer http.ResponseWriter, request *http.Request) {
		h := http.FileServer(http.Dir("./static"))
		h.ServeHTTP(writer, request)
	})

	if err := http.ListenAndServe(c.Listen, r); err != nil {
		log.Err(err)
		return err
	}
	log.Info().Msgf("started to listen at interface %s", c.Listen)
	return nil
}
