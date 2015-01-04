// Copyright 2014 The Cactus Authors. All rights reserved.

package rpc

import (
	"net/rpc"

	"github.com/hjr265/cactus/data"
	"github.com/hjr265/cactus/hub"
)

type Executions struct{}

func (q *Executions) Put(exec *data.Execution, reply *int) error {
	err := exec.Put()
	if err != nil {
		return err
	}
	hub.Send([]interface{}{"SYNC", "executions", exec.Id})
	return nil
}

func init() {
	rpc.Register(&Executions{})
}
