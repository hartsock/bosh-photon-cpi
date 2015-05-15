package esxcloud

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"time"
)

type testServer struct {
	HttpServer *httptest.Server
	StatusCode *int
	Body       *string
}

func (s *testServer) Close() {
	if s.HttpServer != nil {
		s.HttpServer.Close()
	}
}

func (s *testServer) SetResponse(status int, body string) {
	s.StatusCode = &status
	s.Body = &body
}

func (s *testServer) SetResponseJson(status int, v interface{}) {
	s.SetResponse(status, toJson(v))
}

func newTestServer() (server *testServer) {
	status := 200
	body := ""
	server = &testServer{nil, &status, &body}
	server.HttpServer = httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(*server.StatusCode)
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintln(w, *server.Body)
		}))
	return
}

func testSetup() (server *testServer, client *Client) {
	// If TEST_ENDPOINT env var is set, return an empty server and point
	// the client to TEST_ENDPOINT. This lets us run tests as integration tests
	var uri string
	if os.Getenv("TEST_ENDPOINT") != "" {
		server = &testServer{}
		uri = os.Getenv("TEST_ENDPOINT")
	} else {
		server = newTestServer()
		uri = server.HttpServer.URL
	}
	transport := &http.Transport{Proxy: func(req *http.Request) (*url.URL, error) {
		return url.Parse(uri)
	}}

	httpClient := &http.Client{Transport: transport}
	client = NewTestClient(uri, httpClient)
	return
}

func createMockStep(operation, state string) Step {
	return Step{State: state, Operation: operation}
}

func createMockTask(operation, state string, steps ...Step) *Task {
	return &Task{Operation: operation, State: state, Steps: steps}
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandomString(n int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func isRealAgent() bool {
	return os.Getenv("REAL_AGENT") != ""
}

func isIntegrationTest() bool {
	return os.Getenv("TEST_ENDPOINT") != ""
}