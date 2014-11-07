// Copyright 2014 The Cactus Authors. All rights reserved.

package api

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/mxk/go-sqlite/sqlite3"

	"github.com/hjr265/cactus/data"
	"github.com/hjr265/cactus/hub"
)

func ServeAccountList(w http.ResponseWriter, r *http.Request) {
	me, _ := context.Get(r, "me").(*data.Account)

	switch {
	case me == nil:
		err := json.NewEncoder(w).Encode([]*data.Account{})
		catch(err)

	case me.Level == data.Judge, me.Level == data.Administrator:
		accs, err := data.ListAccounts()
		catch(err)

		err = json.NewEncoder(w).Encode(accs)
		catch(err)

	default:
		err := json.NewEncoder(w).Encode([]*data.Account{
			me,
		})
		catch(err)
	}
}

func CreateAccount(w http.ResponseWriter, r *http.Request) {
	me, _ := context.Get(r, "me").(*data.Account)
	if me == nil || me.Level != data.Administrator {
		http.Error(w, "", http.StatusForbidden)
		return
	}

	body := struct {
		Handle   string     `json:"handle"`
		Password string     `json:"password"`
		Level    data.Level `json:"level"`
		Name     string     `json:"name"`
	}{}
	err := json.NewDecoder(r.Body).Decode(&body)
	catch(err)

	switch {
	case len(body.Handle) < 4:
		http.Error(w, "", http.StatusBadRequest)
		return

	case len(body.Password) < 4:
		http.Error(w, "", http.StatusBadRequest)
		return

	case body.Level != data.Participant && body.Level != data.Judge && body.Level != data.Administrator:
		http.Error(w, "", http.StatusBadRequest)
		return

	case body.Name == "":
		body.Name = body.Handle
	}

	acc := &data.Account{}
	acc.Handle = body.Handle
	err = acc.SetPassword(body.Password)
	catch(err)
	acc.Level = body.Level
	acc.Name = body.Name
	err = acc.Put()
	if err, ok := err.(*sqlite3.Error); ok && err.Code() == sqlite3.CONSTRAINT_UNIQUE {
		http.Error(w, "", http.StatusConflict)
		return
	}
	catch(err)

	err = json.NewEncoder(w).Encode(acc)
	catch(err)
	hub.Send([]interface{}{"SYNC", "accounts"})

	err = data.NewActivity(me, fmt.Sprintf("created account %d", acc.Id)).Put()
	catch(err)
	hub.Send([]interface{}{"SYNC", "activities"})
}

func ImportAccounts(w http.ResponseWriter, r *http.Request) {
	me, _ := context.Get(r, "me").(*data.Account)
	if me == nil || me.Level != data.Administrator {
		http.Error(w, "", http.StatusForbidden)
		return
	}

	body := csv.NewReader(r.Body)
	for {
		cols, err := body.Read()
		if err == io.EOF {
			break
		}
		catch(err)

		acc := &data.Account{}
		acc.Handle = cols[0]
		err = acc.SetPassword(cols[1])
		catch(err)
		col2, err := strconv.ParseInt(cols[2], 10, 32)
		catch(err)
		acc.Level = data.Level(col2)
		acc.Name = cols[3]

		switch {
		case len(acc.Handle) < 4:
			continue

		case len(acc.Password) < 4:
			continue

		case acc.Level != data.Participant && acc.Level != data.Judge && acc.Level != data.Administrator:
			continue

		case acc.Name == "":
			acc.Name = acc.Handle
		}

		err = acc.Put()
		if err, ok := err.(*sqlite3.Error); ok && err.Code() == sqlite3.CONSTRAINT_UNIQUE {
			continue
		}
		catch(err)

		err = data.NewActivity(me, fmt.Sprintf("created account %d", acc.Id)).Put()
		catch(err)
		hub.Send([]interface{}{"SYNC", "activities"})
	}
	hub.Send([]interface{}{"SYNC", "accounts"})
}

func ServeAccountMe(w http.ResponseWriter, r *http.Request) {
	me, _ := context.Get(r, "me").(*data.Account)

	err := json.NewEncoder(w).Encode(me)
	catch(err)
}

