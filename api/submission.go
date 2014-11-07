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

func ServeSubmissionList(w http.ResponseWriter, r *http.Request) {
	me, _ := context.Get(r, "me").(*data.Account)

	switch {
	case me == nil:
		http.Error(w, "", http.StatusForbidden)
		return

	case me.Level == data.Judge, me.Level == data.Administrator:
		cursor, err := strconv.ParseInt(r.FormValue("cursor"), 10, 64)
		catch(err)
		subms, err := data.ListSubmissions(cursor)
		catch(err)
		err = json.NewEncoder(w).Encode(subms)
		catch(err)

	default:
		subms, err := data.ListSubmissionsByAuthor(me)
		catch(err)
		err = json.NewEncoder(w).Encode(subms)
		catch(err)
	}
}

func CreateSubmission(w http.ResponseWriter, r *http.Request) {
	me, _ := context.Get(r, "me").(*data.Account)
	if me == nil {
		http.Error(w, "", http.StatusForbidden)
		return
	}

	body := struct {
		ProblemId int64  `json:"problemId"`
		Language  string `json:"language"`
		Source    string `json:"source"`
	}{}
	err := json.NewDecoder(r.Body).Decode(&body)
	catch(err)

	prob, err := data.GetProblem(body.ProblemId)
	catch(err)

	key, err := data.Blobs.Put("", strings.NewReader(body.Source))
	catch(err)

	subm := &data.Submission{
		AuthorId:  me.Id,
		ProblemId: prob.Id,
		Language:  body.Language,
		SourceKey: string(key),
	}
	err = subm.Put()
	catch(err)

	err = json.NewEncoder(w).Encode(subm)
	catch(err)
	hub.Send([]interface{}{"SYNC", "submissions"})

	err = data.NewActivity(me, fmt.Sprintf("created submission %d", subm.Id)).Put()
	catch(err)
	hub.Send([]interface{}{"SYNC", "activities"})

	if prob.Judge == "automatic" {
		exec := &data.Execution{
			SubmissionId: subm.Id,
			Apply:        true,
		}
		err = exec.Put()
		catch(err)
		belt.Push(exec)
	}
}

func ServeSubmission(w http.ResponseWriter, r *http.Request) {
	me, _ := context.Get(r, "me").(*data.Account)
	if me == nil {
		http.Error(w, "", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)

	id, err := strconv.ParseInt(vars["id"], 10, 64)
	catch(err)
	subm, err := data.GetSubmission(id)
	catch(err)

	if me.Level != data.Judge && me.Level != data.Administrator && me.Id != subm.AuthorId {
		http.Error(w, "", http.StatusForbidden)
		return
	}

	err = json.NewEncoder(w).Encode(subm)
	catch(err)
}

func UpdateSubmission(w http.ResponseWriter, r *http.Request) {
	me, _ := context.Get(r, "me").(*data.Account)
	if me == nil || (me.Level != data.Judge && me.Level != data.Administrator) {
		http.Error(w, "", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)

	id, err := strconv.ParseInt(vars["id"], 10, 64)
	catch(err)
	subm, err := data.GetSubmission(id)
	catch(err)

	body := struct {
		Verdict data.Verdict `json:"verdict"`
	}{}
	err = json.NewDecoder(r.Body).Decode(&body)
	catch(err)

	subm.Tamper(body.Verdict)
	err = subm.Put()
	catch(err)

	err = json.NewEncoder(w).Encode(subm)
	catch(err)
	hub.Send([]interface{}{"SYNC", "submissions", subm.Id})

	err = data.NewActivity(me, fmt.Sprintf("updated submission %d", subm.Id)).Put()
	catch(err)
	hub.Send([]interface{}{"SYNC", "activities"})
}

func ServeSubmissionSource(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	id, err := strconv.ParseInt(vars["id"], 10, 64)
	catch(err)
	subm, err := data.GetSubmission(id)
	catch(err)

	w.Header().Add("Content-Type", "text/plain")

	if r.FormValue("download") == "yes" {
		acc, err := data.GetAccount(subm.AuthorId)
		catch(err)

		prob, err := data.GetProblem(subm.ProblemId)
		catch(err)

		w.Header().Add("Content-Disposition", fmt.Sprintf(`attachment; filename="%d-%s-%s.%s"`, subm.Id, acc.Handle, strings.ToLower(prob.Char), subm.Language))
	}

	blob, err := data.Blobs.Get(subm.SourceKey)
	catch(err)
	_, err = io.Copy(w, blob)
	catch(err)
	err = blob.Close()
	catch(err)
}

func ServeSubmissionTestOutput(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	id, err := strconv.ParseInt(vars["id"], 10, 64)
	catch(err)
	subm, err := data.GetSubmission(id)
	catch(err)

	no, err := strconv.ParseInt(vars["no"], 10, 64)
	catch(err)
	test := subm.Tests[no-1]

	w.Header().Add("Content-Type", "text/plain")

	if r.FormValue("download") == "yes" {
		acc, err := data.GetAccount(subm.AuthorId)
		catch(err)

		prob, err := data.GetProblem(subm.ProblemId)
		catch(err)

		w.Header().Add("Content-Disposition", fmt.Sprintf(`attachment; filename="%d-%s-%s-%d.out"`, subm.Id, acc.Handle, strings.ToLower(prob.Char), no))
	}

	blob, err := data.Blobs.Get(test.OutputKey)
	catch(err)
	_, err = io.Copy(w, blob)
	catch(err)
	err = blob.Close()
	catch(err)
}

func ResetSubmission(w http.ResponseWriter, r *http.Request) {
	me, _ := context.Get(r, "me").(*data.Account)
	if me == nil || (me.Level != data.Judge && me.Level != data.Administrator) {
		http.Error(w, "", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)

	id, err := strconv.ParseInt(vars["id"], 10, 64)
	catch(err)
	subm, err := data.GetSubmission(id)
	catch(err)

	subm.Reset()
	err = subm.Put()
	catch(err)
	hub.Send([]interface{}{"SYNC", "submissions", subm.Id})

	err = data.NewActivity(me, fmt.Sprintf("reseted submission %d", subm.Id)).Put()
	catch(err)
	hub.Send([]interface{}{"SYNC", "activities"})
}

func JudgeSubmission(w http.ResponseWriter, r *http.Request) {
	me, _ := context.Get(r, "me").(*data.Account)
	if me == nil || (me.Level != data.Judge && me.Level != data.Administrator) {
		http.Error(w, "", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)

	id, err := strconv.ParseInt(vars["id"], 10, 64)
	catch(err)
	subm, err := data.GetSubmission(id)
	catch(err)

	subm.Reset()
	err = subm.Put()
	catch(err)
	hub.Send([]interface{}{"SYNC", "submissions", subm.Id})

	exec := &data.Execution{
		SubmissionId: subm.Id,
		Apply:        true,
	}
	err = exec.Put()
	catch(err)

	belt.Push(exec)

	err = data.NewActivity(me, fmt.Sprintf("judged submission %d", subm.Id)).Put()
	catch(err)
	hub.Send([]interface{}{"SYNC", "activities"})
}
