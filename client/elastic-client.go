package client

import (
	"context"
	"github.com/olivere/elastic"
)

const (
	IndexName = "files"
)

type File struct {
	Name	string	`json:"name"`
	Description	string	`json:"description"`
}

func NewElasticClient() (*elastic.Client, error){
	client, err := elastic.NewClient(elastic.SetURL("http://elastic:9200"))
	if err != nil{
		return client, err
	}

	//check if it hasn't been created yet
	exists, err := client.IndexExists(IndexName).Do(context.Background())
	if err != nil{
		return client, err
	}
	if !exists{
		_, err = client.CreateIndex(IndexName).Do(context.Background())
		if err != nil{
			return client, err
		}
	}

	return client, nil
}
