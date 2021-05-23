// Copyright 2014 The Cactus Authors. All rights reserved.

package data

import (
	"io/ioutil"
	"log"
	"path"
	"time"

	"database/sql"

	"github.com/hjr265/go-zrsc/zrsc"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func init() {
	db = func() *sql.DB {
		db, err := sql.Open("sqlite3", "cactus.db")
		catch(err)
		return db
	}()

	v := 0
	err := db.QueryRow("PRAGMA user_version").
		Scan(&v)
	if v > 0 && v < 1 {
		log.Fatal("incompatible database; exiting")
	}
	if v == 0 {
		_, err = db.Exec("PRAGMA user_version = 1")
		catch(err)
	}

	f, err := zrsc.Open(path.Join("data", "db-init.sql"))
	catch(err)
	b, err := ioutil.ReadAll(f)
	catch(err)
	_, err = db.Exec(string(b))
	catch(err)
	err = f.Close()
	catch(err)

	cnt, err := GetContest()
	catch(err)
	if !cnt.Ready {
		cnt.Title = "Untitled"
		cnt.Starts = time.Now().Add(1 * time.Hour)
		cnt.Length = 120
		cnt.Ready = true
		cnt.Created = time.Now()
		err = cnt.Put()
		catch(err)
	}

	acc, err := GetAccount(1)
	catch(err)
	if acc == nil {
		acc = &Account{
			Handle: "cactus",
			Level:  Administrator,
			Name:   "Cactus",
		}
		err = acc.SetPassword("cactus")
		catch(err)
		err = acc.Put()
		catch(err)
	}
}
