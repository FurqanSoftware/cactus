// Copyright 2014 The Cactus Authors. All rights reserved.

package main

import (
	"github.com/pelletier/go-toml"
)

// Config holds the application configuration parsed from config.toml.
type Config struct {
	Core struct {
		Addr string `toml:"addr"`
	} `toml:"core"`
	Data struct {
		DB    string `toml:"db"`
		Blobs string `toml:"blobs"`
	} `toml:"data"`
	Belt struct {
		Size int `toml:"size"`
	} `toml:"belt"`
	Sandbox struct {
		Mode        string `toml:"mode"`
		DockerImage string `toml:"docker_image"`
	} `toml:"sandbox"`
}

func parseConfig(path string) (Config, error) {
	tree, err := toml.LoadFile(path)
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := tree.Unmarshal(&cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
