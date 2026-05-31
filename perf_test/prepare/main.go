// prepare provisions a pool of test users against a running Noterian API and
// emits a vegeta NDJSON targets file with 100,000 POST /notes requests
// distributed round-robin across those users.
//
// Each target carries the user's JWT cookie, CSRF cookie and X-CSRF-Token
// header so that vegeta can fire them without any state of its own.
//
// Usage:
//
//	go run ./prepare -base=http://localhost:8000 -users=100 -notes=100000 \
//	    -out=write_targets.json -users-out=users.json
package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
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

type csrfBody struct {
	CSRFToken string `json:"csrf_token"`
}

func main() {
	var (
		baseURL    = flag.String("base", "http://localhost:8000/api", "API base URL")
		userCount  = flag.Int("users", 100, "number of users to create")
		noteCount  = flag.Int("notes", 100_000, "number of POST /notes targets to emit")
		csrfHeader = flag.String("csrf-header", "X-CSRF-Token", "CSRF header name (must match server CSRF_HEADER_NAME)")
		outTargets = flag.String("out", "write_targets.json", "path to vegeta NDJSON targets file for writes")
		outUsers   = flag.String("users-out", "users.json", "path to JSON file with provisioned user credentials")
	)
	flag.Parse()

	if *userCount <= 0 || *noteCount <= 0 {
		log.Fatalf("users and notes must be positive (got %d and %d)", *userCount, *noteCount)
	}

	client := &http.Client{
		Timeout: 15 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	users := make([]userCreds, 0, *userCount)
	for i := 0; i < *userCount; i++ {
		u, err := provisionUser(client, *baseURL)
		if err != nil {
			log.Fatalf("provision user %d: %v", i, err)
		}
		users = append(users, *u)
		if (i+1)%25 == 0 || i+1 == *userCount {
			log.Printf("provisioned %d/%d users", i+1, *userCount)
		}
	}

	if err := writeUsersFile(*outUsers, users); err != nil {
		log.Fatalf("write users file: %v", err)
	}

	if err := writeWriteTargets(*outTargets, *baseURL, *csrfHeader, users, *noteCount); err != nil {
		log.Fatalf("write targets file: %v", err)
	}

	log.Printf("wrote %d users to %s and %d POST /notes targets to %s",
		len(users), *outUsers, *noteCount, *outTargets)
}

func provisionUser(client *http.Client, baseURL string) (*userCreds, error) {
	username, err := randomUsername()
	if err != nil {
		return nil, err
	}
	const password = "Load1Test"

	signupReq := map[string]string{"username": username, "password": password}
	signupBody, _ := json.Marshal(signupReq)

	req, err := http.NewRequest(http.MethodPost, baseURL+"/signup", bytes.NewReader(signupBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("signup http: %w", err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("signup status=%d body=%s", resp.StatusCode, truncate(string(body), 200))
	}

	jwtCookie, err := pickAuthCookie(resp.Cookies())
	if err != nil {
		return nil, fmt.Errorf("signup: %w", err)
	}

	csrfReq, err := http.NewRequest(http.MethodGet, baseURL+"/csrf-token", nil)
	if err != nil {
		return nil, err
	}
	csrfReq.AddCookie(jwtCookie)

	csrfResp, err := client.Do(csrfReq)
	if err != nil {
		return nil, fmt.Errorf("csrf http: %w", err)
	}
	csrfBytes, _ := io.ReadAll(csrfResp.Body)
	csrfResp.Body.Close()
	if csrfResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("csrf status=%d body=%s", csrfResp.StatusCode, truncate(string(csrfBytes), 200))
	}

	var cb csrfBody
	if err := json.Unmarshal(csrfBytes, &cb); err != nil || cb.CSRFToken == "" {
		return nil, fmt.Errorf("parse csrf body: %v body=%s", err, truncate(string(csrfBytes), 200))
	}
	csrfCookie, err := pickAnyCookie(csrfResp.Cookies())
	if err != nil {
		return nil, fmt.Errorf("csrf cookie: %w", err)
	}

	cookieHeader := fmt.Sprintf("%s=%s; %s=%s",
		jwtCookie.Name, jwtCookie.Value,
		csrfCookie.Name, csrfCookie.Value,
	)
	return &userCreds{
		Username:     username,
		Password:     password,
		CookieHeader: cookieHeader,
		CSRFToken:    cb.CSRFToken,
	}, nil
}

// pickAuthCookie returns the cookie set by /signup. The server may use any
// cookie name (configurable via JWT_COOKIE_NAME), so we just take the only one
// it returned.
func pickAuthCookie(cookies []*http.Cookie) (*http.Cookie, error) {
	if len(cookies) == 0 {
		return nil, fmt.Errorf("no Set-Cookie on signup response")
	}
	return cookies[0], nil
}

func pickAnyCookie(cookies []*http.Cookie) (*http.Cookie, error) {
	if len(cookies) == 0 {
		return nil, fmt.Errorf("no Set-Cookie on csrf response")
	}
	return cookies[0], nil
}

func randomUsername() (string, error) {
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", err
	}
	return "load" + hex.EncodeToString(buf[:]), nil
}

func writeUsersFile(path string, users []userCreds) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(users)
}

func writeWriteTargets(path, baseURL, csrfHeader string, users []userCreds, total int) error {
	u, err := url.Parse(baseURL + "/notes")
	if err != nil {
		return err
	}
	postURL := u.String()

	body := []byte(`{"title":"loadtest","is_public":false,"is_favorite":false,"icon":""}`)

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	for i := 0; i < total; i++ {
		u := &users[i%len(users)]
		t := vegetaTarget{
			Method: http.MethodPost,
			URL:    postURL,
			Header: map[string][]string{
				"Cookie":       {u.CookieHeader},
				csrfHeader:     {u.CSRFToken},
				"Content-Type": {"application/json"},
			},
			Body: body,
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
