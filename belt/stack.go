// Copyright 2014 The Cactus Authors. All rights reserved.

package belt

import (
	"io"

	"github.com/hjr265/jail.go/jail"
)

type Stack interface {
	Build(cell *jail.Cell, source io.Reader) (*jail.Cmd, error)
	Run(cell *jail.Cell) *jail.Cmd
}

var Stacks = map[string]Stack{
	"c":   &StackC{},
	"cpp": &StackCpp{},
}
