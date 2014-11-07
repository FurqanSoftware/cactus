// Copyright 2014 The Cactus Authors. All rights reserved.

package data

import (
	"database/sql"
	"time"

	"labix.org/v2/mgo/bson"
)

type Problem struct {
	Id int64 `bson:"-" json:"id"`

	Slug string `json:"slug"`

	Char  string `json:"char"`
	Title string `json:"title"`

	Statement struct {
		Body   string `json:"body"`
		Input  string `json:"input"`
		Output string `json:"output"`
	} `json:"statement"`
	Samples []struct {
		Input  string `json:"input"`
		Answer string `json:"answer"`
	} `json:"samples,omitempty"`
	Notes string `json:"notes"`

	Judge string `json:"judge"`

	Checker struct {
		Language  string `json:"language"`
		SourceKey string `json:"sourceKey"`
	} `json:"checker"`

	Limits struct {
		Cpu    float64 `json:"cpu"`
		Memory int     `json:"memory"`
	} `json:"limits"`

	Languages []string `json:"languages"`

	Tests []struct {
		InputKey  string `json:"inputKey"`
		AnswerKey string `json:"answerKey"`
		Points    int    `json:"points"`
	} `json:"tests,omitempty"`

	Scoring string `json:"scoring"`

	Created  time.Time `json:"created"`
	Modified time.Time `json:"modified"`
}

func ListProblems() ([]*Problem, error) {
	rows, err := db.Query("SELECT id, struct FROM problems ORDER BY char")
	if err != nil {
		return nil, err
	}

	probs := []*Problem{}
	for rows.Next() {
		id := int64(0)
		b := []byte{}
		err = rows.Scan(&id, &b)
		if err != nil {
			return nil, err
		}
		p := &Problem{}
		err = bson.Unmarshal(b, p)
		if err != nil {
			return nil, err
		}
		p.Id = id
		probs = append(probs, p)
	}
	return probs, nil
}

func GetProblem(id int64) (*Problem, error) {
	b := []byte{}
	err := db.QueryRow("SELECT struct FROM problems WHERE id=?", id).
		Scan(&b)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	p := &Problem{}
	err = bson.Unmarshal(b, p)
	if err != nil {
		return nil, err
	}
	p.Id = id
	return p, nil
}

func GetProblemBySlug(slug string) (*Problem, error) {
	id := int64(0)
	b := []byte{}
	err := db.QueryRow("SELECT id, struct FROM problems WHERE slug=?", slug).
		Scan(&id, &b)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	p := &Problem{}
	err = bson.Unmarshal(b, p)
	if err != nil {
		return nil, err
	}
	p.Id = id
	return p, nil
}

func (p *Problem) Put() error {
	p.Modified = time.Now()

	if p.Id == 0 {
		p.Created = p.Modified

		b, err := bson.Marshal(p)
		if err != nil {
			return err
		}

		res, err := db.Exec("INSERT INTO problems(slug, char, struct, created, modified) VALUES(?, ?, ?, ?, ?)", p.Slug, p.Char, b, p.Created, p.Modified)
		if err != nil {
			return err
		}

		p.Id, err = res.LastInsertId()
		return err

	} else {
		b, err := bson.Marshal(p)
		if err != nil {
			return err
		}

		_, err = db.Exec("UPDATE problems SET slug=?, char=?, struct=?, modified=? WHERE id=?", p.Slug, p.Char, b, p.Modified, p.Id)
		return err
	}
}

func (p *Problem) Del() error {
	_, err := db.Exec("DELETE FROM problems WHERE id=?", p.Id)
	return err
}
