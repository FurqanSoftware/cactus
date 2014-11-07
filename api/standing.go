// Copyright 2014 The Cactus Authors. All rights reserved.

package api

import (
	"encoding/json"
	"net/http"

	"github.com/hjr265/cactus/data"
)

func ServeStandingList(w http.ResponseWriter, r *http.Request) {
	stnds, err := data.ListStandings()
	catch(err)
	err = json.NewEncoder(w).Encode(stnds)
	catch(err)
}
