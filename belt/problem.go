// Copyright 2014 The Cactus Authors. All rights reserved.

package belt

import (
	"github.com/hjr265/cactus/data"
)

type Problem struct {
	*data.Problem
}

func (p *Problem) CheckerLanguage() (*data.Language, error) {
	switch Queue := Queue.(type) {
	case *LocalQueue:
		return p.Problem.CheckerLanguage()

	case *RemoteQueue:
		lang := data.Language{}
		err := Queue.c.Call("Languages.Get", p.Problem.Checker.LanguageId, &lang)
		if err != nil {
			return nil, err
		}
		return &lang, nil
	}

	panic("unreachable")
}
