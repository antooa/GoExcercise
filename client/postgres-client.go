package client

import (
	"GoExcercise/handler"
	"database/sql"
	_ "github.com/lib/pq"
)

type PostgresStorage struct {
	*sql.DB
}

func NewPostgresClient(connStr string) (*PostgresStorage, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	storage := &PostgresStorage{db}

	_, err = storage.Exec(`CREATE TABLE IF NOT EXISTS files (
  		  		id SERIAL PRIMARY KEY,
 		  		name varchar(10),
 		  		url TEXT,
 		  		description VARCHAR(255));`)
	if err != nil {
		return nil, err
	}

	return storage, err
}

func (storage *PostgresStorage) Create(file handler.File) (id string, err error)  {
	row := storage.QueryRow("INSERT INTO files (name, url, description) VALUES ($1, $2, $3) RETURNING id",
		file.Name,
		file.Url,
		file.Description)
	err = row.Scan(&id)

	return id, err
}

func (storage *PostgresStorage) Read(id string) (file handler.File, err error) {
	row := storage.QueryRow("SELECT name, url, description FROM files WHERE id=$1", id)
	err = row.Scan(&file.Name, &file.Url, &file.Description)

	return file, err
}

func (storage *PostgresStorage) Update(id string, newFile handler.File) error {
	_, err := storage.Exec("UPDATE files SET name = $1, url = $2, description = $3 WHERE id = $4",
		newFile.Name, newFile.Url, newFile.Description, id)

	return err
}

func (storage *PostgresStorage) Delete(id string) error  {
	_, err := storage.Exec("DELETE FROM files WHERE id = $1", id)

	return err
}