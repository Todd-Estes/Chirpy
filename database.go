package main

import (
	"sync"
	"errors"
	"os"
	"encoding/json"
	"fmt"
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
  // Check if database is empty; if so, intstantiate new map and assign it to dbs struct
	file, error := os.Stat(db.Path)
	if error != nil {
		return dbs, error
	}
	if file.Size() == 0 {
		chirpMap := make(map[int]Chirp)
		dbs.Chirps = chirpMap
		return dbs, nil
	}
	
	// We should get here if database is not empty; we unmarshall data into dbs struct
	fileContents, readErr := os.ReadFile(db.Path)
	if readErr != nil {
		return dbs, readErr
	}
	err := json.Unmarshal(fileContents, &dbs)
	if err != nil {
		fmt.Println(err)
		return dbs, err
	} 
	return dbs, nil
}
