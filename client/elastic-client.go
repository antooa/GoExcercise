// package client provides a new ES Client and CRUD operations for it. Package uses olivere/elastic client.
package client

import (
	"context"
	"encoding/json"
	"github.com/olivere/elastic"
)

const (
	IndexName = "files"
)

type File struct {
	Name        string `json:"name"`
	Url         string `json:"url"`
	Description string `json:"description"`
}

type ElasticStorage struct {
	*elastic.Client
}

// NewElasticClient provides a new elastic.Client allocation and creation of the Index in ES.
//
// Returns a pointer to the created Client and an error
func NewElasticClient() (*ElasticStorage, error) {
	client, err := elastic.NewClient(elastic.SetURL("http://elastic:9200"))
	if err != nil {
		return nil, err
	}
	backend := &ElasticStorage{client}

	//check if Index hasn't been created yet
	exists, err := backend.Client.IndexExists(IndexName).Do(context.Background())
	if err != nil {
		return nil, err
	}
	if !exists {
		_, err = backend.Client.CreateIndex(IndexName).Do(context.Background())
	}

	return backend, err
}

// Create provides a doc creation in the Index using elastic.Client
//
// Returns an Id of the doc and an error
func (storage *ElasticStorage) Create(file File) (string, error) {

	res, err := storage.Client.Index().
		Index(IndexName).
		Type("doc").
		BodyJson(file).
		Refresh("wait_for").
		Do(context.Background())

	return res.Id, err
}

// Read provides an ability to retrieve a doc from Index by Id
//
// Returns the doc and an error
func (storage *ElasticStorage) Read(id string) (File, error) {

	var file File

	res, err := storage.Client.Get().
		Index(IndexName).
		Type("doc").
		Id(id).
		Do(context.Background())
	if err != nil {
		return file, err
	}

	if res.Found != true {
		return file, err
	}

	err = json.Unmarshal(*res.Source, &file)
	return file, err
}

// Delete provides an ability to delete a doc by Id
//
// Returns an error
func (storage *ElasticStorage) Delete(id string) error {

	_, err := storage.Client.Delete().
		Index(IndexName).
		Type("doc").
		Id(id).
		Do(context.Background())

	return err

}

// Update provides an ability to update a doc using the fields of the newFile
//
// Returns an error
func (storage *ElasticStorage) Update(id string, newFile File) error {

	_, err := storage.Client.Update().
		Index(IndexName).
		Type("doc").
		Id(id).
		Doc(map[string]interface{}{
			"name":        newFile.Name,
			"url":         newFile.Url,
			"description": newFile.Description,
		}).DetectNoop(false).
		Do(context.Background())

	return err
}
