package main

import (
	"log"
	"net/http"
)

func (app *application) internalServerError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("internal server error: %s path: %s error: %s", r.Method, r.URL.Path, err)

	writeJsonError(w, http.StatusInternalServerError, err.Error())
}

func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("bad request error: %s path: %s error: %s", r.Method, r.URL.Path, err)

	writeJsonError(w, http.StatusBadRequest, err.Error())
}

func (app *application) NotFoundResponse(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("Not Found error: %s path: %s error: %s", r.Method, r.URL.Path, err)

	writeJsonError(w, http.StatusNotFound, err.Error())
}
