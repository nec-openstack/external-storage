/*
Copyright 2016 The Kubernetes Authors.

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

package volume

import "strings"

// ProvisionerConfig provisioner config for Provision Volume
type ProvisionerConfig struct {
	BrickRootPaths []string
	VolumeType     string
}

type validateError struct {
	reason string
}

func (error *validateError) Error() string {
	return error.reason
}

// NewProvisionerConfig create ProvisionerConfig from parameters of StorageClass
func NewProvisionerConfig(params map[string]string) (*ProvisionerConfig, error) {
	var config ProvisionerConfig

	// Set default volume type
	volumeType := ""
	var brickRootPaths []string

	for k, v := range params {
		switch strings.ToLower(k) {
		case "brickrootpaths":
			brickRootPaths = parseBrickRootPaths(v)
		}
	}

	config.BrickRootPaths = brickRootPaths
	config.VolumeType = volumeType

	err := config.validate()
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func parseBrickRootPaths(param string) []string {
	brickRootPaths := strings.Split(param, ",")
	for i, path := range brickRootPaths {
		brickRootPaths[i] = strings.TrimSpace(path)
	}

	return brickRootPaths
}

func (config *ProvisionerConfig) validate() error {
	if len(config.BrickRootPaths) == 0 {
		return &validateError{reason: "brickRootPaths are not specified"}
	}

	return nil
}
