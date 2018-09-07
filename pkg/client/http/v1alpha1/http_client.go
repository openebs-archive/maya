/*
Copyright 2018 The OpenEBS Authors

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
	"fmt"
	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	. "k8s.io/client-go/rest"
)

type rest struct {
	client *RESTClient // rest client that does the actual API invocation
}

// REST provides a new instance of rest client capable of invoking REST API
// calls
//
// NOTE:
// base url can be of form http://ipaddr:port with no trailing slash
func REST(base string) (r *rest, err error) {
	c, err := RESTClientFor(&Config{
		Host: base,
		ContentConfig: ContentConfig{
			GroupVersion:         &v1.SchemeGroupVersion,
			NegotiatedSerializer: serializer.DirectCodecFactory{CodecFactory: scheme.Codecs},
		},
	})
	if err != nil {
		return
	}
	r = &rest{client: c}
	return
}

func API(verb, baseurl, name string) (b []byte, err error) {
	if len(name) == 0 {
		err = fmt.Errorf("empty resource name: failed to invoke REST API %s %s", verb, baseurl)
		return
	}
	r, err := REST(baseurl)
	if err != nil {
		err = errors.Wrapf(err, "failed to invoke REST API %s %s %s", verb, baseurl, name)
		return
	}
	req := r.client.Verb(verb).Name(name)
	return doRaw(verb, req)
}

func URL(verb, url string) (b []byte, err error) {
	if len(url) == 0 {
		err = fmt.Errorf("empty URL: failed to invoke REST API %s", verb)
		return
	}
	r, err := REST(url)
	if err != nil {
		err = errors.Wrapf(err, "failed to invoke REST API %s %s %s", verb, url)
		return
	}
	// override the formed URL with the one provided
	req := r.client.Verb(verb).RequestURI(url)
	return doRaw(verb, req)
}

func doRaw(verb string, req *Request) (b []byte, err error) {
	b, err = req.DoRaw()
	if err != nil {
		err = errors.Wrapf(err, "error invoking REST API %s %s", verb, req.URL())
		return
	}
	return
}
