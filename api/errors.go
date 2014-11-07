// Copyright 2014 The Cactus Authors. All rights reserved.

package api

func catch(err error) {
	if err != nil {
		panic(err)
	}
}
