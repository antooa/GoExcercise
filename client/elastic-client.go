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
// NewElasticClient provides a new elastic.Client allocation and creation of the Index in ES.
//
// Returns a pointer to the created Client and an error
func NewElasticClient() (*elastic.Client, error) {
	client, err := elastic.NewClient(elastic.SetURL("http://elastic:9200"))
	if err != nil {
		return client, err
	}

	//check if Index hasn't been created yet
	exists, err := client.IndexExists(IndexName).Do(context.Background())
	if err != nil {
		return client, err
	}
	if !exists {
		_, err = client.CreateIndex(IndexName).Do(context.Background())
	}

	return client, err
}

// CreateDoc provides a doc creation in the Index using elastic.Client
//
// Returns an Id of the doc and an error
func CreateDoc(client *elastic.Client, file File) (string , error) {

	res, err := client.Index().
		Index(IndexName).
		Type("doc").
		BodyJson(file).
		Refresh("wait_for").
		Do(context.Background())

	return res.Id, err
}

// ReadDoc provides an ability to retrieve a doc from Index by Id
//
// Returns the doc and an error
func ReadDoc(client *elastic.Client, id string) (File, error) {

	var file File

	res, err := client.Get().
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

// DeleteDoc provides an ability to delete a doc by Id
//
// Returns an error
func DeleteDoc(client *elastic.Client, id string)  error{

	_, err := client.Delete().
		Index(IndexName).
		Type("doc").
		Id(id).
		Do(context.Background())

	return err

}

// UpdateDoc provides an ability to update a doc using the fields of the newFile
//
// Returns an error
func UpdateDoc(client *elastic.Client, id string, newFile File)  error{

	_, err := client.Update().
		Index(IndexName).
		Type("doc").
		Id(id).
		Doc(map[string]interface{}{"name": newFile.Name}).
		DetectNoop(false).
		Do(context.Background())
	if err != nil{
		return err
	}

	_, err = client.Update().
		Index(IndexName).
		Type("doc").
		Id(id).
		Doc(map[string]interface{}{"url": newFile.Url}).
		DetectNoop(false).
		Do(context.Background())
	if err != nil{
		return err
	}

	_, err = client.Update().
		Index(IndexName).
		Type("doc").
		Id(id).
		Doc(map[string]interface{}{"description": newFile.Description}).
		DetectNoop(false).
		Do(context.Background())

	return err
}
