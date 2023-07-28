package main

import (
	"sync"
	"errors"
	"os"
	"encoding/json"
)

type DB struct {
	Path string
	Mux *sync.RWMutex
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
}

type Chirp struct {
	Id int `json:"id"`
	Body string `json:"body"`
}

func NewDB(path string) (*DB, error) {
	database := DB{Path: path}
	err := database.ensureDB()
	if err != nil {
		return nil, err
	}
	return &database, nil
}

// func (db *DB) CreateChirp(body string) (Chirp, error) {

// }

// func (db *DB) getChirps() ([]Chirp, error) {

// }

func (db *DB) ensureDB() error {
	_, err := os.ReadFile(db.Path)
		if err != nil {
			// If file does not exist, create file
			newFile, err := os.Create(db.Path)
			if err != nil {
			// If file creation fails, return error
			return errors.New("There was an error creating the Chirps database")
			} else {
				newFile.Close()
				return nil
		  }
	  }
	return nil
}

func (db *DB) loadDB() (DBStructure, error) {
	dbs := DBStructure{}
	file, readErr := os.ReadFile(db.Path)
	if readErr != nil {
		return dbs, readErr
	}

	err := json.Unmarshal(file, &dbs)
	if err != nil {
		return dbs, readErr
	} 
	return dbs, nil
}

// func (db *DB) writeDB(dbStructure DBStructure) error {
// }
