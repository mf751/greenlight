package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

type envelope map[string]interface{}

func (app *application) readIDParam(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}
	return id, nil
}

func (app *application) writeJSON(
	w http.ResponseWriter,
	status int,
	data envelope,
	headers http.Header,
) error {
	json, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	json = append(json, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(json)

	return nil
}
