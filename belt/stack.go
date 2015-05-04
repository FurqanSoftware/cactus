// Copyright 2014 The Cactus Authors. All rights reserved.

package belt

import (
	"io"
	"strings"

	"github.com/hjr265/cactus/cube"
	"github.com/hjr265/cactus/data"
)

type Stack struct {
	*data.Language

	Source struct {
		Name string
	}
}

func (s *Stack) Build(runCube cube.Cube, source io.Reader) (*cube.Process, error) {
	err := runCube.Create(s.Source.Name, source)
	if err != nil {
		return nil, err
	}

	args := strings.Split(s.Language.Steps.Build, " ")
	for i, arg := range args {
		if arg == "${source.name}" {
			args[i] = s.Source.Name
		}
	}

	return &cube.Process{
		Name: args[0],
		Args: args[1:],
		Limits: cube.Limits{
			Cpu:    16,
			Memory: 1024,
		},
	}, nil
}

func (s *Stack) Run(c cube.Cube) *cube.Process {
	args := strings.Split(s.Language.Steps.Run, " ")
	for i, arg := range args {
		if arg == "${source.name}" {
			args[i] = s.Source.Name
		}
	}

	return &cube.Process{
		Name: args[0],
		Args: args[1:],
	}
}
