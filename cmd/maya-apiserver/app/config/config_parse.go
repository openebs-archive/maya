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

package config

import (
	"bytes"
	"io"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

// ParseMayaConfigFile parses the given path as
// maya config
func ParseMayaConfigFile(path string) (*MayaConfig, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse maya config at '%s'", path)
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse maya config at '%s'", path)
	}
	defer f.Close()

	mconfig, err := ParseMayaConfig(f)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse maya config at '%s'", path)
	}

	return mconfig, nil
}

// ParseMayaConfig parses the config from the given io.Reader.
//
// Due to current internal limitations, the entire contents of
// io.Reader will be copied into memory first before parsing.
func ParseMayaConfig(r io.Reader) (*MayaConfig, error) {
	// Copy the reader into an in-memory buffer first
	// since HCL requires it.
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		return nil, errors.Wrap(err, "failed to parse maya config: failed to copy into buffer")
	}

	// Parse the buffer
	bufs := buf.String()
	root, err := hcl.Parse(bufs)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse maya config: %s", bufs)
	}
	buf.Reset()

	// Top-level item should be a list
	list, ok := root.Node.(*ast.ObjectList)
	if !ok {
		return nil, errors.Errorf("failed to parse maya config: root should be an object")
	}

	var mconfig MayaConfig
	err = parseConfig(&mconfig, list)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse maya config: %+v", list)
	}

	return &mconfig, nil
}

func parseConfig(result *MayaConfig, list *ast.ObjectList) error {
	// Check for invalid keys
	valid := []string{
		"region",
		"datacenter",
		"name",
		"data_dir",
		"log_level",
		"bind_addr",
		"enable_debug",
		"ports",
		"addresses",
		"interfaces",
		"advertise",
		"leave_on_interrupt",
		"leave_on_terminate",
		"enable_syslog",
		"syslog_facility",
		"http_api_response_headers",
	}
	err := checkHCLKeys(list, valid)
	if err != nil {
		return errors.Wrapf(err, "failed to parse maya config: invalid keys found {%+v} : supported keys {%+v}", list, valid)
	}

	// Decode the full thing into a map[string]interface for ease
	var m map[string]interface{}
	err = hcl.DecodeObject(&m, list)
	if err != nil {
		return errors.Wrapf(err, "failed to parse maya config: config keys {%+v}", list)
	}
	delete(m, "ports")
	delete(m, "addresses")
	delete(m, "interfaces")
	delete(m, "advertise")
	delete(m, "http_api_response_headers")

	// Decode the rest
	err = mapstructure.WeakDecode(m, result)
	if err != nil {
		return errors.Wrapf(err, "failed to parse maya config: %+v", m)
	}

	// Parse ports
	if o := list.Filter("ports"); len(o.Items) > 0 {
		err := parsePorts(&result.Ports, o)
		if err != nil {
			return errors.Wrapf(err, "failed to parse maya config: failed to parse ports: %+v", o)
		}
	}

	// Parse addresses
	if o := list.Filter("addresses"); len(o.Items) > 0 {
		err := parseAddresses(&result.Addresses, o)
		if err != nil {
			return errors.Wrapf(err, "failed to parse maya config: failed to parse addresses: %+v", o)
		}
	}

	// Parse advertise
	if o := list.Filter("advertise"); len(o.Items) > 0 {
		err := parseAdvertise(&result.AdvertiseAddrs, o)
		if err != nil {
			return errors.Wrapf(err, "failed to parse maya config: failed to parse advertise addresses: %+v", o)
		}
	}

	// Parse out http_api_response_headers fields. These are in HCL as a list so
	// we need to iterate over them and merge them.
	if headersO := list.Filter("http_api_response_headers"); len(headersO.Items) > 0 {
		for _, o := range headersO.Elem().Items {
			var m map[string]interface{}
			err := hcl.DecodeObject(&m, o.Val)
			if err != nil {
				return errors.Wrapf(err, "failed to parse maya config: failed to parse http response header: %+v", o)
			}
			err = mapstructure.WeakDecode(m, &result.HTTPAPIResponseHeaders)
			if err != nil {
				return errors.Wrapf(err, "failed to parse maya config: failed to parse http response header: %+v", m)
			}
		}
	}

	return nil
}

