package chefapi_lib

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/go-chef/chef"
)

const (
        userid     = "tester"
        privateKey = "-----BEGIN RSA PRIVATE KEY-----"
)

var (
        chefmux *http.ServeMux
        server  *httptest.Server
        client  *chef.Client
)

func TestAllOrgs(t *testing.T) {
        setup()
        defer teardown()

        chefmux.HandleFunc("/organizations", func(w http.ResponseWriter, r *http.Request) {
                switch {
                case r.Method == "GET":
                        fmt.Fprintf(w, `{ "org_name1": "https://url/for/org_name1", "org_name2": "https://url/for/org_name2"}`)
                }
        })
        want := []string{"org_name1", "org_name2"}
        orgs, err := AllOrgs()
        if err != nil {
                t.Errorf("AllOrgs unexpected error %v\n", err)
        }
        if !reflect.DeepEqual(orgs, want) {
                t.Errorf("Organizations.List returned %+v, want %+v", orgs, want)
        }
}

// StdMessageStatus extracts the status code from a go-chef api error message
func TestChefStatus(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer ts.Close()

	resp, _ := http.Get(ts.URL)
	cerr := chef.CheckResponse(resp)
	msg, code := ChefStatus(cerr)
	fmt.Printf("MSG %+v, CODE %+v", msg, code)
	if code != 400 {
		t.Errorf("Expected 400")
	}
}

func TestCleanInput(t *testing.T) {
        expected_in := map[string]string{"valid": "mynode"}
        err := CleanInput(expected_in)
        if err != nil {
                t.Errorf("Error cleaning: %+v Err: %+v\n", expected_in, err)
        }
        expected_in = map[string]string{"invalid": "\nbounceit"}
        err = CleanInput(expected_in)
        if err == nil {
                t.Errorf("CleanInput did not receive expected error cleaning: %+v Err: %+v\n", expected_in, err)
        }
}

// inputerror(w *http.ResponseWriter) {
func TestInputerror(t *testing.T) {
        // Check the status code and response body - invalid request invoked inputerror
        req, err := http.NewRequest("GET", "/organizations/other&org/nodes", nil)
        if err != nil {
                t.Fatal(err)
        }
        rr := httptest.NewRecorder()
        // Invoke the server
        newGetNodesServer().ServeHTTP(rr, req)
        // Check the status code and response body
        if status := rr.Code; status != http.StatusBadRequest {
                t.Errorf("Get Nodes status code is not expected. Got: %v want: %v\n", status, http.StatusBadRequest)
        }
        wantBody := `{"message":"Bad url input value"}`
        if rr.Body.String() != wantBody {
                t.Errorf("Get Nodes unexpected json returned. Expected: %v Got: %v\n", wantBody, rr.Body.String())
        }
}
func setup() {
        chefmux = http.NewServeMux()
        server = httptest.NewServer(chefmux)
        client, _ = chef.NewClient(&chef.Config{
                Name:    userid,
                Key:     privateKey,
                BaseURL: server.URL,
        })
}

func teardown() {
        server.Close()
}

func newGetNodesServer() http.Handler {
        r := mux.NewRouter()
        r.HandleFunc("/organizations/{org}/nodes", getNodes)
        return r
}

func newSingleNodeServer() http.Handler {
        r := mux.NewRouter()
        r.HandleFunc("/organizations/{org}/nodes/{node}", singleNode)
        return r
}
