// Copyright 2014 The Cactus Authors. All rights reserved.

package rpc

import (
	"net/rpc"

	"github.com/hjr265/cactus/belt"
	"github.com/hjr265/cactus/data"
)

type Queue struct{}

func (q *Queue) Push(exec *data.Execution, reply *interface{}) error {
	err := belt.Queue.Push(exec)
	return err
}

func (q *Queue) Pop(args int, reply *data.Execution) error {
	exec, err := belt.Queue.Pop(false)
	if err != nil {
		return err
	}
	if exec != nil {
		*reply = *exec.Execution
	}
	return nil
}

func init() {
	rpc.Register(&Queue{})
}