func parsePorts(result **Ports, list *ast.ObjectList) error {
	list = list.Elem()
	if len(list.Items) > 1 {
		return errors.Errorf("failed to parse ports: only one 'ports' block allowed")
	}

	// Get our ports object
	listVal := list.Items[0].Val

	// Check for invalid keys
	valid := []string{
		"http",
	}
	err := checkHCLKeys(listVal, valid)
	if err != nil {
		return errors.Wrapf(err, "failed to parse ports: invalid keys found {%+v}: supported keys {%+v}", listVal, valid)
	}

	var m map[string]interface{}
	err = hcl.DecodeObject(&m, listVal)
	if err != nil {
		return errors.Wrapf(err, "failed to parse ports: %+v", listVal)
	}

	var ports Ports
	err = mapstructure.WeakDecode(m, &ports)
	if err != nil {
		return errors.Wrapf(err, "failed to parse ports: %+v", m)
	}
	*result = &ports
	return nil
}

func parseAddresses(result **Addresses, list *ast.ObjectList) error {
	list = list.Elem()
	if len(list.Items) > 1 {
		return errors.Errorf("failed to parse addresses: only one 'addresses' block allowed")
	}

	// Get our addresses object
	listVal := list.Items[0].Val

	// Check for invalid keys
	valid := []string{
		"http",
	}
	err := checkHCLKeys(listVal, valid)
	if err != nil {
		return errors.Wrapf(err, "failed to parse addresses: invalid keys found {%+v}: supported keys {%+v}", listVal, valid)
	}

	var m map[string]interface{}
	err = hcl.DecodeObject(&m, listVal)
	if err != nil {
		return errors.Wrapf(err, "failed to parse addresses: %+v", listVal)
	}

	var addresses Addresses
	err = mapstructure.WeakDecode(m, &addresses)
	if err != nil {
		return errors.Wrapf(err, "failed to parse addresses: %+v", m)
	}
	*result = &addresses
	return nil
}

func parseAdvertise(result **AdvertiseAddrs, list *ast.ObjectList) error {
	list = list.Elem()
	if len(list.Items) > 1 {
		return errors.Errorf("failed to parse advertise: only one 'advertise' block allowed")
	}

	// Get our advertise object
	listVal := list.Items[0].Val

	// Check for invalid keys
	valid := []string{
		"http",
	}
	err := checkHCLKeys(listVal, valid)
	if err != nil {
		return errors.Wrapf(err, "failed to parse advertise: invalid keys found {%+v}: supported keys {%+v}", listVal, valid)
	}

	var m map[string]interface{}
	err = hcl.DecodeObject(&m, listVal)
	if err != nil {
		return errors.Wrapf(err, "failed to parse advertise: %+v", listVal)
	}

	var advertise AdvertiseAddrs
	err = mapstructure.WeakDecode(m, &advertise)
	if err != nil {
		return errors.Wrapf(err, "failed to parse advertise: %+v", m)
	}
	*result = &advertise
	return nil
}

func checkHCLKeys(node ast.Node, valid []string) error {
	var list *ast.ObjectList
	switch n := node.(type) {
	case *ast.ObjectList:
		list = n
	case *ast.ObjectType:
		list = n.List
	default:
		return errors.Errorf("failed to check HCL keys: unsupported type '%T'", n)
	}

	validMap := make(map[string]struct{}, len(valid))
	for _, v := range valid {
		validMap[v] = struct{}{}
	}

	var errs []error
	for _, item := range list.Items {
		key := item.Keys[0].Token.Value().(string)
		if _, ok := validMap[key]; !ok {
			errs = append(errs, errors.Errorf("failed to check HCL keys: invalid key '%s'", key))
		}
	}

	if len(errs) > 0 {
		return errors.Errorf("%+v", errs)
	}
	return nil
}
