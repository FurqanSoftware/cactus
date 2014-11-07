// Copyright 2014 The Cactus Authors. All rights reserved.

package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/context"

	"github.com/hjr265/cactus/data"
	"github.com/hjr265/cactus/hub"
)

func ServeContest(w http.ResponseWriter, r *http.Request) {
	cnt, err := data.GetContest()
	catch(err)

	err = json.NewEncoder(w).Encode(cnt)
	catch(err)
}

func UpdateContest(w http.ResponseWriter, r *http.Request) {
	me, _ := context.Get(r, "me").(*data.Account)
	if me == nil || me.Level != data.Administrator {
		http.Error(w, "", http.StatusForbidden)
		return
	}

	cnt, err := data.GetContest()
	catch(err)

	err = json.NewDecoder(r.Body).Decode(cnt)
	catch(err)

	err = cnt.Put()
	catch(err)

	err = json.NewEncoder(w).Encode(cnt)
	catch(err)
	hub.Send([]interface{}{"SYNC", "contest"})

	err = data.NewActivity(me, fmt.Sprintf("updated contest %d", cnt.Id)).Put()
	catch(err)
	hub.Send([]interface{}{"SYNC", "activities"})
}
