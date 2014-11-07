// Copyright 2014 The Cactus Authors. All rights reserved.

package data

import (
	"time"

	"labix.org/v2/mgo/bson"
)

type Activity struct {
	Id int64 `json:"id"`

	ActorId int64 `json:"actorId"`
	Actor   struct {
		Handle string `json:"handle"`
		Name   string `json:"name"`
	} `json:"actor"`
	Record string `json:"record"`

	Created  time.Time `json:"created"`
	Modified time.Time `json:"modified"`
}

func NewActivity(acc *Account, rec string) *Activity {
	a := &Activity{}
	a.ActorId = acc.Id
	a.Actor.Handle = acc.Handle
	a.Actor.Name = acc.Name
	a.Record = rec
	return a
}

func ListActivities(cursor int64) ([]*Activity, error) {
	rows, err := db.Query("SELECT id, struct FROM activities WHERE id<? ORDER BY created DESC LIMIT 256", cursor)
	if err != nil {
		return nil, err
	}

	acts := []*Activity{}
	for rows.Next() {
		id := int64(0)
		b := []byte{}
		err = rows.Scan(&id, &b)
		if err != nil {
			return nil, err
		}
		a := &Activity{}
		err = bson.Unmarshal(b, a)
		if err != nil {
			return nil, err
		}
		a.Id = id
		acts = append(acts, a)
	}
	return acts, nil
}

func (e *Activity) Put() error {
	e.Modified = time.Now()
	e.Created = e.Modified

	b, err := bson.Marshal(e)
	if err != nil {
		return err
	}

	res, err := db.Exec("INSERT INTO activities(actor_id, struct, created, modified) VALUES(?, ?, ?, ?)", e.ActorId, b, e.Created, e.Modified)
	if err != nil {
		return err
	}

	e.Id, err = res.LastInsertId()
	return err
}
