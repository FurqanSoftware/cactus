// Copyright 2014 The Cactus Authors. All rights reserved.

package data

import (
	"github.com/hjr265/bloo"
)

var Blobs *bloo.BS

func openBlobs(path string) error {
	bs, err := bloo.Open(path, 0766)
	if err != nil {
		return err
	}
	Blobs = bs
	return nil
}
