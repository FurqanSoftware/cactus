// Copyright 2014 The Cactus Authors. All rights reserved.

package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"

	"github.com/hjr265/cactus/data"
	"github.com/hjr265/cactus/hub"
)

func ServeProblemList(w http.ResponseWriter, r *http.Request) {
	cnt, err := data.GetContest()
	catch(err)

	me, _ := context.Get(r, "me").(*data.Account)
	if !cnt.Started() && (me == nil || (me.Level != data.Judge && me.Level != data.Administrator)) {
		err = json.NewEncoder(w).Encode([]*data.Problem{})
		catch(err)
		return
	}

	probs, err := data.ListProblems()
	catch(err)
	err = json.NewEncoder(w).Encode(probs)
	catch(err)
}

func CreateProblem(w http.ResponseWriter, r *http.Request) {
	me, _ := context.Get(r, "me").(*data.Account)
	if me == nil || me.Level != data.Administrator {
		http.Error(w, "", http.StatusForbidden)
		return
	}

	body := struct {
		Char  string `json:"char"`
		Title string `json:"title"`
	}{}
	err := json.NewDecoder(r.Body).Decode(&body)
	catch(err)

	switch {
	case len(body.Char) != 1 || !strings.Contains("abcdefghijklmnopqrstuvwxyz", body.Char):
		http.Error(w, "", http.StatusBadRequest)
		return

	case body.Title == "":
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	prob := &data.Problem{
		Slug:    strings.Trim(regexp.MustCompile("[^a-z0-9]+").ReplaceAllString(strings.ToLower(body.Char+" "+body.Title), "-"), " -"),
		Char:    body.Char,
		Title:   body.Title,
		Judge:   "automatic",
		Scoring: "strict",
	}
	err = prob.Put()
	catch(err)

	json.NewEncoder(w).Encode(prob)
	hub.Send([]interface{}{"SYNC", "problems"})

	err = data.NewActivity(me, fmt.Sprintf("created problem %d", prob.Id)).Put()
	catch(err)
	hub.Send([]interface{}{"SYNC", "activities"})
}

func ServeProblemBySlug(w http.ResponseWriter, r *http.Request) {
	cnt, err := data.GetContest()
	catch(err)

	me, _ := context.Get(r, "me").(*data.Account)
	if !cnt.Started() && (me == nil || (me.Level != data.Judge && me.Level != data.Administrator)) {
		http.Error(w, "", http.StatusNotFound)
		return
	}

	prob, err := data.GetProblemBySlug(r.FormValue("slug"))
	catch(err)

	err = json.NewEncoder(w).Encode(prob)
	catch(err)
}

func ServeProblem(w http.ResponseWriter, r *http.Request) {
	cnt, err := data.GetContest()
	catch(err)

	me, _ := context.Get(r, "me").(*data.Account)
	if !cnt.Started() && (me == nil || (me.Level != data.Judge && me.Level != data.Administrator)) {
		http.Error(w, "", http.StatusNotFound)
		return
	}

	vars := mux.Vars(r)

	id, err := strconv.ParseInt(vars["id"], 10, 64)
	catch(err)
	prob, err := data.GetProblem(id)
	catch(err)

	err = json.NewEncoder(w).Encode(prob)
	catch(err)
}

func UpdateProblem(w http.ResponseWriter, r *http.Request) {
	me, _ := context.Get(r, "me").(*data.Account)
	if me == nil || me.Level != data.Administrator {
		http.Error(w, "", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)

	id, err := strconv.ParseInt(vars["id"], 10, 64)
	catch(err)
	prob, err := data.GetProblem(id)
	catch(err)

	if prob == nil {
		http.Error(w, "", http.StatusNotFound)
	}

	body := struct {
		Char      string `json:"char"`
		Title     string `json:"title"`
		Statement struct {
			Body   string `json:"body"`
			Input  string `json:"input"`
			Output string `json:"output"`
		} `json:"statement"`
		Samples []struct {
			Input  string `json:"input"`
			Answer string `json:"answer"`
		} `json:"samples"`
		Notes   string `json:"notes"`
		Judge   string `json:"judge"`
		Checker struct {
			Language  string `json:"language"`
			Source    string `json:"source"`
			SourceKey string `json:"sourceKey"`
		} `json:"checker"`
		Limits struct {
			Cpu    float64 `json:"cpu"`
			Memory int     `json:"memory"`
		} `json:"limits"`
		Languages []string `json:"languages"`
		Tests     []struct {
			Input     string `json:"input"`
			InputKey  string `json:"inputKey"`
			Answer    string `json:"answer"`
			AnswerKey string `json:"answerKey"`
			Points    int    `json:"points"`
		} `json:"tests"`
		Scoring string `json:"scoring"`
	}{}
	err = json.NewDecoder(r.Body).Decode(&body)
	catch(err)

	switch {
	case len(body.Char) != 1 || !strings.Contains("abcdefghijklmnopqrstuvwxyz", body.Char):
		http.Error(w, "", http.StatusBadRequest)
		return

	case body.Title == "":
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	prob.Slug = strings.Trim(regexp.MustCompile("[^a-z0-9]+").ReplaceAllString(strings.ToLower(body.Char+" "+body.Title), "-"), " -")
	prob.Char = body.Char
	prob.Title = body.Title
	prob.Statement = body.Statement
	prob.Samples = body.Samples
	prob.Notes = body.Notes
	prob.Judge = body.Judge
	if body.Checker.Language != "" {
		if body.Checker.SourceKey == "" {
			body.Checker.SourceKey = fmt.Sprintf("problems:%d:checker:source", prob.Id)
			_, err := data.Blobs.Put(body.Checker.SourceKey, strings.NewReader(body.Checker.Source))
			catch(err)
		}
		prob.Checker = struct {
			Language  string `json:"language"`
			SourceKey string `json:"sourceKey"`
		}{
			Language:  body.Checker.Language,
			SourceKey: body.Checker.SourceKey,
		}

	} else {
		prob.Checker = struct {
			Language  string `json:"language"`
			SourceKey string `json:"sourceKey"`
		}{
			Language:  "",
			SourceKey: "",
		}
	}
	prob.Limits = body.Limits
	prob.Languages = body.Languages
	prob.Tests = nil
	for i, test := range body.Tests {
		if test.InputKey == "" {
			test.InputKey = fmt.Sprintf("problems:%d:tests:%d:in", prob.Id, i)
			_, err := data.Blobs.Put(test.InputKey, strings.NewReader(test.Input))
			catch(err)
		}

		if test.AnswerKey == "" {
			test.AnswerKey = fmt.Sprintf("problems:%d:tests:%d:ans", prob.Id, i)
			_, err := data.Blobs.Put(test.AnswerKey, strings.NewReader(test.Answer))
			catch(err)
		}

		prob.Tests = append(prob.Tests, struct {
			InputKey  string `json:"inputKey"`
			AnswerKey string `json:"answerKey"`
			Points    int    `json:"points"`
		}{
			InputKey:  test.InputKey,
			AnswerKey: test.AnswerKey,
			Points:    test.Points,
		})
	}
	prob.Scoring = body.Scoring
	err = prob.Put()
	catch(err)

	err = json.NewEncoder(w).Encode(prob)
	catch(err)
	hub.Send([]interface{}{"SYNC", "problems", prob.Id})

	err = data.NewActivity(me, fmt.Sprintf("updated problem %d", prob.Id)).Put()
	catch(err)
	hub.Send([]interface{}{"SYNC", "activities"})
}

func DeleteProblem(w http.ResponseWriter, r *http.Request) {
	me, _ := context.Get(r, "me").(*data.Account)
	if me == nil || me.Level != data.Administrator {
		http.Error(w, "", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)

	id, err := strconv.ParseInt(vars["id"], 10, 64)
	catch(err)
	prob, err := data.GetProblem(id)
	catch(err)

	err = prob.Del()
	catch(err)

	err = json.NewEncoder(w).Encode(&struct {
		Id int64 `json:"id"`
	}{
		Id: prob.Id,
	})
	catch(err)
	hub.Send([]interface{}{"SYNC", "problems"})

	err = data.NewActivity(me, fmt.Sprintf("deleted problem %d", prob.Id)).Put()
	catch(err)
	hub.Send([]interface{}{"SYNC", "activities"})
}

func ServeProblemTestAnswer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	id, err := strconv.ParseInt(vars["id"], 10, 64)
	catch(err)
	prob, err := data.GetProblem(id)
	catch(err)

	no, err := strconv.ParseInt(vars["no"], 10, 64)
	catch(err)
	test := prob.Tests[no-1]

	w.Header().Add("Content-Type", "text/plain")

	if r.FormValue("download") == "yes" {
		w.Header().Add("Content-Disposition", fmt.Sprintf(`attachment; filename="%s-%d.ans"`, prob.Slug, no))
	}

	blob, err := data.Blobs.Get(test.AnswerKey)
	catch(err)
	_, err = io.Copy(w, blob)
	catch(err)
	err = blob.Close()
	catch(err)
}
