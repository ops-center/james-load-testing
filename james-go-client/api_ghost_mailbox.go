/*
Apache JAMES Web Admin API

No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)

API version: 3.8.0
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package openapi

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
)


// GhostMailboxAPIService GhostMailboxAPI service
type GhostMailboxAPIService service

type ApiCorrectGhostMailboxRequest struct {
	ctx context.Context
	ApiService *GhostMailboxAPIService
	correctGhostMailboxRequest *CorrectGhostMailboxRequest
}

// Parameters for merging mailboxes
func (r ApiCorrectGhostMailboxRequest) CorrectGhostMailboxRequest(correctGhostMailboxRequest CorrectGhostMailboxRequest) ApiCorrectGhostMailboxRequest {
	r.correctGhostMailboxRequest = &correctGhostMailboxRequest
	return r
}

func (r ApiCorrectGhostMailboxRequest) Execute() (*CorrectGhostMailbox201Response, *http.Response, error) {
	return r.ApiService.CorrectGhostMailboxExecute(r)
}

/*
CorrectGhostMailbox Correct ghost mailbox by merging

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @return ApiCorrectGhostMailboxRequest
*/
func (a *GhostMailboxAPIService) CorrectGhostMailbox(ctx context.Context) ApiCorrectGhostMailboxRequest {
	return ApiCorrectGhostMailboxRequest{
		ApiService: a,
		ctx: ctx,
	}
}

// Execute executes the request
//  @return CorrectGhostMailbox201Response
func (a *GhostMailboxAPIService) CorrectGhostMailboxExecute(r ApiCorrectGhostMailboxRequest) (*CorrectGhostMailbox201Response, *http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodPost
		localVarPostBody     interface{}
		formFiles            []formFile
		localVarReturnValue  *CorrectGhostMailbox201Response
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "GhostMailboxAPIService.CorrectGhostMailbox")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/cassandra/mailbox/merging"

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}
	if r.correctGhostMailboxRequest == nil {
		return localVarReturnValue, nil, reportError("correctGhostMailboxRequest is required and must be specified")
	}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{"application/json"}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	// body params
	localVarPostBody = r.correctGhostMailboxRequest
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := io.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = io.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}
