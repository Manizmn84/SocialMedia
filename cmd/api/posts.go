package main

import (
	"net/http"

	"github.com/sikozonpc/social/internal/store"
)

type postPayload struct {
	Title   string   `json:"title"`
	Content string   `json:"content"`
	Tags    []string `json:"tags"`
}

func (app *application) createPostHandler(w http.ResponseWriter, r *http.Request) {
	var postPayload postPayload

	if err := readJson(w, r, &postPayload); err != nil {
		writeJsonError(w, http.StatusBadRequest, err.Error())
		return
	}

	userID := 1

	post := &store.Post{
		Title:   postPayload.Title,
		Content: postPayload.Content,
		// TODO: after auth we should fix that
		UserID: int64(userID),
		Tags: postPayload.Tags,
	}

	if err := app.store.Posts.Create(r.Context(), post); err != nil {
		writeJsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := writeJson(w, http.StatusCreated, post); err != nil {
		writeJsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
}
