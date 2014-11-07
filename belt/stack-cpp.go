// Copyright 2014 The Cactus Authors. All rights reserved.

package belt

import (
	"io"
	"time"

	"github.com/hjr265/jail.go/jail"
)

type StackCpp struct{}

func (s *StackCpp) Build(cell *jail.Cell, source io.Reader) (*jail.Cmd, error) {
	f, err := cell.Create("source.cpp")
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(f, source)
	if err != nil {
		return nil, err
	}
	err = f.Close()
	if err != nil {
		return nil, err
	}

	cmd := cell.Command("g++", "source.cpp")
	cmd.Limits.Cpu = 16 * time.Second
	cmd.Limits.Memory = 1 << 30

	return cmd, nil
}

func (s *StackCpp) Run(cell *jail.Cell) *jail.Cmd {
	return cell.Command("./a.out")
}
