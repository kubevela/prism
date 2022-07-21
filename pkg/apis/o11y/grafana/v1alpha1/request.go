/*
Copyright 2022 The KubeVela Authors.

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

package v1alpha1

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/resource"

	"github.com/kubevela/prism/pkg/util/subresource"
)

// DoRequest do request for the current grafana
func (in *Grafana) DoRequest(ctx context.Context, method string, path string, body io.Reader) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, method, strings.Trim(in.Spec.Endpoint, "/")+path, body)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	// set headers
	switch {
	case in.Spec.Access.Token != nil:
		req.Header.Set("Authorization", "Bearer "+*in.Spec.Access.Token)
	case in.Spec.Access.BasicAuth != nil:
		req.SetBasicAuth(in.Spec.Access.Username, in.Spec.Access.Password)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	defer func() { _ = resp.Body.Close() }()
	bs, err := io.ReadAll(resp.Body)
	return bs, resp.StatusCode, err
}

// GrafanaSubResourceRequest request for grafana subresources
// +kubebuilder:object:generate=false
type GrafanaSubResourceRequest struct {
	resourceName *subresource.CompoundName
	subResource  resource.Object

	method    string
	pathFunc  func() (string, error)
	bodyFunc  func() ([]byte, error)
	onSuccess func(respBody []byte) error
}

// NewGrafanaSubResourceRequest create request for grafana subresource
func NewGrafanaSubResourceRequest(subResource resource.Object, name string) *GrafanaSubResourceRequest {
	return &GrafanaSubResourceRequest{
		resourceName: subresource.NewCompoundName(name),
		subResource:  subResource,
	}
}

func (in *GrafanaSubResourceRequest) WithMethod(method string) *GrafanaSubResourceRequest {
	in.method = method
	return in
}

func (in *GrafanaSubResourceRequest) WithPathFunc(pathFunc func() (string, error)) *GrafanaSubResourceRequest {
	in.pathFunc = pathFunc
	return in
}

func (in *GrafanaSubResourceRequest) WithBodyFunc(bodyFunc func() ([]byte, error)) *GrafanaSubResourceRequest {
	in.bodyFunc = bodyFunc
	return in
}

func (in *GrafanaSubResourceRequest) WithOnSuccess(onSuccess func(respBody []byte) error) *GrafanaSubResourceRequest {
	in.onSuccess = onSuccess
	return in
}

func (in *GrafanaSubResourceRequest) Do(ctx context.Context, cli GrafanaClient) error {
	parent, err := cli.Get(ctx, in.resourceName.ParentResourceName)
	if err != nil {
		return err
	}

	var path string
	if in.pathFunc != nil {
		if path, err = in.pathFunc(); err != nil {
			return err
		}
	}
	var body io.Reader = nil
	if in.bodyFunc != nil {
		var bs []byte
		if bs, err = in.bodyFunc(); err != nil {
			return err
		}
		body = bytes.NewReader(bs)
	}

	respBody, statusCode, err := parent.DoRequest(ctx, in.method, path, body)
	if err != nil {
		return err
	}
	switch statusCode {
	case http.StatusOK:
		if in.onSuccess != nil {
			return in.onSuccess(respBody)
		}
	case http.StatusUnauthorized:
		return errors.NewUnauthorized(string(respBody))
	case http.StatusForbidden:
		return errors.NewForbidden(in.subResource.GetGroupVersionResource().GroupResource(), in.resourceName.String(), fmt.Errorf(string(respBody)))
	case http.StatusNotFound:
		return errors.NewNotFound(in.subResource.GetGroupVersionResource().GroupResource(), in.resourceName.String())
	case http.StatusPreconditionFailed:
		return errors.NewBadRequest(string(respBody))
	default:
		return fmt.Errorf("request grafana %s failed, path: %s, code: %d, detail: %s", parent.Spec.Endpoint, path, statusCode, respBody)
	}
	return nil
}
