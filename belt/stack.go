// Copyright 2014 The Cactus Authors. All rights reserved.

package belt

import (
	"io"

	"github.com/hjr265/cactus/cube"
)

type Stack interface {
	Build(cube.Cube, io.Reader) (*cube.Process, error)
	Run(cube.Cube) *cube.Process
}

var Stacks = map[string]Stack{
	"c":   &StackC{},
	"cpp": &StackCpp{},
}
