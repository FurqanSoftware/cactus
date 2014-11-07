// Copyright 2014 The Cactus Authors. All rights reserved.

package data

import (
	"database/sql"
	"time"

	"labix.org/v2/mgo/bson"
)

type Execution struct {
	Id int64 `json:"id"`

	SubmissionId int64 `json:"submissionId"`

	Status int `json:"status"`

	Build struct {
		Error string `json:"error"`

		Usages struct {
			Cpu    float64 `json:"cpu"`
			Memory int     `json:"memory"`
		} `json:"usages"`
	} `json:"build"`

	Verdict Verdict `json:"verdict"`
	Tests   []struct {
		OutputKey string `json:"outputKey"`

		Difference int `json:"difference"`

		Verdict Verdict `json:"verdict"`
		Usages  struct {
			Cpu    float64 `json:"cpu"`
			Memory int     `json:"memory"`
		} `json:"usages"`

		Points int `json:"points"`
	} `json:"tests"`

	Apply bool `json:"apply"`

	Created  time.Time `json:"created"`
	Modified time.Time `json:"modified"`
}

func GetExecution(id int64) (*Execution, error) {
	b := []byte{}
	err := db.QueryRow("SELECT struct FROM executions WHERE id=?", id).
		Scan(&b)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	e := &Execution{}
	err = bson.Unmarshal(b, e)
	if err != nil {
		return nil, err
	}
	e.Id = id
	return e, nil
}

func (e *Execution) Submission() (*Submission, error) {
	return GetSubmission(e.SubmissionId)
}

func (e *Execution) Put() error {
	e.Modified = time.Now()

	if e.Id == 0 {
		e.Created = e.Modified

		b, err := bson.Marshal(e)
		if err != nil {
			return err
		}

		res, err := db.Exec("INSERT INTO executions(struct, created, modified) VALUES(?, ?, ?)", b, e.Created, e.Modified)
		if err != nil {
			return err
		}

		e.Id, err = res.LastInsertId()
		return err

	} else {
		b, err := bson.Marshal(e)
		if err != nil {
			return err
		}

		_, err = db.Exec("UPDATE executions SET struct=?, modified=? WHERE id=?", b, e.Modified, e.Id)
		return err
	}
}

func (e *Execution) Del() error {
	_, err := db.Exec("DELETE FROM executions WHERE id=?", e.Id)
	return err
}
