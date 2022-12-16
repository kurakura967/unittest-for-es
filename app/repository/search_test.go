package repository_test

import (
	"context"
	"github.com/google/go-cmp/cmp"
	"github.com/kurakura967/unittest-for-es/app/model"
	"github.com/kurakura967/unittest-for-es/app/repository"
	"testing"
)

func TestSearch(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		testTitle string
		keyword   string
		expected  []*model.SearchResult
	}{
		{
			testTitle: "searchTest1",
			keyword:   "Hamlet",
			expected: []*model.SearchResult{
				{
					Author: "William Shakespeare",
					Title:  "Hamlet",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testTitle, func(t *testing.T) {
			repo := repository.NewEsHandler(client)
			got, err := repo.Search(ctx, tt.keyword, IndexName)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tt.expected, got); diff != "" {
				t.Errorf("SearchResults is unmatched (-want, +got): %s\n", diff)
			}
		})
	}

}
