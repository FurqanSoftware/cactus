// Copyright 2014 The Cactus Authors. All rights reserved.

package api

import (
	"github.com/gorilla/sessions"

	"github.com/hjr265/cactus/data"
)

var Store sessions.Store

func init() {
	cnt, err := data.GetContest()
	catch(err)
	Store = sessions.NewCookieStore(cnt.Salt, cnt.Salt)
}
