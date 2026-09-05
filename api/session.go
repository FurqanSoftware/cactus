// Copyright 2014 The Cactus Authors. All rights reserved.

package api

import (
	"context"
	"net/http"

	"github.com/gorilla/sessions"

	"github.com/FurqanSoftware/cactus/data"
)

type contextKey string

// currentAccountKey is the context key used to store the authenticated account.
var currentAccountKey = contextKey("me")

// WithCurrentAccount returns a new request with the given account stored in its context.
func WithCurrentAccount(r *http.Request, acc *data.Account) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), currentAccountKey, acc))
}

// currentAccount returns the authenticated account from the request context, or nil.
func currentAccount(r *http.Request) *data.Account {
	acc, _ := r.Context().Value(currentAccountKey).(*data.Account)
	return acc
}

var Store sessions.Store

func InitStore() error {
	cnt, err := data.GetContest()
	if err != nil {
		return err
	}
	Store = sessions.NewCookieStore(cnt.Salt, cnt.Salt)
	return nil
}
