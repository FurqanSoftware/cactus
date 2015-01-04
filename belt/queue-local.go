// Copyright 2014 The Cactus Authors. All rights reserved.

package belt

import (
	"github.com/hjr265/cactus/data"
)

type LocalQueue struct {
	ch chan *data.Execution
}

func NewLocalQueue() *LocalQueue {
	return &LocalQueue{
		ch: make(chan *data.Execution, 4096),
	}
}

func (q *LocalQueue) Push(exec *data.Execution) error {
	q.ch <- exec
	return nil
}

func (q *LocalQueue) Pop(wait bool) (*Execution, error) {
	if wait {
		return &Execution{<-q.ch}, nil

	} else {
		select {
		case exec := <-q.ch:
			return &Execution{exec}, nil

		default:
			return nil, nil
		}
	}
}
