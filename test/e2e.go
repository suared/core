package test

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

//test client setup
var httpClient *http.Client

func init() {
	//set local http client for validation
	httpClient = &http.Client{}
}

//SimpleGet -Returns the body of an http get request with error if appropriate
func SimpleGet(uri string) (string, error) {
	resp, err := httpClient.Get(uri)
	if err != nil {
		return "", errors.New("Unable to get URI:" + uri)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	stringbody := string(body)

	// Check the status code is what we expect.
	if status := resp.StatusCode; status != http.StatusOK {
		err = errors.New(fmt.Sprint(status) + ": " + stringbody + ": accessing: " + uri)
	}

	return stringbody, err
}

//SimplePost - Returns the resource URI of a newly created resource per the location header
func SimplePost(uri string, jsonBytes []byte) (string, error) {
	resp, err := httpClient.Post(uri, "application/json", bytes.NewReader(jsonBytes))
	if err != nil {
		return "", errors.New("Unable to post URI:" + uri + ", error:" + err.Error())
	}
	defer resp.Body.Close()

	//log.Printf("SimplePost Headers returned header: %v", resp.Header)
	locationArray := resp.Header["Location"]

	// if an error occurs, the body will contain the error
	body, err := ioutil.ReadAll(resp.Body)
	stringbody := string(body)
	//log.Printf("stringbody contains: %v", stringbody)

	// Check the status code is what we expect.
	if status := resp.StatusCode; status != http.StatusOK {
		err = errors.New(fmt.Sprint(status) + ":" + stringbody)
		return "", err
	}

	if len(locationArray) == 0 {
		//log.Printf("Fell into location header miss error, locArray: %v, len: %v", locationArray, len(locationArray))
		err = errors.New(stringbody + ": No or Unexpected Location Header.  Location header length: " + fmt.Sprint(len(locationArray)))
		return "", err
	}

	//Only one location is expected in simplePost
	return locationArray[0], err
}

//SimplePatch - Tests an endpoint and returns an error if unsuccessful
func SimplePatch(uri string, jsonBytes []byte) error {
	req, err := http.NewRequest(http.MethodPatch, uri, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return errors.New("Unable to construct request:" + uri + ", error:" + err.Error())
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)

	if err != nil {
		return errors.New("Unable to get URI:" + uri + ", error:" + err.Error())
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	stringbody := string(body)

	// Check the status code is what we expect.
	if status := resp.StatusCode; status != http.StatusOK {
		err = errors.New(stringbody)
	}

	return err
}

//SimpleDelete - Tests an endpoint and returns an error if unsuccessful
func SimpleDelete(uri string) error {
	req, err := http.NewRequest(http.MethodDelete, uri, nil)
	if err != nil {
		return errors.New("Unable to construct request:" + uri + ", error:" + err.Error())
	}

	resp, err := httpClient.Do(req)

	if err != nil {
		return errors.New("Unable to get URI:" + uri + ", error:" + err.Error())
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	stringbody := string(body)

	// Check the status code is what we expect.
	if status := resp.StatusCode; status != http.StatusOK {
		err = errors.New(stringbody)
	}

	return err
}

//SimplePut - Tests an endpoint and returns an error if unsuccessful
func SimplePut(uri string, jsonBytes []byte) error {
	req, err := http.NewRequest(http.MethodPut, uri, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return errors.New("Unable to construct request:" + uri + ", error:" + err.Error())
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)

	if err != nil {
		return errors.New("Unable to get URI:" + uri + ", error:" + err.Error())
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	stringbody := string(body)

	// Check the status code is what we expect.
	if status := resp.StatusCode; status != http.StatusOK {
		err = errors.New(stringbody)
	}

	return err
}
