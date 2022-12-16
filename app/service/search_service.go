package service

import (
	"context"
	"github.com/kurakura967/unittest-for-es/app/model"
	"github.com/kurakura967/unittest-for-es/app/repository"
	"log"
)

type SearchService struct {
	repo repository.Searcher
}

func NewSearchService(repo repository.Searcher) *SearchService {
	return &SearchService{repo: repo}
}

func (s *SearchService) GetSearchService(ctx context.Context, keyword, indexName string) (sr []*model.SearchResult, err error) {
	res, err := s.repo.Search(ctx, keyword, indexName)
	if err != nil {
		log.Println(err)
	}
	return res, nil
}
