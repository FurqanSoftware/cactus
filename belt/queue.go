// Copyright 2014 The Cactus Authors. All rights reserved.

package belt

import (
	"github.com/hjr265/cactus/data"
)

var Queue interface {
	Push(*data.Execution) error
	Pop(bool) (*Execution, error)
}
