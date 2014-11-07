// Copyright 2014 The Cactus Authors. All rights reserved.

package data

import (
	"time"

	"labix.org/v2/mgo/bson"
)

type Notification struct {
	Id int64 `json:"id"`

	AccountId    int64 `json:"accountId"`
	AccountLevel Level `json:"accountLevel"`

	Message string `json:"message"`

	Created  time.Time `json:"created"`
	Modified time.Time `json:"modified"`
}

func NewNotification(accId int64, accLevel Level, msg string) *Notification {
	return &Notification{
		AccountId:    accId,
		AccountLevel: accLevel,
		Message:      msg,
	}
}

func ListNotificationsForAccount(acc *Account, cursor int64) ([]*Notification, error) {
	rows, err := db.Query("SELECT id, struct FROM notifications WHERE (account_id=? OR account_level=? OR account_id = 0 AND account_level = 0) AND id<? ORDER BY created DESC LIMIT 32", acc.Id, acc.Level, cursor)
	if err != nil {
		return nil, err
	}

	notifs := []*Notification{}
	for rows.Next() {
		id := int64(0)
		b := []byte{}
		err = rows.Scan(&id, &b)
		if err != nil {
			return nil, err
		}
		n := &Notification{}
		err = bson.Unmarshal(b, n)
		if err != nil {
			return nil, err
		}
		n.Id = id
		notifs = append(notifs, n)
	}
	return notifs, nil
}

func (n *Notification) Put() error {
	n.Modified = time.Now()
	n.Created = n.Modified

	b, err := bson.Marshal(n)
	if err != nil {
		return err
	}

	res, err := db.Exec("INSERT INTO notifications(account_id, account_level, struct, created, modified) VALUES(?, ?, ?, ?, ?)", n.AccountId, n.AccountLevel, b, n.Created, n.Modified)
	if err != nil {
		return err
	}

	n.Id, err = res.LastInsertId()
	return err
}
