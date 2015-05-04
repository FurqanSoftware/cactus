// Copyright 2014 The Cactus Authors. All rights reserved.

package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"

	"github.com/hjr265/cactus/data"
	"github.com/hjr265/cactus/hub"
)

func ServeLanguageList(w http.ResponseWriter, r *http.Request) {
	langs, err := data.ListLanguages()
	catch(err)

	err = json.NewEncoder(w).Encode(langs)
	catch(err)
}

func CreateLanguage(w http.ResponseWriter, r *http.Request) {
	me, _ := context.Get(r, "me").(*data.Account)
	if me == nil || me.Level != data.Administrator {
		http.Error(w, "", http.StatusForbidden)
		return
	}

	body := struct {
		Label string `json:"label"`
		Steps struct {
			Build string `json:"build"`
			Run   string `json:"run"`
		} `json:"steps"`
	}{}
	err := json.NewDecoder(r.Body).Decode(&body)
	catch(err)

	lang := &data.Language{}
	lang.Label = body.Label
	lang.Steps = body.Steps
	err = lang.Put()
	catch(err)

	err = json.NewEncoder(w).Encode(lang)
	catch(err)
	hub.Send([]interface{}{"SYNC", "languages"})

	err = data.NewActivity(me, fmt.Sprintf("created language %d", lang.Id)).Put()
	catch(err)
	hub.Send([]interface{}{"SYNC", "activities"})
}

func ServeLanguage(w http.ResponseWriter, r *http.Request) {
	me, _ := context.Get(r, "me").(*data.Account)
	if me == nil || me.Level != data.Administrator {
		http.Error(w, "", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)

	id, err := strconv.ParseInt(vars["id"], 10, 64)
	catch(err)
	lang, err := data.GetLanguage(id)
	catch(err)

	err = json.NewEncoder(w).Encode(lang)
	catch(err)
}

func UpdateLanguage(w http.ResponseWriter, r *http.Request) {
	me, _ := context.Get(r, "me").(*data.Account)
	if me == nil || me.Level != data.Administrator {
		http.Error(w, "", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)

	id, err := strconv.ParseInt(vars["id"], 10, 64)
	catch(err)
	lang, err := data.GetLanguage(id)
	catch(err)

	body := struct {
		Label string `json:"label"`
		Steps struct {
			Build string `json:"build"`
			Run   string `json:"run"`
		} `json:"steps"`
	}{}
	err = json.NewDecoder(r.Body).Decode(&body)
	catch(err)

	lang.Label = body.Label
	lang.Steps = body.Steps
	err = lang.Put()
	catch(err)

	json.NewEncoder(w).Encode(lang)
	hub.Send([]interface{}{"SYNC", "languages", lang.Id})

	err = data.NewActivity(me, fmt.Sprintf("updated language %d", lang.Id)).Put()
	catch(err)
	hub.Send([]interface{}{"SYNC", "activities"})
}

func DeleteLanguage(w http.ResponseWriter, r *http.Request) {
	me, _ := context.Get(r, "me").(*data.Account)
	if me == nil || me.Level != data.Administrator {
		http.Error(w, "", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)

	id, err := strconv.ParseInt(vars["id"], 10, 64)
	catch(err)
	lang, err := data.GetLanguage(id)
	catch(err)

	err = lang.Del()
	catch(err)

	json.NewEncoder(w).Encode(&struct {
		Id int64 `json:"id"`
	}{
		Id: lang.Id,
	})
	hub.Send([]interface{}{"SYNC", "languages"})

	err = data.NewActivity(me, fmt.Sprintf("deleted language %d", lang.Id)).Put()
	catch(err)
	hub.Send([]interface{}{"SYNC", "activities"})
}
