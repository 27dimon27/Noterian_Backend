// harvest reads the users.json produced by prepare, calls GET /notes for each
// user to collect the IDs of notes created during the write load test, and
// emits two vegeta NDJSON targets files for the read phase:
//
//	read_list_targets.json  - one GET /notes per row (lists all of a user's notes)
//	read_get_targets.json   - one GET /notes/{id} per row (single note + blocks)
//
// Usage:
//
//	go run ./harvest -base=http://localhost:8000 -users=users.json \
//	    -reads=100000 -list-out=read_list_targets.json -get-out=read_get_targets.json
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type vegetaTarget struct {
	Method string              `json:"method"`
	URL    string              `json:"url"`
	Header map[string][]string `json:"header,omitempty"`
	Body   []byte              `json:"body,omitempty"`
}

type userCreds struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	CookieHeader string `json:"cookie_header"`
	CSRFToken    string `json:"csrf_token"`
}

type noteBrief struct {
	ID string `json:"id"`
}

type notesResponse struct {
	Notes []noteBrief `json:"notes"`
	Total int         `json:"total"`
}

type userNotes struct {
	user userCreds
	ids  []string
}

func main() {
	var (
		baseURL     = flag.String("base", "http://localhost:8000/api", "API base URL")
		usersPath   = flag.String("users", "users.json", "path to users file produced by prepare")
		readCount   = flag.Int("reads", 100_000, "number of GET /notes/{id} targets to emit")
		listOutPath = flag.String("list-out", "read_list_targets.json", "vegeta targets file for GET /notes (list)")
		getOutPath  = flag.String("get-out", "read_get_targets.json", "vegeta targets file for GET /notes/{id}")
	)
	flag.Parse()

	users, err := loadUsers(*usersPath)
	if err != nil {
		log.Fatalf("load users: %v", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}

	all := make([]userNotes, 0, len(users))
	totalIDs := 0
	for i, u := range users {
		ids, err := fetchNoteIDs(client, *baseURL, u)
		if err != nil {
			log.Fatalf("fetch notes for user %d (%s): %v", i, u.Username, err)
		}
		all = append(all, userNotes{user: u, ids: ids})
		totalIDs += len(ids)
		if (i+1)%25 == 0 || i+1 == len(users) {
			log.Printf("harvested %d/%d users (running total %d note IDs)", i+1, len(users), totalIDs)
		}
	}

	if totalIDs == 0 {
		log.Fatalf("no notes found - did the write phase run successfully?")
	}

	if err := writeListTargets(*listOutPath, *baseURL, users); err != nil {
		log.Fatalf("write list targets: %v", err)
	}
	if err := writeGetTargets(*getOutPath, *baseURL, all, *readCount); err != nil {
		log.Fatalf("write get targets: %v", err)
	}

	log.Printf("collected %d note IDs across %d users", totalIDs, len(users))
	log.Printf("wrote %d GET /notes targets to %s", len(users), *listOutPath)
	log.Printf("wrote %d GET /notes/{id} targets to %s", *readCount, *getOutPath)
}

func loadUsers(path string) ([]userCreds, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var users []userCreds
	if err := json.NewDecoder(f).Decode(&users); err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, fmt.Errorf("users file is empty")
	}
	return users, nil
}

func fetchNoteIDs(client *http.Client, baseURL string, u userCreds) ([]string, error) {
	req, err := http.NewRequest(http.MethodGet, baseURL+"/notes", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Cookie", u.CookieHeader)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET /notes status=%d body=%s", resp.StatusCode, truncate(string(body), 200))
	}

	var nr notesResponse
	if err := json.Unmarshal(body, &nr); err != nil {
		return nil, fmt.Errorf("decode notes list: %w body=%s", err, truncate(string(body), 200))
	}
	ids := make([]string, 0, len(nr.Notes))
	for _, n := range nr.Notes {
		if n.ID != "" {
			ids = append(ids, n.ID)
		}
	}
	return ids, nil
}

func writeListTargets(path, baseURL string, users []userCreds) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	listURL := baseURL + "/notes"
	for _, u := range users {
		t := vegetaTarget{
			Method: http.MethodGet,
			URL:    listURL,
			Header: map[string][]string{
				"Cookie": {u.CookieHeader},
			},
		}
		if err := enc.Encode(t); err != nil {
			return err
		}
	}
	return nil
}

// writeGetTargets emits `total` GET /notes/{id} requests by walking each
// user's note IDs in a round-robin fashion. If a user has fewer IDs than the
// per-user share, we wrap around their list so every emitted target is valid.
func writeGetTargets(path, baseURL string, all []userNotes, total int) error {
	usable := make([]userNotes, 0, len(all))
	for _, un := range all {
		if len(un.ids) > 0 {
			usable = append(usable, un)
		}
	}
	if len(usable) == 0 {
		return fmt.Errorf("no users had any notes - rerun the write phase")
	}

	cursors := make([]int, len(usable))
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)

	for i := 0; i < total; i++ {
		ui := i % len(usable)
		un := usable[ui]
		id := un.ids[cursors[ui]%len(un.ids)]
		cursors[ui]++

		t := vegetaTarget{
			Method: http.MethodGet,
			URL:    fmt.Sprintf("%s/notes/%s", baseURL, id),
			Header: map[string][]string{
				"Cookie": {un.user.CookieHeader},
			},
		}
		if err := enc.Encode(t); err != nil {
			return err
		}
	}
	return nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return strings.TrimSpace(s[:n]) + "..."
}
