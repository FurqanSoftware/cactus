// Copyright 2014 The Cactus Authors. All rights reserved.

package api

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"

	"github.com/hjr265/cactus/data"
)

func CreateUpload(w http.ResponseWriter, r *http.Request) {
	me, _ := context.Get(r, "me").(*data.Account)
	if me == nil || me.Level != data.Administrator {
		http.Error(w, "", http.StatusForbidden)
		return
	}

	f, h, err := r.FormFile("file")
	defer f.Close()
	key := fmt.Sprintf("%s:%d-%d%s", r.FormValue("kind"), time.Now().UnixNano(), rand.Int63(), path.Ext(h.Filename))
	_, err = data.Blobs.Put(key, f)
	catch(err)

	err = json.NewEncoder(w).Encode(strings.Replace(key, ":", "-", -1))
	catch(err)
}

func ServeUpload(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	key := fmt.Sprintf("%s:%s", vars["kind"], vars["id"])
	blob, err := data.Blobs.Get(key)
	catch(err)
	_, err = io.Copy(w, blob)
	catch(err)
	err = blob.Close()
	catch(err)
}
