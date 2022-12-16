package repository

import (
	"context"
	"encoding/json"
	"github.com/kurakura967/unittest-for-es/app/model"
	"github.com/olivere/elastic/v7"
)

type EsHandler struct {
	client *elastic.Client
}

func NewEsHandler(client *elastic.Client) *EsHandler {
	return &EsHandler{client: client}
}

func (e *EsHandler) Search(ctx context.Context, keyword, indexName string) ([]*model.SearchResult, error) {
	termQuery := elastic.NewMatchPhraseQuery("title", keyword)
	res, err := e.client.Search().
		Index(indexName).
		Query(termQuery).
		From(0).Size(10).
		Pretty(true).
		Do(ctx)

	if err != nil {
		return nil, err
	}

	searchArray := make([]*model.SearchResult, 0)

	if res.TotalHits() > 0 {

		for _, hit := range res.Hits.Hits {

			var searchResult model.SearchResult
			if err := json.Unmarshal(hit.Source, &searchResult); err != nil {
				return nil, err
			}
			searchArray = append(searchArray, &searchResult)
		}
	} else {
		return searchArray, nil
	}
	return searchArray, nil
}
