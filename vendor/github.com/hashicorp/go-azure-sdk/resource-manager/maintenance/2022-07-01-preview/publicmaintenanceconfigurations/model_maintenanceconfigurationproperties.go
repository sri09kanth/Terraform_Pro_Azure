package publicmaintenanceconfigurations

// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See NOTICE.txt in the project root for license information.

type MaintenanceConfigurationProperties struct {
	ExtensionProperties *map[string]string       `json:"extensionProperties,omitempty"`
	InstallPatches      *InputPatchConfiguration `json:"installPatches,omitempty"`
	MaintenanceScope    *MaintenanceScope        `json:"maintenanceScope,omitempty"`
	MaintenanceWindow   *MaintenanceWindow       `json:"maintenanceWindow,omitempty"`
	Namespace           *string                  `json:"namespace,omitempty"`
	Visibility          *Visibility              `json:"visibility,omitempty"`
}
