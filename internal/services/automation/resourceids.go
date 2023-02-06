// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package automation

//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=Connection -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/group1/providers/Microsoft.Automation/automationAccounts/account1/connections/connection1
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=ConnectionType -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/group1/providers/Microsoft.Automation/automationAccounts/account1/connectionTypes/type1
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=AutomationAccount -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/group1/providers/Microsoft.Automation/automationAccounts/account1
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=Certificate -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/group1/providers/Microsoft.Automation/automationAccounts/account1/certificates/cert1
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=Credential -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/group1/providers/Microsoft.Automation/automationAccounts/account1/credentials/cred1
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=Schedule -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/group1/providers/Microsoft.Automation/automationAccounts/account1/schedules/schedule1
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=SourceControl -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/group1/providers/Microsoft.Automation/automationAccounts/account1/sourceControls/sample -rewrite=true
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=Module -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/group1/providers/Microsoft.Automation/automationAccounts/account1/modules/module1
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=NodeConfiguration -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/group1/providers/Microsoft.Automation/automationAccounts/account1/nodeConfigurations/nodeconfig1
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=Runbook -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/group1/providers/Microsoft.Automation/automationAccounts/account1/runbooks/runbook1
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=SoftwareUpdateConfiguration -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/group1/providers/Microsoft.Automation/automationAccounts/account1/softwareUpdateConfigurations/up1
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=Configuration -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/group1/providers/Microsoft.Automation/automationAccounts/account1/configurations/config1
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=JobSchedule -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/group1/providers/Microsoft.Automation/automationAccounts/account1/jobSchedules/schedule1
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=Variable -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/group1/providers/Microsoft.Automation/automationAccounts/account1/variables/variable1
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=Webhook -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/group1/providers/Microsoft.Automation/automationAccounts/account1/webHooks/webhook1 -rewrite=true
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=Watcher -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/group1/providers/Microsoft.Automation/automationAccounts/account1/watchers/watcher1
