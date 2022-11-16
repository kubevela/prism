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

package cue_test

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/emicklei/go-restful/v3"
	"github.com/stretchr/testify/require"
	"k8s.io/apiserver/pkg/server"

	"github.com/kubevela/prism/pkg/cue"
)

func TestRegisterGenericAPIServer(t *testing.T) {
	s := &server.GenericAPIServer{Handler: &server.APIServerHandler{
		GoRestfulContainer: restful.NewContainer(),
	}}
	cue.RegisterGenericAPIServer(s)
}

type FakeResponseWriter struct {
	bytes.Buffer
	StatusCode int
	Bad        bool
}

func (in *FakeResponseWriter) Header() http.Header {
	return map[string][]string{}
}

func (in *FakeResponseWriter) Write(i []byte) (int, error) {
	if in.Bad {
		return 0, io.ErrUnexpectedEOF
	}
	return in.Buffer.Write(i)
}

func (in *FakeResponseWriter) WriteHeader(statusCode int) {
	in.StatusCode = statusCode
}

var _ http.ResponseWriter = &FakeResponseWriter{}

type BadReader struct{}

func (in *BadReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

var _ io.Reader = &BadReader{}

func TestHandleEvalRequest(t *testing.T) {
	cases := map[string]struct {
		Body       []byte
		Path       string
		BadWriter  bool
		StatusCode int
		Output     []byte
	}{
		"bad-request": {
			Body:       nil,
			StatusCode: http.StatusBadRequest,
		},
		"compile-error": {
			Body:       []byte(`bad-key: bad value`),
			StatusCode: http.StatusBadRequest,
		},
		"write-error": {
			Body:       []byte(`x: y: z: 5`),
			BadWriter:  true,
			StatusCode: http.StatusInternalServerError,
		},
		"good": {
			Body:       []byte(`x: y: z: 5`),
			Path:       "x.y",
			StatusCode: http.StatusOK,
			Output:     []byte(`{"z":5}`),
		},
	}
	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			var body io.Reader = bytes.NewReader(tt.Body)
			if tt.Body == nil {
				body = &BadReader{}
			}
			raw, err := http.NewRequest("", "/eval?path="+tt.Path, body)
			require.NoError(t, err)
			request := restful.NewRequest(raw)
			writer := &FakeResponseWriter{Bad: tt.BadWriter}
			response := restful.NewResponse(writer)
			cue.HandleEvalRequest(request, response)
			require.Equal(t, tt.StatusCode, writer.StatusCode)
			if tt.StatusCode == http.StatusOK {
				require.Equal(t, tt.Output, writer.Bytes())
			}
		})
	}
}
