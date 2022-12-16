package testdata

import (
	"context"
	"github.com/kurakura967/unittest-for-es/app/model"
)

type EsMockHandler struct{}

func (e EsMockHandler) Search(ctx context.Context, keyword, indexName string) ([]*model.SearchResult, error) {
	return []*model.SearchResult{
		{
			Author: "William Shakespeare",
			Title:  "Hamlet",
		},
	}, nil
}
