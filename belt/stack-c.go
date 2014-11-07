// Copyright 2014 The Cactus Authors. All rights reserved.

package belt

import (
	"io"
	"time"

	"github.com/hjr265/jail.go/jail"
)

type StackC struct{}

func (s *StackC) Build(cell *jail.Cell, source io.Reader) (*jail.Cmd, error) {
	f, err := cell.Create("source.c")
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

	cmd := cell.Command("gcc", "source.c")
	cmd.Limits.Cpu = 16 * time.Second
	cmd.Limits.Memory = 1 << 30

	return cmd, nil
}

func (s *StackC) Run(cell *jail.Cell) *jail.Cmd {
	return cell.Command("./a.out")
}
