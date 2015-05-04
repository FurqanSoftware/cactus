// Copyright 2014 The Cactus Authors. All rights reserved.

package data

import (
	"database/sql"
	"time"

	"labix.org/v2/mgo/bson"
)

type Language struct {
	Id int64 `json:"id"`

	Label string `json:"label"`

	Steps struct {
		Build string `json:"build"`
		Run   string `json:"run"`
	} `json:"steps"`

	Created  time.Time `json:"created"`
	Modified time.Time `json:"modified"`
}

func ListLanguages() ([]*Language, error) {
	rows, err := db.Query("SELECT id, struct FROM languages")
	if err != nil {
		return nil, err
	}

	langs := []*Language{}
	for rows.Next() {
		id := int64(0)
		b := []byte{}
		err = rows.Scan(&id, &b)
		if err != nil {
			return nil, err
		}
		l := &Language{}
		err = bson.Unmarshal(b, l)
		if err != nil {
			return nil, err
		}
		l.Id = id
		langs = append(langs, l)
	}
	return langs, nil
}

func GetLanguage(id int64) (*Language, error) {
	b := []byte{}
	err := db.QueryRow("SELECT struct FROM languages WHERE id=?", id).
		Scan(&b)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	l := &Language{}
	err = bson.Unmarshal(b, l)
	if err != nil {
		return nil, err
	}
	l.Id = id
	return l, nil
}

func (l *Language) Put() error {
	l.Modified = time.Now()

	if l.Id == 0 {
		l.Created = l.Modified

		b, err := bson.Marshal(l)
		if err != nil {
			return err
		}

		res, err := db.Exec("INSERT INTO languages(struct, created, modified) VALUES(?, ?, ?)", b, l.Created, l.Modified)
		if err != nil {
			return err
		}

		l.Id, err = res.LastInsertId()
		return err

	} else {
		b, err := bson.Marshal(l)
		if err != nil {
			return err
		}

		_, err = db.Exec("UPDATE languages SET struct=?, modified=? WHERE id=?", b, l.Modified, l.Id)
		return err
	}
}

func (l *Language) Del() error {
	_, err := db.Exec("DELETE FROM languages WHERE id=?", l.Id)
	return err
}
