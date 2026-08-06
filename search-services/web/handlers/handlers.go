package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	modeIndex          = "index"
	endpointSearch     = "/api/search"
	endpointLogin      = "/api/login"
	endpointIndex      = "/api/isearch"
	endpointStatus     = "/api/db/status"
	endpointStats      = "/api/db/stats"
	endpointUpdate     = "/api/db/update"
	endpointDrop       = "/api/db"
	tmplSearch         = "search.html"
	tmplLogin          = "login.html"
	tmplAdmin          = "admin.html"
	redirectAdmin      = "/admin"
	redirectLogin      = "/login"
	cookieToken        = "token"
	defaultSearchLimit = 10
	maxSearchLimit     = 100
)

type Comic struct {
	ID  int    `json:"id"`
	URL string `json:"url"`
}
type SearchResponse struct {
	Comics []Comic `json:"comics"`
	Total  int     `json:"total"`
}

type SearchPageData struct {
	Phrase string
	Mode   string
	Limit  int
	Comics []Comic
	Total  int
}

type LoginRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

type LoginPageData struct {
	Error string
}

type StatusResponse struct {
	Status string `json:"status"`
}

type StatsResponse struct {
	WordsTotal    int `json:"words_total"`
	WordsUnique   int `json:"words_unique"`
	ComicsFetched int `json:"comics_fetched"`
	ComicsTotal   int `json:"comics_total"`
}

type AdminPageData struct {
	Status        string
	WordsTotal    int
	WordsUnique   int
	ComicsFetched int
	ComicsTotal   int
	Message       string
}

func parseSearchLimit(rawLimit string) int {
	if rawLimit == "" {
		return defaultSearchLimit
	}
	limit, err := strconv.Atoi(rawLimit)
	if err != nil || limit <= 0 || limit > maxSearchLimit {
		return defaultSearchLimit
	}
	return limit
}

func flashMessage(key string) string {
	switch key {
	case "drop_ok":
		return "Database cleared"
	case "update_started":
		return "Update started"
	default:
		return ""
	}
}

func fetchStatus(log *slog.Logger, apiAddress string, client *http.Client) (StatusResponse, error) {
	resp, err := client.Get(apiAddress + endpointStatus)
	if err != nil {
		return StatusResponse{}, err
	}
	defer closeBody(log, resp)
	if resp.StatusCode != http.StatusOK {
		return StatusResponse{}, fmt.Errorf("status code %d", resp.StatusCode)
	}
	var result StatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return StatusResponse{}, err
	}
	return result, nil
}

func fetchStats(log *slog.Logger, apiAddress string, client *http.Client) (StatsResponse, error) {
	resp, err := client.Get(apiAddress + endpointStats)
	if err != nil {
		return StatsResponse{}, err
	}
	defer closeBody(log, resp)
	if resp.StatusCode != http.StatusOK {
		return StatsResponse{}, fmt.Errorf("status code %d", resp.StatusCode)
	}
	var result StatsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return StatsResponse{}, err
	}
	return result, nil
}

