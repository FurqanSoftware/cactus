// Copyright 2014 The Cactus Authors. All rights reserved.

package api

import (
	"bytes"
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

	stnd, err := data.GetStandingByAuthor(me)
	catch(err)

	probs2 := []struct {
		*data.Problem
		Attempt *data.Attempt `json:"attempt"`
	}{}
	for _, prob := range probs {
		probs2 = append(probs2, struct {
			*data.Problem
			Attempt *data.Attempt `json:"attempt"`
		}{
			Problem: prob,
			Attempt: stnd.Attempts[strconv.FormatInt(prob.Id, 10)],
		})
	}

	err = json.NewEncoder(w).Encode(probs2)
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
			LanguageId int64 `json:"languageId"`
			Source     struct {
				Key  string `json:"key"`
				Name string `json:"name"`
				Body []byte `json:"body"`
			} `json:"source"`
		} `json:"checker"`
		Limits struct {
			Cpu    float64 `json:"cpu"`
			Memory int     `json:"memory"`
			Source int     `json:"source"`
		} `json:"limits"`
		LanguageIds []int64 `json:"languageIds"`
		Tests       []struct {
			Input struct {
				Key  string `json:"key"`
				Name string `json:"name"`
				Body []byte `json:"body"`
			} `json:"input"`
			Answer struct {
				Key  string `json:"key"`
				Name string `json:"name"`
				Body []byte `json:"body"`
			} `json:"answer"`
			Points int `json:"points"`
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
	if body.Checker.LanguageId != 0 {
		if body.Checker.Source.Key == "" {
			body.Checker.Source.Key = fmt.Sprintf("problems:%d:checker:source", prob.Id)
			_, err := data.Blobs.Put(body.Checker.Source.Key, bytes.NewReader(body.Checker.Source.Body))
			catch(err)
			prob.Checker.Source.Name = body.Checker.Source.Name
		}
		prob.Checker.LanguageId = body.Checker.LanguageId
		prob.Checker.Source.Key = body.Checker.Source.Key
	} else {
		prob.Checker.LanguageId = 0
		prob.Checker.Source.Key = ""
	}
	prob.Limits = body.Limits
	prob.LanguageIds = body.LanguageIds
	names := map[string]string{}
	for _, test := range prob.Tests {
		names[test.Input.Key] = test.Input.Name
		names[test.Answer.Key] = test.Answer.Name
	}
	prob.Tests = nil
	for i, test := range body.Tests {
		test2 := struct {
			Input struct {
				Key  string `json:"key"`
				Name string `json:"name"`
			} `json:"input"`
			Answer struct {
				Key  string `json:"key"`
				Name string `json:"name"`
			} `json:"answer"`
			Points int `json:"points"`
		}{}
		if test.Input.Key == "" {
			test.Input.Key = fmt.Sprintf("problems:%d:tests:%d:in", prob.Id, i)
			_, err := data.Blobs.Put(test.Input.Key, bytes.NewReader(test.Input.Body))
			catch(err)
			test2.Input.Name = test.Input.Name
		} else {
			test2.Input.Name = names[test.Input.Key]
		}
		test2.Input.Key = test.Input.Key
		if test.Answer.Key == "" {
			test.Answer.Key = fmt.Sprintf("problems:%d:tests:%d:ans", prob.Id, i)
			_, err := data.Blobs.Put(test.Answer.Key, bytes.NewReader(test.Answer.Body))
			catch(err)
			test2.Answer.Name = test.Answer.Name
		} else {
			test2.Answer.Name = names[test.Answer.Key]
		}
		test2.Answer.Key = test.Answer.Key
		test2.Points = test.Points
		prob.Tests = append(prob.Tests, test2)
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
	me, _ := context.Get(r, "me").(*data.Account)
	if me == nil || (me.Level != data.Judge && me.Level != data.Administrator) {
		http.Error(w, "", http.StatusForbidden)
		return
	}

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

	blob, err := data.Blobs.Get(test.Answer.Key)
	catch(err)
	_, err = io.Copy(w, blob)
	catch(err)
	err = blob.Close()
	catch(err)
}
