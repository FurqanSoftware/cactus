// Copyright 2014 The Cactus Authors. All rights reserved.

package belt

import (
	"log"
)

func catch(err error) {
	if err != nil {
		panic(err)
	}
}

func trace(err error) {
	if err != nil {
		log.Print(err)
	}
}
