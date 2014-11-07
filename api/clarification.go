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

func ServeClarificationList(w http.ResponseWriter, r *http.Request) {
	me, _ := context.Get(r, "me").(*data.Account)

	var (
		clars []*data.Clarification
		err   error
	)
	switch {
	case me == nil:
		http.Error(w, "", http.StatusForbidden)
		return

	case me.Level == data.Judge, me.Level == data.Administrator:
		clars, err = data.ListClarifications()
		catch(err)

	default:
		clars, err = data.ListClarificationsForAccount(me)
		catch(err)
	}

	err = json.NewEncoder(w).Encode(clars)
	catch(err)
}

func CreateClarification(w http.ResponseWriter, r *http.Request) {
	me, _ := context.Get(r, "me").(*data.Account)

	body := struct {
		ProblemId int64  `json:"problemId"`
		Question  string `json:"question"`
	}{}
	err := json.NewDecoder(r.Body).Decode(&body)
	catch(err)

	clar := &data.Clarification{
		AskerId:  me.Id,
		Question: body.Question,
	}

	prob, err := data.GetProblem(body.ProblemId)
	catch(err)
	if prob != nil {
		clar.ProblemId = prob.Id
	}

	err = clar.Put()
	catch(err)

	err = json.NewEncoder(w).Encode(clar)
	catch(err)
	hub.Send([]interface{}{"SYNC", "clarifications"})

	err = data.NewActivity(me, fmt.Sprintf("created clarification %d", clar.Id)).Put()
	catch(err)
	hub.Send([]interface{}{"SYNC", "activities"})

	err = data.NewNotification(0, data.Judge, fmt.Sprintf("Clarification %d requested", clar.Id)).Put()
	catch(err)
	err = data.NewNotification(0, data.Administrator, fmt.Sprintf("Clarification %d requested", clar.Id)).Put()
	catch(err)
	hub.Send([]interface{}{"SYNC", "notifications"})
}

func ServeClarification(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	id, err := strconv.ParseInt(vars["id"], 10, 64)
	catch(err)
	clar, err := data.GetClarification(id)
	catch(err)

	err = json.NewEncoder(w).Encode(clar)
	catch(err)
}

func UpdateClarification(w http.ResponseWriter, r *http.Request) {
	me, _ := context.Get(r, "me").(*data.Account)
	if me == nil || (me.Level != data.Judge && me.Level != data.Administrator) {
		http.Error(w, "", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)

	id, err := strconv.ParseInt(vars["id"], 10, 64)
	catch(err)
	clar, err := data.GetClarification(id)
	catch(err)

	body := struct {
		ProblemId int64         `json:"problemId"`
		Question  string        `json:"question"`
		Response  data.Response `json:"response"`
		Message   string        `json:"message"`
	}{}
	err = json.NewDecoder(r.Body).Decode(&body)
	catch(err)

	prob, err := data.GetProblem(body.ProblemId)
	catch(err)
	if prob != nil {
		clar.ProblemId = prob.Id
	} else {
		clar.ProblemId = 0
	}

	clar.Question = body.Question
	clar.Response = body.Response
	clar.Message = body.Message

	err = clar.Put()
	catch(err)

	err = json.NewEncoder(w).Encode(clar)
	catch(err)
	hub.Send([]interface{}{"SYNC", "clarifications", clar.Id})

	err = data.NewActivity(me, fmt.Sprintf("updated clarification %d", clar.Id)).Put()
	catch(err)
	hub.Send([]interface{}{"SYNC", "activities"})

	if clar.Response != data.Unresponded {
		if clar.Response == data.Broadcasted {
			err = data.NewNotification(0, data.Participant, fmt.Sprintf("Clarification %d updated", clar.Id)).Put()
			catch(err)
		} else {
			err = data.NewNotification(clar.AskerId, 0, fmt.Sprintf("Clarification %d updated", clar.Id)).Put()
			catch(err)
		}
		hub.Send([]interface{}{"SYNC", "notifications"})
	}
}

func DeleteClarification(w http.ResponseWriter, r *http.Request) {
	me, _ := context.Get(r, "me").(*data.Account)
	if me == nil || (me.Level != data.Judge && me.Level != data.Administrator) {
		http.Error(w, "", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)

	id, err := strconv.ParseInt(vars["id"], 10, 64)
	catch(err)
	clar, err := data.GetClarification(id)
	catch(err)

	err = clar.Del()
	catch(err)

	json.NewEncoder(w).Encode(&struct {
		Id int64 `json:"id"`
	}{
		Id: clar.Id,
	})
	hub.Send([]interface{}{"SYNC", "clarifications"})

	err = data.NewActivity(me, fmt.Sprintf("deleted clarification %d", clar.Id)).Put()
	catch(err)
	hub.Send([]interface{}{"SYNC", "activities"})
}
