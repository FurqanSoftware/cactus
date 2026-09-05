// Copyright 2014 The Cactus Authors. All rights reserved.

package data

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

// Open opens the database at dbPath and the blob store at blobsPath, applies the
// schema, and seeds the default contest and admin account. It must be called
// before anything else in this package is used.
func Open(dbPath, blobsPath string) error {
	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}

	v := 0
	err = db.QueryRow("PRAGMA user_version").
		Scan(&v)
	if err != nil {
		return err
	}
	if v > 0 && v < 1 {
		return fmt.Errorf("incompatible database %s", dbPath)
	}
	if v == 0 {
		_, err = db.Exec("PRAGMA user_version = 1")
		if err != nil {
			return err
		}
	}

	b, err := dbInitSQL.ReadFile("db-init.sql")
	if err != nil {
		return err
	}
	_, err = db.Exec(string(b))
	if err != nil {
		return err
	}

	err = openBlobs(blobsPath)
	if err != nil {
		return err
	}

	cnt, err := GetContest()
	if err != nil {
		return err
	}
	if !cnt.Ready {
		cnt.Title = "Untitled"
		cnt.Starts = time.Now().Add(1 * time.Hour)
		cnt.Length = 120
		cnt.Ready = true
		cnt.Created = time.Now()
		err = cnt.Put()
		if err != nil {
			return err
		}
	}

	acc, err := GetAccount(1)
	if err != nil {
		return err
	}
	if acc == nil {
		acc = &Account{
			Handle: "cactus",
			Level:  Administrator,
			Name:   "Cactus",
		}
		err = acc.SetPassword("cactus")
		if err != nil {
			return err
		}
		err = acc.Put()
		if err != nil {
			return err
		}
		log.Print("Created default admin account (handle: cactus, password: cactus)")
	}

	return nil
}