func ServeAccountByHandle(w http.ResponseWriter, r *http.Request) {
	me, _ := context.Get(r, "me").(*data.Account)
	if me == nil || (me.Level != data.Judge && me.Level != data.Administrator) {
		http.Error(w, "", http.StatusForbidden)
		return
	}

	acc, err := data.GetAccountByHandle(r.FormValue("handle"))
	catch(err)

	if acc == nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	err = json.NewEncoder(w).Encode(acc)
	catch(err)
}

func ServeAccount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	acc, err := data.GetAccount(id)
	catch(err)

	if acc == nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	err = json.NewEncoder(w).Encode(acc)
	catch(err)
}

func UpdateAccount(w http.ResponseWriter, r *http.Request) {
	me, _ := context.Get(r, "me").(*data.Account)
	if me == nil || me.Level != data.Administrator {
		http.Error(w, "", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)

	id, err := strconv.ParseInt(vars["id"], 10, 64)
	catch(err)
	acc, err := data.GetAccount(id)
	catch(err)

	body := struct {
		Handle   string     `json:"handle"`
		Password string     `json:"password"`
		Level    data.Level `json:"level"`
		Name     string     `json:"name"`
	}{}
	err = json.NewDecoder(r.Body).Decode(&body)
	catch(err)

	acc.Handle = body.Handle
	if body.Password != "" {
		err = acc.SetPassword(body.Password)
		catch(err)
	}
	acc.Level = body.Level
	acc.Name = body.Name
	err = acc.Put()
	catch(err)

	json.NewEncoder(w).Encode(acc)
	hub.Send([]interface{}{"SYNC", "accounts", acc.Id})

	err = data.NewActivity(me, fmt.Sprintf("updated account %d", acc.Id)).Put()
	catch(err)
	hub.Send([]interface{}{"SYNC", "activities"})
}

func UpdateAccountPart(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	id, err := strconv.ParseInt(vars["id"], 10, 64)
	catch(err)
	acc, err := data.GetAccount(id)
	catch(err)

	me, _ := context.Get(r, "me").(*data.Account)
	if me == nil || me.Id != acc.Id {
		http.Error(w, "", http.StatusForbidden)
		return
	}

	body := struct {
		Notified time.Time `json:"notified"`
	}{}
	err = json.NewDecoder(r.Body).Decode(&body)
	catch(err)

	acc.Notified = body.Notified
	err = acc.Put()
	catch(err)

	json.NewEncoder(w).Encode(acc)
	hub.Send([]interface{}{"SYNC", "accounts", acc.Id})
}

func DeleteAccount(w http.ResponseWriter, r *http.Request) {
	me, _ := context.Get(r, "me").(*data.Account)
	if me == nil || me.Level != data.Administrator {
		http.Error(w, "", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)

	id, err := strconv.ParseInt(vars["id"], 10, 64)
	catch(err)
	acc, err := data.GetAccount(id)
	catch(err)

	if acc.Id == me.Id || acc.Id == 1 {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	err = acc.Del()
	catch(err)

	json.NewEncoder(w).Encode(&struct {
		Id int64 `json:"id"`
	}{
		Id: acc.Id,
	})
	hub.Send([]interface{}{"SYNC", "accounts"})

	err = data.NewActivity(me, fmt.Sprintf("deleted account %d", acc.Id)).Put()
	catch(err)
	hub.Send([]interface{}{"SYNC", "activities"})
}

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	acc, err := data.GetAccountByHandle(r.FormValue("handle"))
	catch(err)
	if acc == nil {
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	ok, err := acc.CmpPassword(r.FormValue("password"))
	catch(err)
	if !ok {
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	sess, err := Store.Get(r, "s")
	catch(err)

	sess.Values["me.id"] = acc.Id
	err = sess.Save(r, w)
	catch(err)

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	catch(err)
	err = data.NewActivity(acc, fmt.Sprintf("logged in from %s", host)).Put()
	catch(err)
	hub.Send([]interface{}{"SYNC", "activities"})
}

func HandleLogout(w http.ResponseWriter, r *http.Request) {
	sess, err := Store.Get(r, "s")
	catch(err)

	delete(sess.Values, "me.id")
	sess.Options.MaxAge = -1
	err = sess.Save(r, w)
	catch(err)
}
