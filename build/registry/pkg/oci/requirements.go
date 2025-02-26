// SPDX-License-Identifier: Apache-2.0
/*
Copyright (C) 2023 The Falco Authors.

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

package oci

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/blang/semver"
	"github.com/falcosecurity/falcoctl/pkg/oci"
	"github.com/falcosecurity/plugin-sdk-go/pkg/loader"
	"github.com/falcosecurity/plugins/build/registry/pkg/common"
)

const (
	rulesEngineAnchor = "- required_engine_version"
)

// ErrReqNotFound error when the requirements are not found in the rulesfile.
var ErrReqNotFound = errors.New("requirements not found")

// rulesfileRequirement given a rulesfile in yaml format it scans it and extracts its requirements.
func rulesfileRequirement(filePath string) (*oci.ArtifactRequirement, error) {
	var requirement string
	// Open the file.
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to open file %q: %v", filePath, file)
	}

	defer file.Close()

	// Prepare the file to be read line by line.
	fileScanner := bufio.NewScanner(file)
	fileScanner.Split(bufio.ScanLines)

	for fileScanner.Scan() {
		if strings.HasPrefix(fileScanner.Text(), rulesEngineAnchor) {
			requirement = fileScanner.Text()
			break
		}
	}

	if requirement == "" {
		return nil, fmt.Errorf("requirements for rulesfile %q: %w", filePath, ErrReqNotFound)
	}

	// Split the requirement and parse the version to semVer.
	// In case the requirement was expressed as a numeric value,
	// we convert it to semver and treat it as minor version.
	tokens := strings.Split(fileScanner.Text(), ":")
	reqVer, err := semver.Parse(tokens[1])
	if err != nil {
		reqVer, err = semver.ParseTolerant(tokens[1])
		if err != nil {
			return nil, fmt.Errorf("unable to parse requirement %q: expected a numeric value or a valid semver string", tokens[1])
		}
		reqVer = semver.Version{
			Major: 0,
			Minor: reqVer.Major,
			Patch: 0,
		}
	}

	return &oci.ArtifactRequirement{
		Name:    common.EngineVersionKey,
		Version: reqVer.String(),
	}, nil
}

// pluginRequirement given a plugin as a shared library it loads it and gets the api version
// required by the plugin.
func pluginRequirement(filePath string) (*oci.ArtifactRequirement, error) {
	plugin, err := loader.NewPlugin(filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to open plugin %q: %w", filePath, err)
	}

	return &oci.ArtifactRequirement{
		Name:    common.PluginAPIVersion,
		Version: plugin.Info().RequiredAPIVersion,
	}, nil
}
