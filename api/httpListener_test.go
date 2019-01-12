package api

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	_ "github.com/suared/core/infra"

	"github.com/gorilla/mux"
)

//Using this as more of the integration test across the core arch setup
//Other testers in this package will only handle the non-integration needs

//test setup
var httpClient *http.Client

//sample struct that implements Config for route setup
type testRoutes struct {
}

func (routes testRoutes) SetupRoutes(router *mux.Router) {
	//add future test config here
	log.Println("Setting up http listener")
}

func init() {
	//start listener
	log.Println("Init called on listener test")
	go StartHTTPListener(testRoutes{})
	// TODO: listen to startup vs. assuming like below, this will work for now
	time.Sleep(1 * time.Second)
	//set local http client for validation
	httpClient = &http.Client{}
}

//TODO: Setup and validate auth + cors + csrf + metrics middleware

//TestHealthCheck - Validates healthcheck was setup as expected (validates default host/port as side effect)
func TestHealthCheck(t *testing.T) {
	//Setting as real integration style client vs. unit test style in mux example
	resp, err := httpClient.Get(os.Getenv("PROCESS_LISTEN_URI") + "/health")
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
