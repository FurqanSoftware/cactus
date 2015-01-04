// Copyright 2014 The Cactus Authors. All rights reserved.

package belt

import (
	"bytes"
	"io"
	"io/ioutil"

	"github.com/hjr265/cactus/data"
)

func GetBlob(key string) (io.ReadCloser, error) {
	switch Queue := Queue.(type) {
	case *LocalQueue:
		return data.Blobs.Get(key)

	case *RemoteQueue:
		b := []byte{}
		err := Queue.c.Call("Blobs.Get", key, &b)
		if err != nil {
			return nil, err
		}
		return ioutil.NopCloser(bytes.NewReader(b)), nil

	}

	panic("unreachable")
}

func PutBlob(key string, blob io.Reader) (string, error) {
	switch Queue := Queue.(type) {
	case *LocalQueue:
		return data.Blobs.Put(key, blob)

	case *RemoteQueue:
		args := struct {
			Key  string
			Blob []byte
		}{}
		args.Key = key
		b, err := ioutil.ReadAll(blob)
		if err != nil {
			return "", err
		}
		args.Blob = b
		err = Queue.c.Call("Blobs.Put", args, &key)
		if err != nil {
			return "", err
		}
		return key, nil

	}

	panic("unreachable")
}
