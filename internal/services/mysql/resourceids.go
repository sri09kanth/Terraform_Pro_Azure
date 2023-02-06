// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mysql

//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=AzureActiveDirectoryAdministrator -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.DBforMySQL/servers/server1/administrators/activeDirectory
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=Configuration -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.DBforMySQL/servers/server1/configurations/config1
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=Database -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.DBforMySQL/servers/server1/databases/database1
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=FirewallRule -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.DBforMySQL/servers/server1/firewallRules/firewallRule1
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=FlexibleServer -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.DBforMySQL/flexibleServers/flexibleServer1
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=FlexibleDatabase -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.DBforMySQL/flexibleServers/flexibleServer1/databases/database1
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=FlexibleServerConfiguration -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.DBforMySQL/flexibleServers/flexibleServer1/configurations/config1
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=FlexibleServerFirewallRule -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.DBforMySQL/flexibleServers/flexibleServer1/firewallRules/firewallRule1
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=Key -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.DBforMySQL/servers/server1/keys/key1
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=Server -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.DBforMySQL/servers/server1
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=VirtualNetworkRule -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.DBforMySQL/servers/server1/virtualNetworkRules/virtualNetworkRule1
