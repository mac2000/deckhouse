/*
Copyright 2021 Flant JSC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package checker

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"

	"d8.io/upmeter/pkg/check"
)

// httpChecker implements the checker for HTTP endpoints
type httpChecker struct {
	client   *http.Client
	verifier httpVerifier
	req      *http.Request
}

func newHTTPChecker(client *http.Client, verifier httpVerifier) check.Checker {
	return &httpChecker{
		client:   client,
		verifier: verifier,
	}
}

func (c *httpChecker) Check() check.Error {
	c.req = c.verifier.Request()
	body, err := doRequest(c.client, c.req)
	if err != nil {
		return err
	}
	err = c.verifier.Verify(body)
	return err
}

func (c *httpChecker) BusyWith() string {
	// It might feel that here we can be caught by a race condition.
	// But in normal case request is always created before timeout message is used.
	if c.req == nil {
		return "(request not ready)"
	}
	return c.req.URL.String()
}

// httpVerifier defines HTTP request and body verification for an HTTP endpoint check
type httpVerifier interface {
	// Request to endpoint
	Request() *http.Request

	// Verify the response body from endpoint
	Verify(body []byte) check.Error
}

// newGetRequest prepares request object for given URL with auth token
func newGetRequest(endpoint, authToken string) (*http.Request, check.Error) {
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, check.ErrUnknown("cannot create request: %s", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))

	return req, nil
}

// doRequest sends the request to the endpoint with passed client
func doRequest(client *http.Client, req *http.Request) ([]byte, check.Error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, check.ErrUnknown("cannot dial %q: %v", req.URL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, check.ErrFail("got HTTP status %q", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, check.ErrFail("cannot read response body: %v", err)
	}

	return body, nil
}

// Insecure transport is useful when kube-rbac-proxy generates self-signed certificates, causing cert validation error
var insecureClient = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	},
}
