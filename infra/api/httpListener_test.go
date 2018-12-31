package api

import (
	"io/ioutil"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/mux"
)

//test setup
var httpClient *http.Client

type testRoutes struct {
}

func (routes testRoutes) SetupRoutes(router *mux.Router) {
	//add future test config here
	log.Println("Setting up http listener")
}

func init() {
	//start listener
	log.Println("Init called on listener test")
	go StartHttpListener(testRoutes{})
	// TODO: listen to startup vs. assuming like below
	time.Sleep(1 * time.Second)
	//set local http client for validation
	httpClient = &http.Client{}
}

//validate health check middleware (validates default host/port as side effect)
//validate logging middleware

//TestHealthCheck - Validates healthcheck was setup as expected
func TestHealthCheck(t *testing.T) {
	//Setting as real integration style client vs. unit test style in mux example

	log.Println("about to call health endpoint")
	resp, err := httpClient.Get("http://127.0.0.1:8080/health")
	log.Println("endpoipnt called...")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	stringbody := string(body)

	// Check the status code is what we expect.
	if status := resp.StatusCode; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body is what we expect.
	expected := `{"alive": true}`
	if stringbody != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			stringbody, expected)
	}
}
