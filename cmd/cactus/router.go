// Copyright 2014 The Cactus Authors. All rights reserved.

package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/context"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"github.com/hjr265/cactus/api"
	"github.com/hjr265/cactus/data"
	"github.com/hjr265/cactus/hub"
	"github.com/hjr265/cactus/ui"
)

func handlePanic(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Print(err)
				http.Error(w, "", http.StatusInternalServerError)
			}
		}()

		h.ServeHTTP(w, r)
	})
}

func handleIdentity(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := api.Store.Get(r, "s")
		if err != nil {
			r.Header.Del("Cookie")
			context.Clear(r)
		}

		sess, err := api.Store.Get(r, "s")
		catch(err)

		id, _ := sess.Values["me.id"].(int64)
		acc, err := data.GetAccount(id)
		catch(err)
		context.Set(r, "me", acc)

		h.ServeHTTP(w, r)
	})
}

func init() {
	r := mux.NewRouter()

	r.NewRoute().
		Methods("POST").
		Path("/api/login").
		Handler(handleIdentity(http.HandlerFunc(api.HandleLogin)))
	r.NewRoute().
		Methods("POST").
		Path("/api/logout").
		Handler(handleIdentity(http.HandlerFunc(api.HandleLogout)))

	r.NewRoute().
		Methods("GET").
		Path("/api/accounts").
		Handler(handleIdentity(http.HandlerFunc(api.ServeAccountList)))
	r.NewRoute().
		Methods("POST").
		Path("/api/accounts").
		Handler(handleIdentity(http.HandlerFunc(api.CreateAccount)))
	r.NewRoute().
		Methods("POST").
		Path("/api/accounts/import").
		Handler(handleIdentity(http.HandlerFunc(api.ImportAccounts)))
	r.NewRoute().
		Methods("GET").
		Path("/api/accounts/me").
		Handler(handleIdentity(http.HandlerFunc(api.ServeAccountMe)))
	r.NewRoute().
		Methods("GET").
		Path("/api/accounts/by_handle").
		Handler(handleIdentity(http.HandlerFunc(api.ServeAccountByHandle)))
	r.NewRoute().
		Methods("GET").
		Path("/api/accounts/{id}").
		Handler(handleIdentity(http.HandlerFunc(api.ServeAccount)))
	r.NewRoute().
		Methods("PUT").
		Path("/api/accounts/{id}").
		Handler(handleIdentity(http.HandlerFunc(api.UpdateAccount)))
	r.NewRoute().
		Methods("PATCH").
		Path("/api/accounts/{id}").
		Handler(handleIdentity(http.HandlerFunc(api.UpdateAccountPart)))
	r.NewRoute().
		Methods("DELETE").
		Path("/api/accounts/{id}").
		Handler(handleIdentity(http.HandlerFunc(api.DeleteAccount)))

	r.NewRoute().
		Methods("GET").
		Path("/api/contests/1").
		Handler(handleIdentity(http.HandlerFunc(api.ServeContest)))
	r.NewRoute().
		Methods("PUT").
		Path("/api/contests/1").
		Handler(handleIdentity(http.HandlerFunc(api.UpdateContest)))

	r.NewRoute().
		Methods("GET").
		Path("/api/standings").
		Handler(handleIdentity(http.HandlerFunc(api.ServeStandingList)))

	r.NewRoute().
		Methods("GET").
		Path("/api/clarifications").
		Handler(handleIdentity(http.HandlerFunc(api.ServeClarificationList)))
	r.NewRoute().
		Methods("POST").
		Path("/api/clarifications").
		Handler(handleIdentity(http.HandlerFunc(api.CreateClarification)))
	r.NewRoute().
		Methods("GET").
		Path("/api/clarifications/{id}").
		Handler(handleIdentity(http.HandlerFunc(api.ServeClarification)))
	r.NewRoute().
		Methods("PUT").
		Path("/api/clarifications/{id}").
		Handler(handleIdentity(http.HandlerFunc(api.UpdateClarification)))
	r.NewRoute().
		Methods("DELETE").
		Path("/api/clarifications/{id}").
		Handler(handleIdentity(http.HandlerFunc(api.DeleteClarification)))

	r.NewRoute().
		Methods("GET").
		Path("/api/problems").
		Handler(handleIdentity(http.HandlerFunc(api.ServeProblemList)))
	r.NewRoute().
		Methods("POST").
		Path("/api/problems").
		Handler(handleIdentity(http.HandlerFunc(api.CreateProblem)))
	r.NewRoute().
		Methods("GET").
		Path("/api/problems/by_slug").
		Handler(handleIdentity(http.HandlerFunc(api.ServeProblemBySlug)))
	r.NewRoute().
		Methods("GET").
		Path("/api/problems/{id}").
		Handler(handleIdentity(http.HandlerFunc(api.ServeProblem)))
	r.NewRoute().
		Methods("PUT").
		Path("/api/problems/{id}").
		Handler(handleIdentity(http.HandlerFunc(api.UpdateProblem)))
	r.NewRoute().
		Methods("DELETE").
		Path("/api/problems/{id}").
		Handler(handleIdentity(http.HandlerFunc(api.DeleteProblem)))
	r.NewRoute().
		Methods("GET").
		Path("/api/problems/{id}/tests/{no}/answer").
		Handler(handleIdentity(http.HandlerFunc(api.ServeProblemTestAnswer)))

	r.NewRoute().
		Methods("GET").
		Path("/api/submissions").
		Handler(handleIdentity(http.HandlerFunc(api.ServeSubmissionList)))
	r.NewRoute().
		Methods("POST").
		Path("/api/submissions").
		Handler(handleIdentity(http.HandlerFunc(api.CreateSubmission)))
	r.NewRoute().
		Methods("GET").
		Path("/api/submissions/{id}").
		Handler(handleIdentity(http.HandlerFunc(api.ServeSubmission)))
	r.NewRoute().
		Methods("PUT").
		Path("/api/submissions/{id}").
		Handler(handleIdentity(http.HandlerFunc(api.UpdateSubmission)))
	r.NewRoute().
		Methods("GET").
		Path("/api/submissions/{id}/source").
		Handler(handleIdentity(http.HandlerFunc(api.ServeSubmissionSource)))
	r.NewRoute().
		Methods("GET").
		Path("/api/submissions/{id}/tests/{no}/output").
		Handler(handleIdentity(http.HandlerFunc(api.ServeSubmissionTestOutput)))
	r.NewRoute().
		Methods("POST").
		Path("/api/submissions/{id}/reset").
		Handler(handleIdentity(http.HandlerFunc(api.ResetSubmission)))
	r.NewRoute().
		Methods("POST").
		Path("/api/submissions/{id}/judge").
		Handler(handleIdentity(http.HandlerFunc(api.JudgeSubmission)))

	r.NewRoute().
		Methods("POST").
		Path("/api/executions").
		Handler(handleIdentity(http.HandlerFunc(api.CreateExecution)))
	r.NewRoute().
		Methods("GET").
		Path("/api/executions/{id}").
		Handler(handleIdentity(http.HandlerFunc(api.ServeExecution)))
	r.NewRoute().
		Methods("POST").
		Path("/api/executions/{id}/apply").
		Handler(handleIdentity(http.HandlerFunc(api.ApplyExecution)))
	r.NewRoute().
		Methods("GET").
		Path("/api/executions/{id}/tests/{no}/output").
		Handler(handleIdentity(http.HandlerFunc(api.ServeExecutionTestOutput)))

	r.NewRoute().
		Methods("GET").
		Path("/api/activities").
		Handler(handleIdentity(http.HandlerFunc(api.ServeActivityList)))

	r.NewRoute().
		Methods("GET").
		Path("/api/notifications").
		Handler(handleIdentity(http.HandlerFunc(api.ServeNotificationList)))

	r.NewRoute().
		Methods("GET").
		PathPrefix("/api").
		Handler(http.NotFoundHandler())

	r.NewRoute().
		Methods("GET").
		PathPrefix("/assets").
		Handler(http.StripPrefix("/assets", http.HandlerFunc(ui.ServeAsset)))
	r.NewRoute().
		Methods("GET").
		PathPrefix("/").
		Handler(handleIdentity(http.HandlerFunc(ui.ServeIndex)))

	http.Handle("/hub", handlePanic(handleIdentity(http.HandlerFunc(hub.HandleConnect))))
	http.Handle("/", handlePanic(http.TimeoutHandler(handlers.CompressHandler(r), 8*time.Second, "")))
}
