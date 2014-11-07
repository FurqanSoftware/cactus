// Copyright 2014 The Cactus Authors. All rights reserved.

package data

import (
	"database/sql"
	"time"

	"labix.org/v2/mgo/bson"
)

type Verdict int

const (
	_ Verdict = iota
	Accepted
	WrongAnswer
	CpuLimitExceeded
	MemoryLimitExceeded
	CompilationError
)

type Submission struct {
	Id int64 `json:"id"`

	AuthorId  int64  `json:"authorId"`
	ProblemId int64  `json:"problemId"`
	Language  string `json:"language"`
	SourceKey string `json:"sourceKey,omitempty"`

	Manual  bool    `json:"manual"`
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

	Created  time.Time `json:"created"`
	Modified time.Time `json:"modified"`
}

func ListSubmissions(cursor int64) ([]*Submission, error) {
	rows, err := db.Query("SELECT id, struct FROM submissions WHERE id<? ORDER BY created DESC LIMIT 128", cursor)
	if err != nil {
		return nil, err
	}

	subms := []*Submission{}
	for rows.Next() {
		id := int64(0)
		b := []byte{}
		err = rows.Scan(&id, &b)
		if err != nil {
			return nil, err
		}
		s := &Submission{}
		err = bson.Unmarshal(b, s)
		if err != nil {
			return nil, err
		}
		s.Id = id
		subms = append(subms, s)
	}
	return subms, nil
}

func ListSubmissionsByAuthor(acc *Account) ([]*Submission, error) {
	rows, err := db.Query("SELECT id, struct FROM submissions WHERE author_id = ? ORDER BY created DESC", acc.Id)
	if err != nil {
		return nil, err
	}

	subms := []*Submission{}
	for rows.Next() {
		id := int64(0)
		b := []byte{}
		err = rows.Scan(&id, &b)
		if err != nil {
			return nil, err
		}
		s := &Submission{}
		err = bson.Unmarshal(b, s)
		if err != nil {
			return nil, err
		}
		s.Id = id
		subms = append(subms, s)
	}
	return subms, nil
}

func GetSubmission(id int64) (*Submission, error) {
	b := []byte{}
	err := db.QueryRow("SELECT struct FROM submissions WHERE id=?", id).
		Scan(&b)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	s := &Submission{}
	err = bson.Unmarshal(b, s)
	if err != nil {
		return nil, err
	}
	s.Id = id
	return s, nil
}

func (s *Submission) Problem() (*Problem, error) {
	return GetProblem(s.ProblemId)
}

func (s *Submission) Apply(exec *Execution) {
	s.Manual = false
	s.Verdict = exec.Verdict
	s.Tests = exec.Tests
}

func (s *Submission) Tamper(verdict Verdict) {
	s.Manual = true
	s.Verdict = verdict
	s.Tests = nil
}

func (s *Submission) Reset() {
	s.Manual = false
	s.Verdict = 0
	s.Tests = nil
}

func (s *Submission) Put() error {
	s.Modified = time.Now()

	if s.Id == 0 {
		s.Created = s.Modified

		b, err := bson.Marshal(s)
		if err != nil {
			return err
		}

		res, err := db.Exec("INSERT INTO submissions(author_id, struct, created, modified) VALUES(?, ?, ?, ?)", s.AuthorId, b, s.Created, s.Modified)
		if err != nil {
			return err
		}
		chZerk <- s.AuthorId

		s.Id, err = res.LastInsertId()
		return err

	} else {
		b, err := bson.Marshal(s)
		if err != nil {
			return err
		}

		_, err = db.Exec("UPDATE submissions SET author_id=?, struct=?, modified=? WHERE id=?", s.AuthorId, b, s.Modified, s.Id)
		chZerk <- s.AuthorId
		return err
	}
}
