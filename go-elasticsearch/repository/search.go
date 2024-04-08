package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

type esHandler struct {
	client *elasticsearch.Client
}

func NewEsHandler(client *elasticsearch.Client) *esHandler {
	return &esHandler{
		client: client,
	}
}

func (e *esHandler) Search(ctx context.Context, index, query string) ([]byte, error) {
	req := esapi.SearchRequest{
		Index: []string{index},
		Body:  strings.NewReader(query),
	}

	res, err := req.Do(ctx, e.client)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		var e errResponse
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, fmt.Errorf("faild to read err response body: %w", err)
		}
		if err := json.Unmarshal(body, &e); err != nil {
			return nil, fmt.Errorf("faild to unmarsal err response body: %w", err)
		}
		return nil, fmt.Errorf("failt to search: [%d] %s", e.Status, e.Error.Cause[0].Reason)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

type errResponse struct {
	Status int      `json:"status"`
	Error  errCause `json:"error"`
}

type errCause struct {
	Cause []errReason `json:"root_cause"`
}

type errReason struct {
	Reason string `json:"reason"`
}
