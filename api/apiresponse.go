package api

import (
	"context"
	"encoding/json"
	"net/http"

	coreerrors "github.com/suared/core/errors"
)

/* Successful Response body will always be as follows by request type:
	GET - Full Object Requested
	POST - Location Header only, no body data.  e.g. Location:/books/12345  (assumes server will not add anything not already available from resource creation)
	PUT - No data  (assume success = client held values match server, same for below)
	PATCH - No data
	DELETE - No data

    Unexpected errors will use  -
	Client Did something wrong - 400-404, with APIError struct in body
	API Did something wrong - 500-501, with APIError struct

	For simplicity as a library, will support all error interface types, core.errors types will have more information associated with them
*/

//WriteGetAPIResponse - writes GET responses
func WriteGetAPIResponse(ctx context.Context, w http.ResponseWriter, r *http.Request, jsonResponse interface{}, err error) {
	if err != nil {
		coreerror := getCoreError(ctx, err)
		//write error message
		w.WriteHeader(coreerror.ErrorType)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(coreerror)
	} else {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(jsonResponse)
	}
}

//WritePostAPIResponse - writes POST responses
func WritePostAPIResponse(ctx context.Context, w http.ResponseWriter, r *http.Request, resourceLocation string, err error) {
	if err != nil {
		coreerror := getCoreError(ctx, err)
		//write error message
		w.WriteHeader(coreerror.ErrorType)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(coreerror)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Location", resourceLocation)
	}
}

//WritePutAPIResponse - writes PUT responses
func WritePutAPIResponse(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	if err != nil {
		coreerror := getCoreError(ctx, err)
		//write error message
		w.WriteHeader(coreerror.ErrorType)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(coreerror)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

//WritePatchAPIResponse - writes Patch responses
func WritePatchAPIResponse(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	if err != nil {
		coreerror := getCoreError(ctx, err)
		//write error message
		w.WriteHeader(coreerror.ErrorType)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(coreerror)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

//WriteDeleteAPIResponse - writes Delete responses
func WriteDeleteAPIResponse(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	if err != nil {
		coreerror := getCoreError(ctx, err)
		//write error message
		w.WriteHeader(coreerror.ErrorType)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(coreerror)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func getCoreError(ctx context.Context, err error) coreerrors.Error {
	coreType, ok := err.(coreerrors.Error)
	if ok {
		return coreType
	} else {
		coreErr := coreerrors.NewError(ctx, err)
		return coreErr
	}
}
