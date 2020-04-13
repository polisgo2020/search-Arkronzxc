package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

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

func SearchHandler(input string) http.HandlerFunc {
	log.Debug().Str("Input", input)

	searchIndex, err := util.UnmarshalFile(input)
	if err != nil {
		panic(err)
	}
	log.Debug().Interface("Index", searchIndex)

	log.Info().Msg("Handler is complete")

	return func(writer http.ResponseWriter, request *http.Request) {

		writer.Header().Set("Content-Type", "application/json")

		log.Info().Str("Received", request.FormValue("search")).Msg("Got request")

		parsedSearchPhrase, err, errCode := parseSearchPhrase(request)
		if err != nil {
			log.Err(err).Int("Status", errCode).Msg("Error while parsing search phrase")
			http.Error(writer, http.StatusText(errCode), errCode)
			return
		}
		log.Debug().Strs("Parse search phrase", parsedSearchPhrase).Msg("Search phrase parsed")

		resp, err, errCode := answerFormation(searchIndex, parsedSearchPhrase)
		if err != nil {
			log.Err(err).Int("Status", errCode).Msg("Error while creating answer")
			http.Error(writer, http.StatusText(errCode), errCode)
			return
		}
		log.Debug().Interface("Resp", resp)

		finalJson, err := json.Marshal(resp)
		if err != nil {
			log.Err(err).Msg("Error while serializing final JSON")
			http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		log.Debug().Interface("Final json", finalJson)

		if _, err := fmt.Fprint(writer, string(finalJson)); err != nil {
			log.Err(err).Msg("Error while ")
			http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
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

	cleanedUserInput := make([]string, 0)
	for i := range parsedUserInput {
		w, err := util.CleanUserData(parsedUserInput[i])
		if err != nil {
			log.Err(err).Msg("Error while cleaning each word in query")
			return nil, err, http.StatusBadRequest
		}
		if w != "" {
			cleanedUserInput = append(cleanedUserInput, w)
		}
	}
	log.Debug().Strs("Clean user input: ", cleanedUserInput).Msg("User input parsed")
	return cleanedUserInput, nil, -1
}

func answerFormation(index *index.Index, cleanedUserInput []string) ([]*searchResponse, error, int) {
	log.Debug().Interface("Index", index).Strs("Cleaned user input", cleanedUserInput)

	ans, err := index.BuildSearchIndex(cleanedUserInput)
	if err != nil {
		log.Err(err).Msg("Error while building search index with cleaned user input")
		return nil, err, http.StatusInternalServerError
	}
	log.Debug().Interface("Ans", ans)

	var resp []*searchResponse
	for s := range ans {
		resp = append(resp, &searchResponse{
			Filename:         s,
			WordsEncountered: ans[s],
		})
		log.Printf("filename: %s, words encountered : %d", s, ans[s])
	}

	log.Debug().Interface("Search response", resp).Msg("Search response created")
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

func StartingWeb(index string, port string) error {
	log.Debug().Msg("Initialize web application")

	r := chi.NewRouter()
	filesDir := http.Dir("./static")

	corsPolicy := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	})

	r.Use(corsPolicy.Handler)

	r.Use(logMiddleware)

	r.Get("/api", SearchHandler(index))
	fileServer(r, "/", filesDir)

	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Err(err).Msg("can't start server")
	}
	return nil
}

func fileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("fileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		log.Debug().Interface("request", r).Msg("Раздаем статику")
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}
