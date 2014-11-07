// Copyright 2014 The Cactus Authors. All rights reserved.

package data

import (
	"database/sql"
	"time"

	"labix.org/v2/mgo/bson"
)

type Response int

const (
	Unresponded Response = iota
	Ignored
	Answered
	Broadcasted
)

type Clarification struct {
	Id int64 `bson:"-" json:"id"`

	AskerId   int64    `json:"askerId"`
	ProblemId int64    `json:"problemId"`
	Question  string   `json:"question"`
	Response  Response `json:"response"`
	Message   string   `json:"message"`

	Created  time.Time `json:"created"`
	Modified time.Time `json:"modified"`
}

func ListClarifications() ([]*Clarification, error) {
	rows, err := db.Query("SELECT id, struct FROM clarifications")
	if err != nil {
		return nil, err
	}

	clars := []*Clarification{}
	for rows.Next() {
		id := int64(0)
		b := []byte{}
		err = rows.Scan(&id, &b)
		if err != nil {
			return nil, err
		}
		c := &Clarification{}
		err = bson.Unmarshal(b, c)
		if err != nil {
			return nil, err
		}
		c.Id = id
		clars = append(clars, c)
	}
	return clars, nil
}

func ListClarificationsForAccount(acc *Account) ([]*Clarification, error) {
	rows, err := db.Query("SELECT id, struct FROM clarifications WHERE asker_id=? OR response=?", acc.Id, Broadcasted)
	if err != nil {
		return nil, err
	}

	clars := []*Clarification{}
	for rows.Next() {
		id := int64(0)
		b := []byte{}
		err = rows.Scan(&id, &b)
		if err != nil {
			return nil, err
		}
		c := &Clarification{}
		err = bson.Unmarshal(b, c)
		if err != nil {
			return nil, err
		}
		c.Id = id
		clars = append(clars, c)
	}
	return clars, nil
}

func GetClarification(id int64) (*Clarification, error) {
	b := []byte{}
	err := db.QueryRow("SELECT struct FROM clarifications WHERE id=?", id).
		Scan(&b)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	c := &Clarification{}
	err = bson.Unmarshal(b, c)
	if err != nil {
		return nil, err
	}
	c.Id = id
	return c, nil
}

func (c *Clarification) Put() error {
	c.Modified = time.Now()

	if c.Id == 0 {
		c.Created = c.Modified

		b, err := bson.Marshal(c)
		if err != nil {
			return err
		}

		res, err := db.Exec("INSERT INTO clarifications(asker_id, response, struct, created, modified) VALUES(?, ?, ?, ?, ?)", c.AskerId, c.Response, b, c.Created, c.Modified)
		if err != nil {
			return err
		}

		c.Id, err = res.LastInsertId()
		return err

	} else {
		b, err := bson.Marshal(c)
		if err != nil {
			return err
		}

		_, err = db.Exec("UPDATE clarifications SET asker_id=?, response=?, struct=?, modified=? WHERE id=?", c.AskerId, c.Response, b, c.Modified, c.Id)
		return err
	}
}

func (c *Clarification) Del() error {
	_, err := db.Exec("DELETE FROM clarifications WHERE id=?", c.Id)
	return err
}
