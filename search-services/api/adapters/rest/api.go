package rest

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"yadro.com/course/api/core"
)

type PingResponse struct {
	Replies map[string]string `json:"replies"`
}

type WordsResponse struct {
	Words []string `json:"words"`
	Total int      `json:"total"`
}

type UpdateStatusResponse struct {
	Status string `json:"status"`
}

type UpdateStatsResponse struct {
	WordsTotal    int `json:"words_total"`
	WordsUnique   int `json:"words_unique"`
	ComicsFetched int `json:"comics_fetched"`
	ComicsTotal   int `json:"comics_total"`
}

type Comic struct {
	ID  int    `json:"id"`
	URL string `json:"url"`
}

type SearchResponse struct {
	Comics []Comic `json:"comics"`
	Total  int     `json:"total"`
}

const (
	defaultLimit = 10
)

func toHTTPStatus(err error) int {
	switch {
	case errors.Is(err, core.ErrTooLarge), errors.Is(err, core.ErrBadArguments):
		return http.StatusBadRequest
	case errors.Is(err, core.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, core.ErrAlreadyExists):
		return http.StatusAccepted
	default:
		return http.StatusInternalServerError
	}
}

func NewPingHandler(log *slog.Logger, pingers map[string]core.Pinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		replies := map[string]string{}
		for name, pinger := range pingers {
			if err := pinger.Ping(r.Context()); err != nil {
				replies[name] = "unavailable"
				continue
			}
			replies[name] = "ok"
		}
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(PingResponse{
			Replies: replies,
		})
		if err != nil {
			log.Error("Failed to encode response", "error", err)
		}
	}
}

func NewSearchIndexHandler(log *slog.Logger, searcher core.Searcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var limit int
		params := r.URL.Query()
		phrase := params.Get("phrase")
		queryLimit := params.Get("limit")
		if queryLimit == "" {
			limit = defaultLimit
		} else {
			limit, err = strconv.Atoi(queryLimit)
			if err != nil {
				log.Error("Failed to parse query parameter 'limit'", "error", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if limit <= 0 {
				log.Error("Invalid query parameter 'limit'", "error", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}
		if phrase == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		searched, err := searcher.SearchIndex(r.Context(), phrase, limit)
		if err != nil {
			log.Error("Failed to search", "error", err)
			w.WriteHeader(toHTTPStatus(err))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		comics := make([]Comic, 0, len(searched))
		for _, data := range searched {
			comics = append(comics, Comic{ID: data.ID, URL: data.URL})
		}
		err = json.NewEncoder(w).Encode(SearchResponse{
			Comics: comics,
			Total:  len(searched),
		})
		if err != nil {
			log.Error("Failed to encode response", "error", err)
		}
	}
}

func NewSearchHandler(log *slog.Logger, searcher core.Searcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var limit int
		params := r.URL.Query()
		phrase := params.Get("phrase")
		queryLimit := params.Get("limit")
		if queryLimit == "" {
			limit = defaultLimit
		} else {
			limit, err = strconv.Atoi(queryLimit)
			if err != nil {
				log.Error("Failed to parse query parameter 'limit'", "error", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if limit <= 0 {
				log.Error("Invalid query parameter 'limit'", "error", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}
		if phrase == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		searched, err := searcher.Search(r.Context(), phrase, limit)
		if err != nil {
			log.Error("Failed to search", "error", err)
			w.WriteHeader(toHTTPStatus(err))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		comics := make([]Comic, 0, len(searched))
		for _, data := range searched {
			comics = append(comics, Comic{ID: data.ID, URL: data.URL})
		}
		err = json.NewEncoder(w).Encode(SearchResponse{
			Comics: comics,
			Total:  len(searched),
		})
		if err != nil {
			log.Error("Failed to encode response", "error", err)
		}
	}
}

func NewWordsHandler(log *slog.Logger, normalizer core.Normalizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		normPhrase, err := normalizer.Norm(r.Context(), r.URL.Query().Get("phrase"))
		if err != nil {
			log.Error("Failed to normalize", "error", err)
			w.WriteHeader(toHTTPStatus(err))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(WordsResponse{
			Words: normPhrase,
			Total: len(normPhrase),
		})
		if err != nil {
			log.Error("Failed to encode response", "error", err)
		}
	}
}

func NewUpdateHandler(log *slog.Logger, updater core.Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := updater.Update(r.Context()); err != nil {
			log.Error("Failed to update", "error", err)
			w.WriteHeader(toHTTPStatus(err))
			return
		}
	}
}

func NewUpdateStatsHandler(log *slog.Logger, updater core.Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats, err := updater.Stats(r.Context())
		if err != nil {
			log.Error("Failed to get stats", "error", err)
			w.WriteHeader(toHTTPStatus(err))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err = json.NewEncoder(w).Encode(UpdateStatsResponse{
			WordsTotal:    stats.WordsTotal,
			WordsUnique:   stats.WordsUnique,
			ComicsFetched: stats.ComicsFetched,
			ComicsTotal:   stats.ComicsTotal,
		}); err != nil {
			log.Error("Failed to encode response", "error", err)
		}
	}
}

func NewUpdateStatusHandler(log *slog.Logger, updater core.Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status, err := updater.Status(r.Context())
		if err != nil {
			log.Error("Failed to get stats", "error", err)
			w.WriteHeader(toHTTPStatus(err))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err = json.NewEncoder(w).Encode(UpdateStatusResponse{
			Status: string(status),
		}); err != nil {
			log.Error("Failed to encode response", "error", err)
		}
	}
}

func NewDropHandler(log *slog.Logger, updater core.Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := updater.Drop(r.Context()); err != nil {
			log.Error("Failed to drop", "error", err)
			w.WriteHeader(toHTTPStatus(err))
			return
		}
	}
}
