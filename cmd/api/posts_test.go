package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sikozonpc/social/internal/store"
)

type postTestStore struct {
	post                         store.Post
	getErr, updateErr, deleteErr error
	reads, updates, deletes      int
}

func (s *postTestStore) Create(context.Context, *store.Post) error { return nil }
func (s *postTestStore) GetByID(context.Context, uint) (store.Post, error) {
	s.reads++
	return s.post, s.getErr
}
func (s *postTestStore) Update(_ context.Context, p *store.Post) error {
	s.updates++
	if s.updateErr != nil {
		return s.updateErr
	}
	p.Version++
	s.post = *p
	return nil
}
func (s *postTestStore) Delete(context.Context, int64) error { s.deletes++; return s.deleteErr }

func TestPostMutations(t *testing.T) {
	for _, tc := range []struct {
		name, method, id, body string
		status                 int
		getErr, mutationErr    error
	}{
		{name: "partial title", method: "PATCH", id: "1", body: `{"title":"new"}`, status: 200},
		{name: "partial content", method: "PATCH", id: "1", body: `{"content":"new"}`, status: 200},
		{name: "empty title", method: "PATCH", id: "1", body: `{"title":""}`, status: 200},
		{name: "invalid JSON", method: "PATCH", id: "1", body: `{`, status: 400},
		{name: "title too long", method: "PATCH", id: "1", body: `{"title":"` + strings.Repeat("a", 101) + `"}`, status: 400},
		{name: "unknown field", method: "PATCH", id: "1", body: `{"user_id":2}`, status: 400},
		{name: "missing post", method: "PATCH", id: "1", body: `{}`, status: 404, getErr: store.ErrNotFound},
		{name: "update failure", method: "PATCH", id: "1", body: `{}`, status: 500, mutationErr: errors.New("db failure")},
		{name: "deleted before update", method: "PATCH", id: "1", body: `{}`, status: 404, mutationErr: store.ErrNotFound},
		{name: "version conflict", method: "PATCH", id: "1", body: `{}`, status: 409, mutationErr: store.ErrEditConflict},
		{name: "stale client version", method: "PATCH", id: "1", body: `{"version":2,"title":"new"}`, status: 409},
		{name: "valid client version", method: "PATCH", id: "1", body: `{"version":1,"title":"new"}`, status: 200},
		{name: "invalid client version", method: "PATCH", id: "1", body: `{"version":0}`, status: 400},
		{name: "delete", method: "DELETE", id: "1", status: 204},
		{name: "delete failure", method: "DELETE", id: "1", status: 500, mutationErr: errors.New("db failure")},
		{name: "deleted before delete", method: "DELETE", id: "1", status: 404, mutationErr: store.ErrNotFound},
		{name: "negative ID", method: "DELETE", id: "-1", status: 400},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := &postTestStore{post: store.Post{ID: 1, Title: "old title", Content: "old content", Version: 1}, getErr: tc.getErr, updateErr: tc.mutationErr, deleteErr: tc.mutationErr}
			app := &application{store: store.Storage{Posts: s}}
			w := httptest.NewRecorder()
			app.mount().ServeHTTP(w, httptest.NewRequest(tc.method, "/v1/posts/"+tc.id, strings.NewReader(tc.body)))
			if w.Code != tc.status {
				t.Fatalf("want %d got %d: %s", tc.status, w.Code, w.Body.String())
			}
			if tc.status == 400 || tc.getErr != nil {
				if s.updates+s.deletes != 0 {
					t.Fatal("invalid request mutated store")
				}
			}
			if tc.status == 204 && w.Body.Len() != 0 {
				t.Fatal("delete response must be empty")
			}
			if tc.status == 200 {
				var response struct {
					Data store.Post `json:"data"`
				}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatal(err)
				}
				got := response.Data
				if got.Version != 2 {
					t.Fatalf("expected incremented version, got %d", got.Version)
				}
				if s.reads != 1 {
					t.Fatalf("expected one middleware lookup, got %d", s.reads)
				}
				if tc.name == "partial title" && (got.Title != "new" || got.Content != "old content") {
					t.Fatal("partial update changed omitted field")
				}
				if tc.name == "partial content" && (got.Content != "new" || got.Title != "old title") {
					t.Fatal("partial update changed omitted field")
				}
				if tc.name == "empty title" && got.Title != "" {
					t.Fatal("explicit empty string ignored")
				}
			}
		})
	}
}
