// Copyright 2014 The Cactus Authors. All rights reserved.

package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/context"

	"github.com/hjr265/cactus/data"
)

func ServeActivityList(w http.ResponseWriter, r *http.Request) {
	me, _ := context.Get(r, "me").(*data.Account)
	if me == nil || (me.Level != data.Judge && me.Level != data.Administrator) {
		http.Error(w, "", http.StatusForbidden)
		return
	}

	cursor, err := strconv.ParseInt(r.FormValue("cursor"), 10, 64)
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	subms, err := data.ListActivities(cursor)
	catch(err)

	err = json.NewEncoder(w).Encode(subms)
	catch(err)
}
