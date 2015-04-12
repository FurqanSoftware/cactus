// Copyright 2014 The Cactus Authors. All rights reserved.

package cube

import (
	"io"
)

type Cube interface {
	Create(string, io.Reader) error
	Execute(*Process) error

	Dispose() error
}

func New() (Cube, error) {
	if DockerExe != "" {
		c := &Docker{}
		err := c.Init()
		if err != nil {
			return nil, err
		}
		return c, nil

	} else {
		c := &Plain{}
		err := c.Init()
		if err != nil {
			return nil, err
		}
		return c, nil
	}
}
