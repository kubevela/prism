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

package options

import (
	"time"

	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apiserver/pkg/server"
)

// ServerRunOptions is the extension option for configuring the APIServer
type ServerRunOptions struct {
	RequestTimeout time.Duration
}

func NewServerRunOptions() *ServerRunOptions {
	defaults := server.NewConfig(serializer.CodecFactory{})
	return &ServerRunOptions{
		RequestTimeout: defaults.RequestTimeout,
	}
}

// ApplyTo set the params in server.RecommendConfig
func (s *ServerRunOptions) ApplyTo(c *server.Config) error {
	c.RequestTimeout = s.RequestTimeout
	return nil
}

// AddFlags add flags for a specific APIServer to the specified FlagSet
func (s *ServerRunOptions) AddFlags(fs *pflag.FlagSet) {
	fs.DurationVar(&s.RequestTimeout, "request-timeout", s.RequestTimeout, ""+
		"An optional field indicating the duration a handler must keep a request open before timing "+
		"it out. This is the default request timeout for requests but may be overridden by flags such as "+
		"--min-request-timeout for specific types of requests.")
}

var defaultServerRunOptions = NewServerRunOptions()

// WrapConfig wraps server.RecommendedConfig with default ServerRunOptions
func WrapConfig(config *server.RecommendedConfig) *server.RecommendedConfig {
	runtime.Must(defaultServerRunOptions.ApplyTo(&config.Config))
	return config
}

// AddServerRunFlags add ServerRunOptions flags to pflag.FlagSet
func AddServerRunFlags(fs *pflag.FlagSet) {
	defaultServerRunOptions.AddFlags(fs)
}
