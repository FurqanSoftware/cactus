// Copyright 2014 The Cactus Authors. All rights reserved.

package data

import (
	"database/sql"
	"strconv"
	"time"

	"labix.org/v2/mgo/bson"

	"github.com/hjr265/cactus/hub"
)

type Standing struct {
	Id int64 `json:"id"`

	AuthorId int64 `json:"authorId"`
	Author   struct {
		Handle string `json:"handle"`
		Name   string `json:"name"`
	} `json:"author"`

	Score   int `json:"score"`
	Penalty int `json:"penalty"`

	Attempts map[string]*Attempt `json:"attempts,omitempty"`

	Created  time.Time `json:"created"`
	Modified time.Time `json:"modified"`
}

type Attempt struct {
	ProblemId int64 `json:"problemId"`

	Tries   int  `json:"tries"`
	Extras  int  `json:"extras"`
	Score   int  `json:"score"`
	Penalty int  `json:"penalty"`
	Perfect bool `json:"perfect"`
}

func ListStandings() ([]*Standing, error) {
	rows, err := db.Query("SELECT id, struct FROM standings ORDER BY score DESC, penalty")
	if err != nil {
		return nil, err
	}

	stnds := []*Standing{}
	for rows.Next() {
		id := int64(0)
		b := []byte{}
		err = rows.Scan(&id, &b)
		if err != nil {
			return nil, err
		}
		s := &Standing{}
		err = bson.Unmarshal(b, s)
		if err != nil {
			return nil, err
		}
		s.Id = id
		stnds = append(stnds, s)
	}
	return stnds, nil
}

func GetStandingByAuthor(acc *Account) (*Standing, error) {
	id := int64(0)
	b := []byte{}
	err := db.QueryRow("SELECT id, struct FROM standings WHERE author_id=?", acc.Id).
		Scan(&id, &b)
	if err == sql.ErrNoRows {
		return &Standing{
			AuthorId: acc.Id,
		}, nil
	}
	if err != nil {
		return nil, err
	}
	s := &Standing{}
	err = bson.Unmarshal(b, s)
	if err != nil {
		return nil, err
	}
	s.Id = id
	return s, nil
}

func (s *Standing) Put() error {
	s.Modified = time.Now()

	if s.Id == 0 {
		s.Created = s.Modified

		b, err := bson.Marshal(s)
		if err != nil {
			return err
		}

		res, err := db.Exec("INSERT INTO standings(author_id, score, penalty, struct, created, modified) VALUES(?, ?, ?, ?, ?, ?)", s.AuthorId, s.Score, s.Penalty, b, s.Created, s.Modified)
		if err != nil {
			return err
		}

		s.Id, err = res.LastInsertId()
		return err

	} else {
		b, err := bson.Marshal(s)
		if err != nil {
			return err
		}

		_, err = db.Exec("UPDATE standings SET author_id=?, score=?, penalty=?, struct=?, modified=? WHERE id=?", s.AuthorId, s.Score, s.Penalty, b, s.Modified, s.Id)
		return err
	}
}

func (s *Standing) Del() error {
	if s.Id == 0 {
		return nil
	}

	_, err := db.Exec("DELETE FROM standings WHERE id=?", s.Id)
	return err
}

var chZerk = make(chan int64, 4096)

func init() {
	go func() {
		for id := range chZerk {
			cnt, err := GetContest()
			catch(err)

			acc, err := GetAccount(id)
			catch(err)

			s, err := GetStandingByAuthor(acc)
			catch(err)
			if s == nil {
				continue
			}

			if acc.Level != Participant {
				err = s.Del()
				catch(err)
				continue
			}

			s.Author.Handle = acc.Handle
			s.Author.Name = acc.Name

			s.Attempts = map[string]*Attempt{}
			probs := map[int64]*Problem{}
			subms, err := ListSubmissionsByAuthor(acc)
			catch(err)
			for i := len(subms) - 1; i >= 0; i-- {
				subm := subms[i]
				prob, _ := probs[subm.ProblemId]
				if prob == nil {
					prob, err = GetProblem(subm.ProblemId)
					catch(err)
					probs[subm.ProblemId] = prob

					delete(s.Attempts, strconv.FormatInt(prob.Id, 10))
				}

				att := s.Attempts[strconv.FormatInt(prob.Id, 10)]
				if att == nil {
					att = &Attempt{
						ProblemId: prob.Id,
					}
					s.Attempts[strconv.FormatInt(prob.Id, 10)] = att
				}

				if subm.Verdict == Accepted {
					score := 0
					perfect := true
					if subm.Manual {
						for _, test := range prob.Tests {
							score += test.Points
						}
					} else {
						for _, test := range subm.Tests {
							if test.Verdict == Accepted {
								score += test.Points
							} else {
								perfect = false
							}
						}
					}

					if score > att.Score {
						att.Score = score
						att.Tries += att.Extras + 1
						att.Extras = 0
						att.Penalty = (att.Tries-1)*20 + int(subm.Created.Sub(cnt.Starts)/time.Minute)
						att.Perfect = perfect
					}
				} else if subm.Verdict != CompilationError {
					att.Extras += 1
				}
			}
			s.Score = 0
			s.Penalty = 0
			for _, att := range s.Attempts {
				s.Score += att.Score
				s.Penalty += att.Penalty
			}

			err = s.Put()
			catch(err)

			hub.Send([]interface{}{"SYNC", "standings"})
		}
	}()
}
