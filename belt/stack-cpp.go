// Copyright 2014 The Cactus Authors. All rights reserved.

package belt

import (
	"io"

	"github.com/hjr265/cactus/cube"
)

type StackCpp struct{}

func (s *StackCpp) Build(runCube cube.Cube, source io.Reader) (*cube.Process, error) {
	err := runCube.Create("source.cpp", source)
	if err != nil {
		return nil, err
	}

	return &cube.Process{
		Name: "g++",
		Args: []string{"source.cpp"},
		Limits: cube.Limits{
			Cpu:    16,
			Memory: 1024,
		},
	}, nil
}

func (s *StackCpp) Run(runCube cube.Cube) *cube.Process {
	return &cube.Process{
		Name: "./a.out",
		Args: []string{},
	}
}
