// Copyright 2014 The Cactus Authors. All rights reserved.

package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"

	"github.com/hjr265/cactus/belt"
	"github.com/hjr265/cactus/data"
	"github.com/hjr265/cactus/hub"
)

func CreateExecution(w http.ResponseWriter, r *http.Request) {
	me, _ := context.Get(r, "me").(*data.Account)
	if me == nil || (me.Level != data.Judge && me.Level != data.Administrator) {
		http.Error(w, "", http.StatusForbidden)
		return
	}

	body := struct {
		SubmissionId int64 `json:"submissionId"`
	}{}
	err := json.NewDecoder(r.Body).Decode(&body)
	catch(err)

	subm, err := data.GetSubmission(body.SubmissionId)
	catch(err)

	exec := &data.Execution{
		SubmissionId: subm.Id,
	}
	err = exec.Put()
	catch(err)

	belt.Push(exec)

	err = json.NewEncoder(w).Encode(exec)
	catch(err)

	err = data.NewActivity(me, fmt.Sprintf("created execution %d for submission %d", exec.Id, subm.Id)).Put()
	catch(err)
	hub.Send([]interface{}{"SYNC", "activities"})
}

func ServeExecution(w http.ResponseWriter, r *http.Request) {
	me, _ := context.Get(r, "me").(*data.Account)
	if me == nil || (me.Level != data.Judge && me.Level != data.Administrator) {
		http.Error(w, "", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)

	id, err := strconv.ParseInt(vars["id"], 10, 64)
	catch(err)
	exec, err := data.GetExecution(id)
	catch(err)

	err = json.NewEncoder(w).Encode(exec)
	catch(err)
}

func ApplyExecution(w http.ResponseWriter, r *http.Request) {
	me, _ := context.Get(r, "me").(*data.Account)
	if me == nil || (me.Level != data.Judge && me.Level != data.Administrator) {
		http.Error(w, "", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)

	id, err := strconv.ParseInt(vars["id"], 10, 64)
	catch(err)
	exec, err := data.GetExecution(id)
	catch(err)

	if exec.Status != 7 {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	subm, err := exec.Submission()
	catch(err)

	subm.Apply(exec)
	err = subm.Put()
	catch(err)
	hub.Send([]interface{}{"SYNC", "submissions", subm.Id})

	err = data.NewActivity(me, fmt.Sprintf("applied execution %d to submission %d", exec.Id, subm.Id)).Put()
	catch(err)
	hub.Send([]interface{}{"SYNC", "activities"})
}

func ServeExecutionTestOutput(w http.ResponseWriter, r *http.Request) {
	me, _ := context.Get(r, "me").(*data.Account)
	if me == nil || (me.Level != data.Judge && me.Level != data.Administrator) {
		http.Error(w, "", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)

	id, err := strconv.ParseInt(vars["id"], 10, 64)
	catch(err)
	exec, err := data.GetExecution(id)
	catch(err)

	no, err := strconv.ParseInt(vars["no"], 10, 64)
	catch(err)
	test := exec.Tests[no-1]

	w.Header().Add("Content-Type", "text/plain")

	if r.FormValue("download") == "yes" {
		subm, err := exec.Submission()
		catch(err)
		acc, err := data.GetAccount(subm.AuthorId)
		catch(err)
		prob, err := data.GetProblem(subm.ProblemId)
		catch(err)
		w.Header().Add("Content-Disposition", fmt.Sprintf(`attachment; filename="%d-%s-%s-%d-%d.out"`, subm.Id, acc.Handle, strings.ToLower(prob.Char), exec.Id, no))
	}

	blob, err := data.Blobs.Get(test.OutputKey)
	catch(err)
	_, err = io.Copy(w, blob)
	catch(err)
	err = blob.Close()
	catch(err)
}
