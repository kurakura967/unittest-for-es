package service_test

import (
	"context"
	"github.com/google/go-cmp/cmp"
	"github.com/kurakura967/unittest-for-es/app/model"
	"github.com/kurakura967/unittest-for-es/app/service"
	"github.com/kurakura967/unittest-for-es/app/service/testdata"
	"testing"
)

func TestSearchService(t *testing.T) {
	ctx := context.Background()
	ser := service.NewSearchService(testdata.EsMockHandler{})

	tests := []struct {
		testTitle string
		keyword   string
		expected  []*model.SearchResult
	}{
		{
			testTitle: "searchServiceTest1",
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
		got, err := ser.GetSearchService(ctx, tt.keyword, "test_index")
		if err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(tt.expected, got); diff != "" {
			t.Errorf("SearchResults is unmatched (-want, +got): %s\n", diff)
		}
	}

}
