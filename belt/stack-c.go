// Copyright 2014 The Cactus Authors. All rights reserved.

package belt

import (
	"io"

	"github.com/hjr265/cactus/cube"
)

type StackC struct{}

func (s *StackC) Build(runCube cube.Cube, source io.Reader) (*cube.Process, error) {
	err := runCube.Create("source.c", source)
	if err != nil {
		return nil, err
	}

	return &cube.Process{
		Name: "gcc",
		Args: []string{"source.c"},
		Limits: cube.Limits{
			Cpu:    16,
			Memory: 1024,
		},
	}, nil
}

func (s *StackC) Run(runCube cube.Cube) *cube.Process {
	return &cube.Process{
		Name: "./a.out",
		Args: []string{},
	}
}
