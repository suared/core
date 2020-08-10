package comms

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/suared/core/security"
)

//Utilizing the test/e2e to start as a copy to allow this to grow as needed for service to service comms use cases
//only moving Get for now with auth, will add others as  the need arises if before eventing issetup
//Starting wtih simple user propogation for current need

//test client setup
var httpClient *http.Client

func init() {
	//set local http client for validation
	httpClient = &http.Client{Timeout: time.Second * 5}
}

//SimpleGet -Returns the body of an http get request with error if appropriate
func SimpleGet(ctx context.Context, uri string) (string, error) {
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		log.Println(fmt.Sprintf("Comms:SimpleGet:Failed reading request to uri: %v, err:%v", uri, err))
		return "", err
	}

	req.Header.Set("Authorization", security.GetAuth(ctx).GetAuthHeader())

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Println(fmt.Sprintf("Comms:SimpleGet:Failed reading response to: %v, err:%v", uri, err))
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	stringbody := string(body)

	// Check the status code is what we expect.
	if status := resp.StatusCode; status != http.StatusOK {
		err = errors.New(string(status) + ": " + stringbody + ": accessing: " + uri)
	}

	return stringbody, err
}
