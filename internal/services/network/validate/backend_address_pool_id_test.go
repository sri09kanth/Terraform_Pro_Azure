// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validate

// NOTE: this file is generated via 'go:generate' - manual changes will be overwritten

import "testing"

func TestBackendAddressPoolID(t *testing.T) {
	cases := []struct {
		Input string
		Valid bool
	}{

		{
			// empty
			Input: "",
			Valid: false,
		},

		{
			// missing SubscriptionId
			Input: "/",
			Valid: false,
		},

		{
			// missing value for SubscriptionId
			Input: "/subscriptions/",
			Valid: false,
		},

		{
			// missing ResourceGroup
			Input: "/subscriptions/12345678-1234-9876-4563-123456789012/",
			Valid: false,
		},

		{
			// missing value for ResourceGroup
			Input: "/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/",
			Valid: false,
		},

		{
			// missing ApplicationGatewayName
			Input: "/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/group1/providers/Microsoft.Network/",
			Valid: false,
		},

		{
			// missing value for ApplicationGatewayName
			Input: "/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/group1/providers/Microsoft.Network/applicationGateways/",
			Valid: false,
		},

		{
			// missing Name
			Input: "/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/group1/providers/Microsoft.Network/applicationGateways/applicationGateway1/",
			Valid: false,
		},

		{
			// missing value for Name
			Input: "/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/group1/providers/Microsoft.Network/applicationGateways/applicationGateway1/backendAddressPools/",
			Valid: false,
		},

		{
			// valid
			Input: "/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/group1/providers/Microsoft.Network/applicationGateways/applicationGateway1/backendAddressPools/beap1",
			Valid: true,
		},

		{
			// upper-cased
			Input: "/SUBSCRIPTIONS/12345678-1234-9876-4563-123456789012/RESOURCEGROUPS/GROUP1/PROVIDERS/MICROSOFT.NETWORK/APPLICATIONGATEWAYS/APPLICATIONGATEWAY1/BACKENDADDRESSPOOLS/BEAP1",
			Valid: false,
		},
	}
	for _, tc := range cases {
		t.Logf("[DEBUG] Testing Value %s", tc.Input)
		_, errors := BackendAddressPoolID(tc.Input, "test")
		valid := len(errors) == 0

		if tc.Valid != valid {
			t.Fatalf("Expected %t but got %t", tc.Valid, valid)
		}
	}
}
