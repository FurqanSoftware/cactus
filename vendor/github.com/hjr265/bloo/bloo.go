// Copyright 2014 The Bloo Authors. All rights reserved.

package bloo

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
)

type BS struct {
	path string
}

func Open(path string, mode os.FileMode) (*BS, error) {
	err := os.Mkdir("cactus.bs", mode)
	if err != nil && !os.IsExist(err) {
		return nil, err
	}
	return &BS{
		path: path,
	}, nil
}

func (b *BS) Path() string {
	return b.path
}

func (b *BS) Get(key string) (io.ReadCloser, error) {
	sum := sha1.Sum([]byte(key))
	name := hex.EncodeToString(sum[:])
	f, err := os.Open(filepath.Join(b.path, name[:2], name))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return f, nil
}

func (b *BS) Put(key string, blob io.Reader) (string, error) {
	if key == "" {
		for i := 0; i < 1<<10; i++ {
			p := make([]byte, 16)
			_, err := rand.Read(p)
			if err != nil {
				return "", err
			}
			key = hex.EncodeToString(p)

			sum := sha1.Sum([]byte(key))
			name := hex.EncodeToString(sum[:])

			err = os.MkdirAll(filepath.Join(b.path, name[:2]), 0766)
			if err != nil {
				return "", err
			}

			f, err := os.OpenFile(filepath.Join(b.path, name[:2], name), os.O_CREATE|os.O_EXCL, 0666)
			if err != nil {
				if os.IsExist(err) {
					continue
				}
				return "", err
			}
			err = f.Close()
			if err != nil {
				return "", err
			}

			break
		}
	}

	sum := sha1.Sum([]byte(key))
	name := hex.EncodeToString(sum[:])

	err := os.MkdirAll(filepath.Join(b.path, name[:2]), 0766)
	if err != nil {
		return "", err
	}

	f, err := os.Create(filepath.Join(b.path, name[:2], name))
	if err != nil {
		return "", err
	}
	_, err = io.Copy(f, blob)
	if err != nil {
		return "", err
	}
	err = f.Close()
	if err != nil {
		return "", err
	}
	return key, nil
}
