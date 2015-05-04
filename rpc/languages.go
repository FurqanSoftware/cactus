// Copyright 2014 The Cactus Authors. All rights reserved.

package rpc

import (
	"net/rpc"

	"github.com/hjr265/cactus/data"
)

type Languages struct{}

func (q *Languages) Get(id int64, reply *data.Language) error {
	lang, err := data.GetLanguage(id)
	if err != nil {
		return err
	}
	*reply = *lang
	return nil
}

func init() {
	rpc.Register(&Languages{})
}
