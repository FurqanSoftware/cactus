// Copyright 2014 The Cactus Authors. All rights reserved.

package cube

type Process struct {
	Name string
	Args []string

	Stdin  []byte
	Stdout []byte
	Stderr []byte

	Limits Limits
	Usages Limits

	Success bool

	cube Cube
}

type Limits struct {
	Time   float64
	Cpu    float64
	Memory int
}
