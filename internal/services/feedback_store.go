package services

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// Reply is a response to a Feedback.
type Reply struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	Upvotes   int       `json:"upvotes"`
}

// Feedback is a user-submitted suggestion. Feedback and replies are fully
// anonymous — no identity is stored.
type Feedback struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	Upvotes   int       `json:"upvotes"`
	Replies   []Reply   `json:"replies"`
}

// FeedbackStore persists feedback to a JSON file. All mutations are guarded
// by a mutex and flushed to disk immediately.
type FeedbackStore struct {
	mu    sync.Mutex
	items []Feedback
	path  string
}

// NewFeedbackStore loads existing feedback from path (if any) and returns a
// store ready for use.
func NewFeedbackStore(path string) *FeedbackStore {
	s := &FeedbackStore{path: path}
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &s.items)
	}
	return s
}

// List returns feedback sorted by upvotes desc, then created_at asc.
func (s *FeedbackStore) List() []Feedback {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Feedback, len(s.items))
	copy(out, s.items)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Upvotes != out[j].Upvotes {
			return out[i].Upvotes > out[j].Upvotes
		}
		return out[i].CreatedAt.Before(out[j].CreatedAt)
	})
	return out
}

// Add appends a new feedback and persists it.
func (s *FeedbackStore) Add(content string) Feedback {
	s.mu.Lock()
	defer s.mu.Unlock()
	fb := Feedback{
		ID:        newID(),
		Content:   content,
		CreatedAt: time.Now(),
		Replies:   []Reply{},
	}
	s.items = append(s.items, fb)
	s.save()
	return fb
}

// Upvote increments a feedback's upvote count. Returns false if not found.
func (s *FeedbackStore) Upvote(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.items {
		if s.items[i].ID == id {
			s.items[i].Upvotes++
			s.save()
			return true
		}
	}
	return false
}

// AddReply appends a reply to a feedback. Returns false if the feedback is
// not found.
func (s *FeedbackStore) AddReply(feedbackID, content string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.items {
		if s.items[i].ID == feedbackID {
			s.items[i].Replies = append(s.items[i].Replies, Reply{
				ID:        newID(),
				Content:   content,
				CreatedAt: time.Now(),
			})
			s.save()
			return true
		}
	}
	return false
}

// UpvoteReply increments a reply's upvote count. Returns false if not found.
func (s *FeedbackStore) UpvoteReply(feedbackID, replyID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.items {
		if s.items[i].ID != feedbackID {
			continue
		}
		for j := range s.items[i].Replies {
			if s.items[i].Replies[j].ID == replyID {
				s.items[i].Replies[j].Upvotes++
				s.save()
				return true
			}
		}
	}
	return false
}

// save writes the store to disk atomically (temp file + rename). Callers
// must hold the mutex.
func (s *FeedbackStore) save() {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return
	}
	data, err := json.MarshalIndent(s.items, "", "  ")
	if err != nil {
		return
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return
	}
	_ = os.Rename(tmp, s.path)
}

func newID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
