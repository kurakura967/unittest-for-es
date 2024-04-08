package repository

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/elastic/go-elasticsearch/v8"
)

type mockTransport struct {
	response *http.Response
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.response, nil
}

func TestSearch(t *testing.T) {

	body := `{
	 "took" : 64,
	 "timed_out" : false,
	 "_shards" : {
		"total" : 1,
		"successful" : 1,
		"skipped" : 0,
		"failed" : 0
	 },
	 "hits" : {
		"total" : {
		  "value" : 1,
		  "relation" : "eq"
		},
		"max_score" : 1.0,
		"hits" : [
		  {
			"_index" : "test_index",
			"_id" : "1",
			"_score" : 1.0,
			"_source" : {
			  "title" : "hogehoge"
			}
		  }
		]
	 }
	}`

	//errBody := `{
	//  "error" : {
	//	"root_cause" : [
	//	  {
	//		"type" : "illegal_argument_exception",
	//		"reason" : "no mapping found for furigna in order to collapse on"
	//	  }
	//	],
	//	"type" : "search_phase_execution_exception",
	//	"reason" : "all shards failed",
	//	"phase" : "query",
	//	"grouped" : true,
	//	"failed_shards" : [
	//	  {
	//		"shard" : 0,
	//		"index" : "kurasawatest_debug_v3_20240131",
	//		"node" : "IaE5IhL8R92vM7-lfqiy9A",
	//		"reason" : {
	//		  "type" : "illegal_argument_exception",
	//		  "reason" : "no mapping found for furigna in order to collapse on"
	//		}
	//	  }
	//	],
	//	"caused_by" : {
	//	  "type" : "illegal_argument_exception",
	//	  "reason" : "no mapping found for furigna in order to collapse on",
	//	  "caused_by" : {
	//		"type" : "illegal_argument_exception",
	//		"reason" : "no mapping found for furigna in order to collapse on"
	//	  }
	//	}
	//  },
	//  "status" : 400
	//}`

	mockTrans := &mockTransport{
		response: &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     http.Header{"X-Elastic-Product": []string{"Elasticsearch"}},
		},
	}

	cfg := elasticsearch.Config{
		Transport: mockTrans,
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	es := NewEsHandler(client)
	res, err := es.Search(context.Background(), "test", "")
	if err != nil {
		t.Fatal(err)
	}
	var r esResponse
	if err := json.Unmarshal(res, &r); err != nil {
		t.Fatal(err)
	}

	if r.Hits.Hits[0].Source.Title != "hogehoge" {
		t.Errorf("unexpected result: %s", r.Hits.Hits[0].Source.Title)
	}
}

type esResponse struct {
	Hits esHitsOuter `json:"hits"`
}

type esHitsOuter struct {
	Hits []esHitsInner `json:"hits"`
}

type esHitsInner struct {
	Source esSource `json:"_source"`
}

type esSource struct {
	Title string `json:"title"`
}
