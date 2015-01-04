// Copyright 2014 The Cactus Authors. All rights reserved.

package belt

import (
	"net/rpc"
	"sync"
	"time"

	"github.com/hjr265/cactus/data"
)

type RemoteQueue struct {
	c *rpc.Client

	t *time.Ticker
	sync.Mutex
}

func NewRemoteQueue(addr string) (*RemoteQueue, error) {
	c, err := rpc.DialHTTP("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &RemoteQueue{
		c: c,
		t: time.NewTicker(250 * time.Millisecond),
	}, nil
}

func (q *RemoteQueue) Push(*data.Execution) error {
	panic("unimplemented")
}

func (q *RemoteQueue) Pop(wait bool) (*Execution, error) {
	q.Lock()
	defer q.Unlock()

	for ; ; <-q.t.C {
		exec := data.Execution{}
		err := q.c.Call("Queue.Pop", 0, &exec)
		if err != nil {
			return nil, err
		}

		if exec.Id == 0 {
			if !wait {
				return nil, nil
			} else {
				continue
			}
		}
		return &Execution{&exec}, nil
	}
}
