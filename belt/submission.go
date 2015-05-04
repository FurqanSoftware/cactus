// Copyright 2014 The Cactus Authors. All rights reserved.

package belt

import (
	"github.com/hjr265/cactus/data"
	"github.com/hjr265/cactus/hub"
)

type Submission struct {
	*data.Submission
}

func (s *Submission) Problem() (*Problem, error) {
	switch Queue := Queue.(type) {
	case *LocalQueue:
		prob, err := s.Submission.Problem()
		if err != nil {
			return nil, err
		}
		return &Problem{prob}, nil

	case *RemoteQueue:
		prob := data.Problem{}
		err := Queue.c.Call("Problems.Get", s.Submission.ProblemId, &prob)
		if err != nil {
			return nil, err
		}
		return &Problem{&prob}, nil
	}

	panic("unreachable")
}

func (s *Submission) Language() (*data.Language, error) {
	switch Queue := Queue.(type) {
	case *LocalQueue:
		return s.Submission.Language()

	case *RemoteQueue:
		lang := data.Language{}
		err := Queue.c.Call("Languages.Get", s.Submission.LanguageId, &lang)
		if err != nil {
			return nil, err
		}
		return &lang, nil
	}

	panic("unreachable")
}

func (s *Submission) Apply(exec *Execution) {
	s.Submission.Apply(exec.Execution)
}

func (s *Submission) Put() error {
	switch Queue := Queue.(type) {
	case *LocalQueue:
		err := s.Submission.Put()
		if err != nil {
			return err
		}
		hub.Send([]interface{}{"SYNC", "submissions", s.Id})
		return nil

	case *RemoteQueue:
		return Queue.c.Call("Submissions.Put", s.Submission, nil)
	}

	panic("unreachable")
}
