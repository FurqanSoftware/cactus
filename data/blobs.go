// Copyright 2014 The Cactus Authors. All rights reserved.

package data

import (
	"github.com/hjr265/bloo"
)

var Blobs *bloo.BS

func init() {
	Blobs = func() *bloo.BS {
		bs, err := bloo.Open("cactus.bs", 0766)
		catch(err)
		return bs
	}()
}
