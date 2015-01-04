// Copyright 2014 The Cactus Authors. All rights reserved.

package rpc

import (
	"net/rpc"

	"github.com/hjr265/cactus/data"
	"github.com/hjr265/cactus/hub"
)

type Submissions struct{}

func (q *Submissions) Get(id int64, reply *data.Submission) error {
	subm, err := data.GetSubmission(id)
	if err != nil {
		return err
	}
	*reply = *subm
	return nil
}

func (q *Submissions) Put(subm *data.Submission, reply *int) error {
	err := subm.Put()
	if err != nil {
		return err
	}
	hub.Send([]interface{}{"SYNC", "submissions", subm.Id})
	return nil
}

func init() {
	rpc.Register(&Submissions{})
}
