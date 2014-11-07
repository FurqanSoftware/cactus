// Copyright 2014 The Cactus Authors. All rights reserved.

package data

import (
	"database/sql"
	"time"

	"labix.org/v2/mgo/bson"
)

type Contest struct {
	Id int64 `json:"id"`

	Title  string `json:"title"`
	Footer string `json:"footer"`

	Starts time.Time `json:"starts"`
	Length int       `json:"length"`

	Stacks struct {
		C struct {
			Flags string `json:"flags"`
		} `json:"c"`
		Cpp struct {
			Flags string `json:"flags"`
		} `json:"cpp"`
		Java struct {
			Flags string `json:"flags"`
		} `json:"java"`
	} `json:"stacks"`

	Salt  []byte `json:"-"`
	Ready bool   `json:"ready"`

	Created  time.Time `json:"created"`
	Modified time.Time `json:"modified"`
}

func GetContest() (*Contest, error) {
	id := int64(0)
	b := []byte{}
	salt := []byte{}
	ready := false
	err := db.QueryRow("SELECT id, struct, salt, ready FROM contests LIMIT 1").
		Scan(&id, &b, &salt, &ready)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	c := &Contest{}
	if len(b) > 0 {
		err = bson.Unmarshal(b, c)
		if err != nil {
			return nil, err
		}
	}
	c.Id = id
	c.Salt = salt
	c.Ready = ready
	return c, nil
}

func (c *Contest) Put() error {
	c.Modified = time.Now()

	b, err := bson.Marshal(c)
	if err != nil {
		return err
	}

	_, err = db.Exec("UPDATE contests SET struct=?, ready=?, created=?, modified=? WHERE id=?", b, c.Ready, c.Created, c.Modified, c.Id)
	return err
}

func (c *Contest) Started() bool {
	return c.Starts.Before(time.Now())
}

func (c *Contest) Ends() time.Time {
	return c.Starts.Add(time.Duration(c.Length) * time.Minute)
}

func (c *Contest) Ended() bool {
	return c.Ends().Before(time.Now())
}

func (c *Contest) Running() bool {
	return c.Started() && !c.Ended()
}
