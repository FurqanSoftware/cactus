// Copyright 2014 The Cactus Authors. All rights reserved.

package rpc

import (
	"net/rpc"

	"github.com/hjr265/cactus/data"
)

type Problems struct{}

func (q *Problems) Get(id int64, reply *data.Problem) error {
	exec, err := data.GetProblem(id)
	if err != nil {
		return err
	}
	*reply = *exec
	return nil
}

func init() {
	rpc.Register(&Problems{})
}
