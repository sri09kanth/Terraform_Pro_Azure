// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package paths

import (
	"fmt"
)

func parsePathResource(input string) (PathResource, error) {
	switch input {
	case "file":
		return PathResourceFile, nil
	case "directory":
		return PathResourceDirectory, nil
	}
	return "", fmt.Errorf("Unhandled path resource type %q", input)
}
