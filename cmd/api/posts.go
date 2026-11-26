package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/sikozonpc/social/internal/store"
)

type postKey string

const postCtx postKey = "post"

type postPayload struct {
	Title   string   `json:"title" validate:"required,max=100"`
	Content string   `json:"content" validate:"required,max=1000"`
	Tags    []string `json:"tags"`
}

func (app *application) createPostHandler(w http.ResponseWriter, r *http.Request) {
	var postPayload postPayload

	if err := readJson(w, r, &postPayload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(postPayload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	userID := 1

	post := &store.Post{
		Title:   postPayload.Title,
		Content: postPayload.Content,
		// TODO: after auth we should fix that
		UserID: int64(userID),
		Tags:   postPayload.Tags,
	}

	if err := app.store.Posts.Create(r.Context(), post); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := jsonResponse(w, http.StatusCreated, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) getPostByID(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r)
	comments, err := app.store.Comments.GetByPostID(r.Context(), uint(post.ID))
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	post.Comments = comments

	if err := jsonResponse(w, http.StatusOK, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}

}

func (app *application) postsContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("postID"), 10, strconv.IntSize)
		if err != nil {
			app.badRequestResponse(w, r, err)
			return
		}
		if id <= 0 {
			app.badRequestResponse(w, r, errors.New("postID must be a positive integer"))
			return
		}

		ctx := r.Context()

		post, err := app.store.Posts.GetByID(ctx, uint(id))
		if err != nil {
			switch {
			case errors.Is(err, store.ErrNotFound):
				app.NotFoundResponse(w, r, err)
			default:
				app.internalServerError(w, r, err)
			}
			return
		}

		ctx = context.WithValue(ctx, postCtx, post)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getPostFromCtx(r *http.Request) store.Post {
	return r.Context().Value(postCtx).(store.Post)
}

type UpdatePostPayload struct {
	Version *int    `json:"version" validate:"omitempty,min=1"`
	Title   *string `json:"title" validate:"omitempty,max=100"`
	Content *string `json:"content" validate:"omitempty,max=1000"`
}

func (app *application) updatePostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r)
	var payload UpdatePostPayload
	if err := readJson(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	if payload.Version != nil && *payload.Version != post.Version {
		writeJsonError(w, http.StatusConflict, store.ErrEditConflict.Error())
		return
	}
	if payload.Title != nil {
		post.Title = *payload.Title
	}
	if payload.Content != nil {
		post.Content = *payload.Content
	}
	if err := app.store.Posts.Update(r.Context(), &post); err != nil {
		if errors.Is(err, store.ErrEditConflict) {
			writeJsonError(w, http.StatusConflict, err.Error())
		} else if errors.Is(err, store.ErrNotFound) {
			app.NotFoundResponse(w, r, err)
		} else {
			app.internalServerError(w, r, err)
		}
		return
	}
	if err := jsonResponse(w, http.StatusOK, post); err != nil {
		app.internalServerError(w, r, err)
	}
}

func (app *application) deletePostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r)
	if err := app.store.Posts.Delete(r.Context(), post.ID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			app.NotFoundResponse(w, r, err)
		} else {
			app.internalServerError(w, r, err)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
