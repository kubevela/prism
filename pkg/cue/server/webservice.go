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

package server

import (
	"io"
	"net/http"

	"github.com/emicklei/go-restful/v3"
	"k8s.io/apiserver/pkg/server"
)

const (
	webserviceRootPath            = "/cue"
	webserviceEvalPath            = "/eval"
	webserviceParameterKeyPath    = "path"
	webserviceParameterKeyCompile = "compile"
)

func RegisterGenericAPIServer(server *server.GenericAPIServer) *server.GenericAPIServer {
	ws := &restful.WebService{}
	ws.Path(webserviceRootPath)
	ws.Route(ws.POST(webserviceEvalPath).To(HandleEvalRequest))
	server.Handler.GoRestfulContainer.Add(ws)
	return server
}

func HandleEvalRequest(request *restful.Request, response *restful.Response) {
	bs, err := io.ReadAll(request.Request.Body)
	if err != nil {
		_ = response.WriteError(http.StatusBadRequest, err)
		return
	}
	var path []string
	if p := request.QueryParameter(webserviceParameterKeyPath); len(p) > 0 {
		path = append(path, p)
	}
	f := Render
	if p := request.QueryParameter(webserviceParameterKeyCompile); len(p) > 0 {
		f = Compile
	}
	res, err := f(bs, path...)
	if err != nil {
		_ = response.WriteError(http.StatusBadRequest, err)
		return
	}
	response.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	response.WriteHeader(http.StatusOK)
	if _, err = response.Write(res); err != nil {
		_ = response.WriteError(http.StatusInternalServerError, err)
		return
	}
}
