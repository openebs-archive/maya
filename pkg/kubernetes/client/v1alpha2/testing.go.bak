/*
Copyright 2019 The OpenEBS Authors

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

package v1alpha2

import (
	"context"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// FakeGetOk fakes kubernetes get API call
func FakeGetOk() GetFunc {
	return func(c client.Client, ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
		return nil
	}
}

// FakeGetErr fakes kubernetes get API call
// and always returns an error
func FakeGetErr() GetFunc {
	return func(c client.Client, ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
		return errors.New("fake get error")
	}
}

// FakeCreateOk fakes kubernetes create API
// call
func FakeCreateOk() CreateFunc {
	return func(c client.Client, ctx context.Context, obj runtime.Object) error {
		return nil
	}
}

// FakeCreateErr fakes kubernetes create API
// call and always returns an error
func FakeCreateErr() CreateFunc {
	return func(c client.Client, ctx context.Context, obj runtime.Object) error {
		return errors.New("fake create error")
	}
}

// FakeDeleteOk fakes kubernetes delete API
// call
func FakeDeleteOk() DeleteFunc {
	return func(c client.Client, ctx context.Context, obj runtime.Object) error {
		return nil
	}
}

// FakeDeleteErr fakes kubernetes delete API
// call and always returns an error
func FakeDeleteErr() DeleteFunc {
	return func(c client.Client, ctx context.Context, obj runtime.Object) error {
		return errors.New("fake delete error")
	}
}

// FakeGetClientOk fakes kubernetes getClient API
// call
func FakeGetClientOk() GetClientFunc {
	return func() (client.Client, error) {
		return nil, nil
	}
}

// FakeGetClientErr fakes kubernetes getClient API
// call and always returns an error
func FakeGetClientErr() GetClientFunc {
	return func() (client.Client, error) {
		return nil, errors.New("fake client error")
	}
}
