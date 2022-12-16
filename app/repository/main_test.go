package repository_test

import (
	"context"
	"github.com/kurakura967/unittest-for-es/app/model"
	"github.com/olivere/elastic/v7"
	"log"
	"os"
	"strconv"
	"testing"
)

const IndexName = "test_index"

var client *elastic.Client

func connectEs() error {
	var err error
	client, err = elastic.NewClient(elastic.SetSniff(false))
	if err != nil {
		log.Println("failed to connect es")
		return err
	}
	return nil
}

func setupIndex() error {
	mapping := `{
	  "settings": {
		"index": {
		  "refresh_interval": "60s",
		  "auto_expand_replicas": "0-all"
		}
	  },
	  "mappings": {
		"properties": {
		  "author": {
			"type": "text"
		  },
		  "title": {
			"type": "text"
		  }
		}
	  }
	}`

	index, err := client.CreateIndex(IndexName).BodyString(mapping).Do(context.Background())
	if err != nil {
		return err
	}

	if !index.Acknowledged {
		panic("failed to create index")
	}
	return nil
}

func setupDocument() error {
	docs := []*model.SearchResult{
		{
			Author: "William Shakespeare",
			Title:  "Hamlet",
		},
	}

	for i, d := range docs {
		put, err := client.Index().Index(IndexName).OpType("index").Id(strconv.Itoa(i)).BodyJson(d).Do(context.Background())
		if err != nil {
			return err
		}
		log.Printf("Add document to index: %s, type: %s \n", put.Index, put.Type)
	}

	_, err := client.Refresh(IndexName).Do(context.Background())
	if err != nil {
		return err
	}

	return nil
}

func setup() (err error) {
	err = connectEs()
	if err != nil {
		return err
	}
	err = setupIndex()
	if err != nil {
		return err
	}
	err = setupDocument()
	if err != nil {
		return err
	}
	return nil
}

func cleanup() {
	deleteIndex()
	client.Stop()
}

func deleteIndex() error {
	deleteIndex, err := client.DeleteIndex(IndexName).Do(context.Background())
	if err != nil {
		return err
	}
	if !deleteIndex.Acknowledged {
		panic("failed to delete index")
	}
	log.Println("deleted index")
	return nil
}

func TestMain(m *testing.M) {
	err := setup()
	if err != nil {
		os.Exit(1)
	}
	defer cleanup()
	m.Run()
}
