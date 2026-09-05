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

	"github.com/gorilla/mux"
	"github.com/mattn/go-sqlite3"

	"github.com/FurqanSoftware/cactus/data"
	"github.com/FurqanSoftware/cactus/hub"
)

func ServeAccountList(w http.ResponseWriter, r *http.Request) {
	me := currentAccount(r)

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
	me := currentAccount(r)
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
	if err, ok := err.(sqlite3.Error); ok && err.ExtendedCode == sqlite3.ErrConstraintUnique {
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

type accountImportSkip struct {
	Line   int    `json:"line"`
	Reason string `json:"reason"`
}

func ImportAccounts(w http.ResponseWriter, r *http.Request) {
	me := currentAccount(r)
	if me == nil || me.Level != data.Administrator {
		http.Error(w, "", http.StatusForbidden)
		return
	}

	body := csv.NewReader(r.Body)
	body.FieldsPerRecord = -1

	imported := 0
	skipped := []accountImportSkip{}
	for {
		cols, err := body.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			http.Error(w, fmt.Sprintf("%s (%d accounts imported before this line)", err, imported), http.StatusBadRequest)
			return
		}

		line := 0
		if len(cols) > 0 {
			line, _ = body.FieldPos(0)
		}
		skip := func(reason string) {
			skipped = append(skipped, accountImportSkip{Line: line, Reason: reason})
		}

		if len(cols) < 4 {
			skip("expected the fields handle, password, level and name")
			continue
		}

		handle, password, name := cols[0], cols[1], cols[3]
		col2, err := strconv.ParseInt(cols[2], 10, 32)
		level := data.Level(col2)

		switch {
		case len(handle) < 4:
			skip("handle is shorter than 4 characters")
			continue

		case len(password) < 4:
			skip("password is shorter than 4 characters")
			continue

		case err != nil:
			skip("level is not a number")
			continue

		case level != data.Participant && level != data.Judge && level != data.Administrator:
			skip("level is not one of 1, 2 or 3")
			continue

		case name == "":
			name = handle
		}

		acc := &data.Account{}
		acc.Handle = handle
		err = acc.SetPassword(password)
		catch(err)
		acc.Level = level
		acc.Name = name

		err = acc.Put()
		if err, ok := err.(sqlite3.Error); ok && err.ExtendedCode == sqlite3.ErrConstraintUnique {
			skip("handle is already taken")
			continue
		}
		catch(err)
		imported++

		err = data.NewActivity(me, fmt.Sprintf("created account %d", acc.Id)).Put()
		catch(err)
		hub.Send([]interface{}{"SYNC", "activities"})
	}
	hub.Send([]interface{}{"SYNC", "accounts"})

	err := json.NewEncoder(w).Encode(struct {
		Imported int                 `json:"imported"`
		Skipped  []accountImportSkip `json:"skipped"`
	}{
		Imported: imported,
		Skipped:  skipped,
	})
	catch(err)
}

func ServeAccountMe(w http.ResponseWriter, r *http.Request) {
	me := currentAccount(r)

	err := json.NewEncoder(w).Encode(me)
	catch(err)
}

func ServeAccountByHandle(w http.ResponseWriter, r *http.Request) {
	me := currentAccount(r)
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
	me := currentAccount(r)
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

	me := currentAccount(r)
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
	me := currentAccount(r)
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
