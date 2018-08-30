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
	menv "github.com/openebs/maya/pkg/env/v1alpha1"
)

// EnvStatus represents the status of operation against an env instance
type EnvStatus string

const (
	EnvSetSuccess EnvStatus = "set succeeded"
	EnvSetErr     EnvStatus = "set failed"
	EnvSetSkip    EnvStatus = "set skipped"
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

// envPredicate abstracts evaluation condition of the given instance and returns
// the name of evaluation along with result of evaluation
type envPredicate func(given *env) (name string, success bool)

// isEnvNotPresent returns true if env in not set
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

// envMiddleware abstracts updating the given env instance
type envMiddleware func(given *env) (updated *env)

// EnvUpdateStatus updates the env instance with provided status info
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

// EnvUpdateError updates the env instance with provided error
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

// envSetP executes set operation conditionally on the given env instance
func envSetP(ctx string, p envPredicate) envMiddleware {
	return func(given *env) (updated *env) {
		pCtx, pOk := p(given)
		if !pOk {
			return EnvUpdateStatus(ctx, pCtx, EnvSetSkip)(given)
		}
		err := menv.Set(given.Key, given.Value)
		if err != nil {
			return EnvUpdateError(ctx, err)(given)
		}
		return EnvUpdateStatus(ctx, "env set was successful", EnvSetSuccess)(given)
	}
}

// SetP executes set operation conditionally on the given env instance
func (e *env) SetP(ctx string, p envPredicate) (u *env) {
	return envSetP(ctx, p)(e)
}

// envList represents a list of environment variables
type envList struct {
	Items []*env
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
		msgs = append(msgs, fmt.Sprintf("env '%s': val '%s': status: '%s' '%s' '%s'", env.Context, env.Status, env.Reason))
	}
	return
}

// envListMiddleware abstracts updating of envList instance
type envListMiddleware func(given *envList) (updated *envList)

// envListSetP executes set operation conditionally on all the list of env
// instances
func envListSetP(ctx string, p envPredicate) envListMiddleware {
	return func(given *envList) (updated *envList) {
		if given == nil {
			return
		}
		updated = &envList{}
		for _, env := range given.Items {
			if env == nil {
				continue
			}
			updated.Items = append(updated.Items, env.SetP(ctx, p))
		}
		return
	}
}

// SetP executes set operation conditionally on all the list of env instances
func (l *envList) SetP(ctx string, p envPredicate) (u *envList) {
	return envListSetP(ctx, p)(l)
}
