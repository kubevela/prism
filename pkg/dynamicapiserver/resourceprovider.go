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

package apiserver

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	"github.com/kubevela/prism/pkg/apis/dynamicresource"
	"github.com/kubevela/prism/pkg/util/singleton"
)

const (
	defaultNamespace = "vela-system"
	encoderKey       = "encoder"
	decoderKey       = "decoder"
)

func StartDynamicResourceFactoryWithConfigMapInformer(stopCh <-chan struct{}) {
	factory := informers.NewSharedInformerFactoryWithOptions(
		singleton.StaticClient.Get(), 0,
		informers.WithNamespace(defaultNamespace))
	informer := factory.Core().V1().ConfigMaps().Informer()
	defer runtime.HandleCrash()
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if err := addResource(obj); err != nil {
				klog.Error(err)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if err := removeResource(oldObj); err != nil {
				klog.Error(err)
			}
			if err := addResource(newObj); err != nil {
				klog.Error(err)
			}
		},
		DeleteFunc: func(obj interface{}) {
			if err := removeResource(obj); err != nil {
				klog.Error(err)
			}
		},
	})
	informer.Run(stopCh)
}

func handleResource(obj interface{}, action string, handler func(ResourceProvider) error) error {
	var r ResourceProvider
	var err error
	switch o := obj.(type) {
	case *corev1.ConfigMap:
		r, err = newDynamicResourceFromConfigMap(o)
	default:
		return fmt.Errorf("cannot recognize %T type", obj)
	}
	if err != nil {
		return err
	}
	if r == nil {
		return nil
	}
	klog.Infof("Handle Resource %s.%s (%s)", r.GetGroupVersionResource().Resource, r.GetGroupVersion(), action)
	return handler(r)
}

func addResource(obj interface{}) error {
	return handleResource(obj, "Add", DefaultDynamicAPIServer.AddResource)
}

func removeResource(obj interface{}) error {
	return handleResource(obj, "Remove", DefaultDynamicAPIServer.RemoveResource)
}

func newDynamicResourceFromConfigMap(cm *corev1.ConfigMap) (obj ResourceProvider, err error) {
	encoderTemplate, encoderExists := cm.Data[encoderKey]
	decoderTemplate, decoderExists := cm.Data[decoderKey]
	if !encoderExists || !decoderExists {
		return nil, nil
	}
	obj, err = dynamicresource.NewDynamicResourceWithCodec(encoderTemplate, decoderTemplate)
	return obj, err
}
