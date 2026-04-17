package xkcd

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"yadro.com/course/update/core"
)

const InfoFileName = "/info.0.json"

type Client struct {
	log    *slog.Logger
	client http.Client
	url    string
}

type xkcdResponse struct {
	Num        int    `json:"num"`
	Img        string `json:"img"`
	Title      string `json:"title"`
	SafeTitle  string `json:"safe_title"`
	Transcript string `json:"transcript"`
	Alt        string `json:"alt"`
}

func NewClient(url string, timeout time.Duration, log *slog.Logger) (*Client, error) {
	if url == "" {
		return nil, fmt.Errorf("empty base url specified")
	}
	return &Client{
		client: http.Client{Timeout: timeout},
		log:    log,
		url:    url,
	}, nil
}

func (c Client) Get(ctx context.Context, id int) (core.XKCDInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url+"/"+strconv.Itoa(id)+InfoFileName, nil)
	if err != nil {
		c.log.Error("xkcd request", "error", err)
		return core.XKCDInfo{}, core.ErrInternalServerError
	}
	resp, err := c.client.Do(req)
	if err != nil {
		c.log.Error("xkcd response", "error", err)
		return core.XKCDInfo{}, core.ErrInternalServerError
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.log.Error("xkcd response body close", "error", closeErr)
		}
	}()
	if resp.StatusCode == http.StatusNotFound {
		c.log.Info("xkcd not found", "id", id)
		return core.XKCDInfo{}, core.ErrNotFound
	}
	info := xkcdResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		c.log.Error("decode error", "error", err)
		return core.XKCDInfo{}, core.ErrInternalServerError
	}
	return core.XKCDInfo{
		ID:          info.Num,
		URL:         info.Img,
		Title:       info.Title,
		SafeTitle:   info.SafeTitle,
		Transcript:  info.Transcript,
		Description: info.Alt,
	}, nil
}

func (c Client) LastID(ctx context.Context) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url+InfoFileName, nil)
	if err != nil {
		c.log.Error("xkcd request", "error", err)
		return 0, core.ErrInternalServerError
	}
	resp, err := c.client.Do(req)
	if err != nil {
		c.log.Error("xkcd response", "error", err)
		return 0, core.ErrInternalServerError
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.log.Error("xkcd response body close", "error", closeErr)
		}
	}()
	if resp.StatusCode == http.StatusNotFound {
		return 0, core.ErrNotFound
	}
	info := xkcdResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		c.log.Error("decode error", "error", err)
		return 0, core.ErrInternalServerError
	}
	return info.Num, nil
}
