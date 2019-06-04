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
	"strings"

	menv "github.com/openebs/maya/pkg/env/v1alpha1"
	ver "github.com/openebs/maya/pkg/version"
)

// EnvStatus represents the status of operation
// against an env instance
type EnvStatus string

// env set completion statuses
const (
	EnvSetSuccess EnvStatus = "set env succeeded"
	EnvSetErr     EnvStatus = "set env failed"
	EnvSetSkip    EnvStatus = "set env skipped"
)

// env represents a environment variable
type env struct {
	Key     menv.ENVKey
	Value   string
	Status  EnvStatus
	Reason  string
	Context string
	Err     error
}

// envPredicate abstracts evaluation condition of
// the given instance and returns the name of
// evaluation along with result of evaluation
type envPredicate func(given *env) (name string, success bool)

// isEnvNotPresent returns true if env in not set
// previously
func isEnvNotPresent(given *env) (name string, success bool) {
	name = "isEnvNotPresent"
	if given == nil {
		return
	}
	_, present := menv.Lookup(given.Key)
	if !present {
		success = true
	}
	return
}

// isEnvError returns true if env has error
func isEnvError(given *env) (name string, hasErr bool) {
	name = "isEnvError"
	if given == nil {
		return
	}
	if given.Err != nil {
		hasErr = true
	}
	return
}

// envMiddleware abstracts updating given env instance
type envMiddleware func(given *env) (updated *env)

// EnvUpdateStatus updates the env instance with
// provided status info
func EnvUpdateStatus(context, reason string, status EnvStatus) envMiddleware {
	return func(given *env) (updated *env) {
		if given == nil {
			return
		}
		updated = &env{Key: given.Key, Value: given.Value}
		updated.Context = context
		updated.Reason = reason
		updated.Status = status
		return
	}
}

// EnvUpdateError updates the env instance with
// provided error
func EnvUpdateError(context string, err error) envMiddleware {
	return func(given *env) (updated *env) {
		if given == nil || err == nil {
			return
		}
		updated = EnvUpdateStatus(context, err.Error(), EnvSetErr)(given)
		updated.Err = err
		return
	}
}

// EnvUpdateSuccess updates the env instance with
// success status
func EnvUpdateSuccess(context string) envMiddleware {
	return func(given *env) (updated *env) {
		return EnvUpdateStatus(context, "", EnvSetSuccess)(given)
	}
}

// SetIf executes set conditionally on the given env instance
func (e *env) SetIf(ctx string, p envPredicate) (u *env) {
	pCtx, pOk := p(e)
	if !pOk {
		return EnvUpdateStatus(ctx, pCtx+" predicate failed", EnvSetSkip)(e)
	}
	err := menv.Set(e.Key, e.Value)
	if err != nil {
		return EnvUpdateError(ctx, err)(e)
	}
	return EnvUpdateSuccess(ctx)(e)
}

// envList represents a list of environment variables
type envList struct {
	Items []*env
}

// EnvLister abstracts listing environment variables
type EnvLister interface {
	List() (l *envList, err error)
}

// Errors returns the list of errors present in env instances
func (l *envList) Errors() (errs []error) {
	if l == nil {
		return
	}
	for _, env := range l.Items {
		if env != nil && env.Err != nil {
			errs = append(errs, env.Err)
		}
	}
	return
}

// Infos returns the list of infos present in env instances
func (l *envList) Infos() (msgs []string) {
	if l == nil {
		return
	}
	for _, env := range l.Items {
		if env == nil || env.Err != nil {
			continue
		}
		msgs = append(
			msgs,
			fmt.Sprintf(
				"{env '%s': val '%s': msg: '%s' '%s' '%s'}",
				env.Key,
				env.Value,
				env.Context,
				env.Status,
				env.Reason,
			),
		)
	}
	return
}

// SetIf executes set conditionally on its env instances
func (l *envList) SetIf(ctx string, p envPredicate) (u *envList) {
	if l == nil {
		return
	}
	u = &envList{}
	for _, env := range l.Items {
		if env == nil {
			continue
		}
		u.Items = append(u.Items, env.SetIf(ctx, p))
	}
	return
}

