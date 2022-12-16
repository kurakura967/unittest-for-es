package repository

import (
	"context"
	"github.com/kurakura967/unittest-for-es/app/model"
)

type Searcher interface {
	Search(ctx context.Context, keyword, indexName string) ([]*model.SearchResult, error)
}
