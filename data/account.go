// Copyright 2014 The Cactus Authors. All rights reserved.

package data

import (
	"crypto/rand"
	"crypto/sha1"
	"database/sql"
	"strings"
	"time"

	"labix.org/v2/mgo/bson"
)

type Level int

const (
	_ Level = iota
	Participant
	Judge
	Administrator
)

type Account struct {
	Id int64 `json:"id"`

	Handle   string `json:"handle"`
	Password string `json:"-"`
	Salt     []byte `json:"-"`

	Level Level `json:"level"`

	Name string `json:"name"`

	Notified time.Time `json:"notified"`

	Created  time.Time `json:"created"`
	Modified time.Time `json:"modified"`
}

func ListAccounts() ([]*Account, error) {
	rows, err := db.Query("SELECT id, struct FROM accounts ORDER BY name_lower")
	if err != nil {
		return nil, err
	}

	accs := []*Account{}
	for rows.Next() {
		id := int64(0)
		b := []byte{}
		err = rows.Scan(&id, &b)
		if err != nil {
			return nil, err
		}
		a := &Account{}
		err = bson.Unmarshal(b, a)
		if err != nil {
			return nil, err
		}
		a.Id = id
		accs = append(accs, a)
	}
	return accs, nil
}

func GetAccount(id int64) (*Account, error) {
	b := []byte{}
	err := db.QueryRow("SELECT struct FROM accounts WHERE id=?", id).
		Scan(&b)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	a := &Account{}
	err = bson.Unmarshal(b, a)
	if err != nil {
		return nil, err
	}
	a.Id = id
	return a, nil
}

func GetAccountByHandle(handle string) (*Account, error) {
	id := int64(0)
	b := []byte{}
	err := db.QueryRow("SELECT id, struct FROM accounts WHERE handle_lower=?", handle).
		Scan(&id, &b)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	a := &Account{}
	err = bson.Unmarshal(b, a)
	if err != nil {
		return nil, err
	}
	a.Id = id
	return a, nil
}

func (a *Account) SetPassword(clear string) error {
	_, err := rand.Read(a.Salt)
	if err != nil {
		return err
	}
	sum := sha1.Sum([]byte(string(a.Salt) + clear))
	a.Password = string(sum[:])
	return nil
}

func (a *Account) CmpPassword(clear string) (bool, error) {
	sum := sha1.Sum([]byte(string(a.Salt) + clear))
	if a.Password != string(sum[:]) {
		return false, nil
	}
	return true, nil
}

func (a *Account) Put() error {
	a.Modified = time.Now()

	if a.Id == 0 {
		a.Created = a.Modified

		b, err := bson.Marshal(a)
		if err != nil {
			return err
		}

		res, err := db.Exec("INSERT INTO accounts(handle_lower, name_lower, level, struct, created, modified) VALUES(?, ?, ?, ?, ?, ?)", strings.ToLower(a.Handle), strings.ToLower(a.Name), a.Level, b, a.Created, a.Modified)
		if err != nil {
			return err
		}

		a.Id, err = res.LastInsertId()
		if err != nil {
			return err
		}
		chZerk <- a.Id
		return nil

	} else {
		b, err := bson.Marshal(a)
		if err != nil {
			return err
		}

		_, err = db.Exec("UPDATE accounts SET handle_lower=?, name_lower=?, level=?, struct=?, modified=? WHERE id=?", strings.ToLower(a.Handle), strings.ToLower(a.Name), a.Level, b, a.Modified, a.Id)
		chZerk <- a.Id
		return err
	}
}

func (a *Account) Del() error {
	_, err := db.Exec("DELETE FROM accounts WHERE id=?", a.Id)
	return err
}
