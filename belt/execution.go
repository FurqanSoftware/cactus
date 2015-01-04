// Copyright 2014 The Cactus Authors. All rights reserved.

package belt

import (
	"github.com/hjr265/cactus/data"
	"github.com/hjr265/cactus/hub"
)

type Execution struct {
	*data.Execution
}

func (e *Execution) Submission() (*Submission, error) {
	switch Queue := Queue.(type) {
	case *LocalQueue:
		subm, err := e.Execution.Submission()
		if err != nil {
			return nil, err
		}
		return &Submission{subm}, nil

	case *RemoteQueue:
		subm := data.Submission{}
		err := Queue.c.Call("Submissions.Get", e.Execution.SubmissionId, &subm)
		if err != nil {
			return nil, err
		}
		return &Submission{&subm}, nil
	}

	panic("unreachable")
}

func (e *Execution) Put() error {
	switch Queue := Queue.(type) {
	case *LocalQueue:
		err := e.Execution.Put()
		if err != nil {
			return err
		}
		hub.Send([]interface{}{"SYNC", "executions", e.Id})
		return nil

	case *RemoteQueue:
		return Queue.c.Call("Executions.Put", e.Execution, nil)
	}

	panic("unreachable")
}