// envInstall manages environment variables required for openebs install
type envInstall struct{}

// EnvInstall returns a new instance of envInstall
func EnvInstall() *envInstall { return &envInstall{} }

// List returns a list of env instances required for openebs install
func (e *envInstall) List() (l *envList, err error) {
	l = &envList{}
	l.Items = append(l.Items, &env{
		Key:   menv.OpenEBSVersion,
		Value: ver.Current(),
	})
	l.Items = append(l.Items, &env{
		Key:   DefaultCstorSparsePool,
		Value: "false",
	})
	l.Items = append(l.Items, &env{
		Key:   menv.CASTemplateFeatureGateENVK,
		Value: "true",
	})
	l.Items = append(l.Items, &env{
		Key:   menv.CASTemplateToCreateJivaVolumeENVK,
		Value: ver.WithSuffix("jiva-volume-create-default"),
	})
	l.Items = append(l.Items, &env{
		Key:   menv.CASTemplateToReadJivaVolumeENVK,
		Value: ver.WithSuffix("jiva-volume-read-default"),
	})
	l.Items = append(l.Items, &env{
		Key:   menv.CASTemplateToDeleteJivaVolumeENVK,
		Value: ver.WithSuffix("jiva-volume-delete-default"),
	})
	l.Items = append(l.Items, &env{
		Key:   menv.CASTemplateToCreateCStorVolumeENVK,
		Value: ver.WithSuffix("cstor-volume-create-default"),
	})
	l.Items = append(l.Items, &env{
		Key:   menv.CASTemplateToReadCStorVolumeENVK,
		Value: ver.WithSuffix("cstor-volume-read-default"),
	})
	l.Items = append(l.Items, &env{
		Key:   menv.CASTemplateToDeleteCStorVolumeENVK,
		Value: ver.WithSuffix("cstor-volume-delete-default"),
	})
	l.Items = append(l.Items, &env{
		Key:   menv.CASTemplateToCreatePoolENVK,
		Value: ver.WithSuffix("cstor-pool-create-default"),
	})
	l.Items = append(l.Items, &env{
		Key:   menv.CASTemplateToDeletePoolENVK,
		Value: ver.WithSuffix("cstor-pool-delete-default"),
	})
	l.Items = append(l.Items, &env{
		Key: menv.CASTemplateToListVolumeENVK,
		Value: strings.Join(ver.WithSuffixesIf(
			[]string{
				"jiva-volume-list-default-0.6.0",
				"jiva-volume-list-default",
				"cstor-volume-list-default",
			},
			ver.IsNotVersioned), ","),
	})
	l.Items = append(l.Items, &env{
		Key:   menv.CASTemplateToCreateCStorSnapshotENVK,
		Value: ver.WithSuffix("cstor-snapshot-create-default"),
	})
	l.Items = append(l.Items, &env{
		Key:   menv.CASTemplateToDeleteCStorSnapshotENVK,
		Value: ver.WithSuffix("cstor-snapshot-delete-default"),
	})
	l.Items = append(l.Items, &env{
		Key:   menv.CASTemplateToCreateJivaSnapshotENVK,
		Value: ver.WithSuffix("jiva-snapshot-create-default"),
	})
	l.Items = append(l.Items, &env{
		Key:   menv.CASTemplateToDeleteJivaSnapshotENVK,
		Value: ver.WithSuffix("jiva-snapshot-delete-default"),
	})
	l.Items = append(l.Items, &env{
		Key:   menv.CASTemplateToReadVolumeStatsENVK,
		Value: ver.WithSuffix("cas-volume-stats-default"),
	})
	l.Items = append(l.Items, &env{
		Key:   menv.CASTemplateToListStoragePoolENVK,
		Value: ver.WithSuffix("storage-pool-list-default"),
	})
	l.Items = append(l.Items, &env{
		Key:   menv.CASTemplateToReadStoragePoolENVK,
		Value: ver.WithSuffix("storage-pool-read-default"),
	})
	return
}