func NewSearchHandler(tmpl *template.Template, log *slog.Logger, apiAddress string, client *http.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		queryParams := r.URL.Query()
		phrase := queryParams.Get("phrase")
		mode := queryParams.Get("mode")
		limit := parseSearchLimit(queryParams.Get("limit"))

		if phrase == "" {
			pageData := SearchPageData{Limit: defaultSearchLimit}
			if err := tmpl.ExecuteTemplate(w, tmplSearch, pageData); err != nil {
				log.Error("failed to execute template", "error", err)
				http.Error(w, "internal error", http.StatusInternalServerError)
			}
			return
		}

		params := url.Values{}
		params.Set("phrase", phrase)
		params.Set("limit", strconv.Itoa(limit))
		endpoint := endpointSearch
		if mode == modeIndex {
			endpoint = endpointIndex
		}
		apiURL := apiAddress + endpoint + "?" + params.Encode()

		resp, err := client.Get(apiURL)
		if err != nil {
			log.Error("failed to call api", "error", err)
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}
		defer closeBody(log, resp)
		if resp.StatusCode != http.StatusOK {
			log.Error("api returned error", "status", resp.StatusCode)
			http.Error(w, "search failed", http.StatusBadGateway)
			return
		}
		var result SearchResponse
		if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
			log.Error("failed to decode search response", "error", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if err = tmpl.ExecuteTemplate(w, tmplSearch, SearchPageData{
			Phrase: phrase,
			Mode:   mode,
			Limit:  limit,
			Comics: result.Comics,
			Total:  result.Total,
		}); err != nil {
			log.Error("failed to execute template", "error", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	}
}

func NewLoginPageHandler(tmpl *template.Template, log *slog.Logger, apiAddress string, client *http.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			if err := tmpl.ExecuteTemplate(w, tmplLogin, nil); err != nil {
				log.Error("failed to execute template", "error", err)
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			return
		}
		name := r.FormValue("name")
		password := r.FormValue("password")
		if name == "" || password == "" {
			err := tmpl.ExecuteTemplate(w, tmplLogin, LoginPageData{
				Error: "all fields are required",
			})
			if err != nil {
				log.Error("failed to execute template", "error", err)
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			return
		}
		body, err := json.Marshal(LoginRequest{
			Name:     name,
			Password: password,
		})
		if err != nil {
			log.Error("failed to marshal login request", "error", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		resp, err := client.Post(apiAddress+endpointLogin, "application/json", bytes.NewReader(body))
		if err != nil {
			log.Error("failed to call api", "error", err)
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}
		defer closeBody(log, resp)
		if resp.StatusCode != http.StatusOK {
			err = tmpl.ExecuteTemplate(w, tmplLogin, LoginPageData{
				Error: "invalid login or password",
			})
			if err != nil {
				log.Error("failed to execute template", "error", err)
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			return
		}
		tokenBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Error("failed to read response body", "error", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		token := strings.TrimSpace(string(tokenBytes))
		log.Info("login successful", "name", name)
		http.SetCookie(w, &http.Cookie{
			Name:     cookieToken,
			Value:    token,
			HttpOnly: true,
			Path:     "/",
		})
		http.Redirect(w, r, redirectAdmin, http.StatusSeeOther)

	}
}

func requireAuth(w http.ResponseWriter, r *http.Request) (string, bool) {
	cookie, err := r.Cookie(cookieToken)
	if err != nil {
		http.Redirect(w, r, redirectLogin, http.StatusSeeOther)
		return "", false
	}
	return cookie.Value, true
}

func NewAdminHandler(tmpl *template.Template, log *slog.Logger, apiAddress string, client *http.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, ok := requireAuth(w, r)
		if !ok {
			return
		}
		status, err := fetchStatus(log, apiAddress, client)
		if err != nil {
			log.Error("failed to call api status", "error", err)
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}
		stats, err := fetchStats(log, apiAddress, client)
		if err != nil {
			log.Error("failed to call api stats", "error", err)
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}
		if err = tmpl.ExecuteTemplate(w, tmplAdmin, AdminPageData{
			Status:        status.Status,
			WordsTotal:    stats.WordsTotal,
			WordsUnique:   stats.WordsUnique,
			ComicsFetched: stats.ComicsFetched,
			ComicsTotal:   stats.ComicsTotal,
			Message:       flashMessage(r.URL.Query().Get("msg")),
		}); err != nil {
			log.Error("failed to execute template", "error", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	}
}

func NewAdminStatusAPIHandler(log *slog.Logger, apiAddress string, client *http.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAuth(w, r); !ok {
			return
		}
		status, err := fetchStatus(log, apiAddress, client)
		if err != nil {
			log.Error("failed to call api status", "error", err)
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(status); err != nil {
			log.Error("failed to encode status", "error", err)
		}
	}
}

func NewAdminStatsAPIHandler(log *slog.Logger, apiAddress string, client *http.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAuth(w, r); !ok {
			return
		}
		stats, err := fetchStats(log,apiAddress, client)
		if err != nil {
			log.Error("failed to call api stats", "error", err)
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(stats); err != nil {
			log.Error("failed to encode stats", "error", err)
		}
	}
}

func NewUpdateHandler(log *slog.Logger, apiAddress string, client *http.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, ok := requireAuth(w, r)
		if !ok {
			return
		}
		log.Info("update requested")
		apiURL := apiAddress + endpointUpdate
		req, err := http.NewRequest(http.MethodPost, apiURL, nil)
		if err != nil {
			log.Error("failed to call api", "error", err)
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}
		req.Header.Set("Authorization", "Token "+token)
		resp, err := client.Do(req)
		if err != nil {
			log.Error("failed to call api", "error", err)
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}
		defer closeBody(log, resp)
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
			log.Error("api returned error", "status", resp.StatusCode)
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}
		log.Info("update started successfully")
		http.Redirect(w, r, redirectAdmin+"?msg=update_started", http.StatusSeeOther)
	}
}

func NewDropHandler(log *slog.Logger, apiAddress string, client *http.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, ok := requireAuth(w, r)
		if !ok {
			return
		}
		log.Info("drop requested")
		apiURL := apiAddress + endpointDrop
		req, err := http.NewRequest(http.MethodDelete, apiURL, nil)
		if err != nil {
			log.Error("failed to call api", "error", err)
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}
		req.Header.Set("Authorization", "Token "+token)
		resp, err := client.Do(req)
		if err != nil {
			log.Error("failed to call api", "error", err)
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}
		defer closeBody(log, resp)
		if resp.StatusCode != http.StatusOK {
			log.Error("api returned error", "status", resp.StatusCode)
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}
		log.Info("drop completed successfully")
		http.Redirect(w, r, redirectAdmin+"?msg=drop_ok", http.StatusSeeOther)
	}
}

func closeBody(log *slog.Logger, resp *http.Response) {
    if err := resp.Body.Close(); err != nil {
        log.Error("failed to close response body", "error", err)
    }
}