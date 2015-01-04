// Copyright 2014 The Cactus Authors. All rights reserved.

package rpc

import (
	"bytes"
	"io/ioutil"
	"net/rpc"

	"github.com/hjr265/cactus/data"
)

type Blobs struct{}

func (o *Blobs) Get(key string, reply *[]byte) error {
	r, err := data.Blobs.Get(key)
	if err != nil {
		return err
	}
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	*reply = b
	return nil
}

type BlobPutArgs struct {
	Key  string
	Blob []byte
}

func (o *Blobs) Put(args *BlobPutArgs, reply *string) error {
	key, err := data.Blobs.Put(args.Key, ioutil.NopCloser(bytes.NewReader(args.Blob)))
	if err != nil {
		return err
	}
	*reply = key
	return nil
}

func init() {
	rpc.Register(&Blobs{})
}
