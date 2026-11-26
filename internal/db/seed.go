package db

import (
	"context"
	"fmt"
	"log"
	"math/rand"

	"github.com/sikozonpc/social/internal/store"
	"golang.org/x/crypto/bcrypt"
)

var usernames = []string{"alice", "bob", "charlie", "dave", "eve", "frank", "grace", "heidi"}
var titles = []string{"Learning Go", "Working with PostgreSQL", "Building an API", "Understanding Context", "Testing HTTP Handlers", "Database Transactions"}
var contents = []string{
	"Today I practiced writing Go programs and learned something new.",
	"Here are some observations from building a small backend project.",
	"A practical example helped me understand this topic more clearly.",
	"I tried a different approach and wanted to share the results.",
}
var tags = []string{"go", "postgresql", "backend", "api", "testing", "learning"}
var comments = []string{"Thanks for the example!", "This helped me understand the topic.", "I will try this in my project.", "Could you share more details?", "Nice explanation!"}

// Seed inserts development data after migrations have been applied.
// It is explicit and insert-only: it does not run during normal API startup.
// On error, previously inserted records remain. Re-running with the same usernames
// fails their unique constraint; use a fresh development database for a full seed.
func Seed(ctx context.Context, storage store.Storage) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	// UsersStore currently stores Password directly, so hash this development password here.
	hash, err := bcrypt.GenerateFromPassword([]byte("seed-password"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash seed password: %w", err)
	}
	users := generateUsers(100)
	for _, user := range users {
		user.Password = string(hash)
		if err := storage.Users.Create(ctx, user); err != nil {
			return fmt.Errorf("seed user %q: %w", user.Username, err)
		}
	}
	posts := generatePosts(200, users)
	for _, post := range posts {
		if err := storage.Posts.Create(ctx, post); err != nil {
			return fmt.Errorf("seed post %q: %w", post.Title, err)
		}
	}
	for _, comment := range generateComments(500, users, posts) {
		if err := storage.Comments.Create(ctx, comment); err != nil {
			return fmt.Errorf("seed comment for post %d: %w", comment.PostID, err)
		}
	}
	log.Println("Seeding complete: 100 users, 200 posts, 500 comments")
	return nil
}

func generateUsers(num int) []*store.User {
	users := make([]*store.User, num)
	for i := range users {
		username := fmt.Sprintf("%s%d", usernames[i%len(usernames)], i)
		users[i] = &store.User{Username: username, Email: username + "@example.com"}
	}
	return users
}

func generatePosts(num int, users []*store.User) []*store.Post {
	if len(users) == 0 {
		return nil
	}
	posts := make([]*store.Post, num)
	for i := range posts {
		posts[i] = &store.Post{
			UserID:  users[rand.Intn(len(users))].ID,
			Title:   titles[rand.Intn(len(titles))],
			Content: contents[rand.Intn(len(contents))],
			Tags:    []string{tags[rand.Intn(len(tags))], tags[rand.Intn(len(tags))]},
		}
	}
	return posts
}

func generateComments(num int, users []*store.User, posts []*store.Post) []*store.Comment {
	if len(users) == 0 || len(posts) == 0 {
		return nil
	}
	result := make([]*store.Comment, num)
	for i := range result {
		result[i] = &store.Comment{
			UserID:  users[rand.Intn(len(users))].ID,
			PostID:  posts[rand.Intn(len(posts))].ID,
			Content: comments[rand.Intn(len(comments))],
		}
	}
	return result
}
