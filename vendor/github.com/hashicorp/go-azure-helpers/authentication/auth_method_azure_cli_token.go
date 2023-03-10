// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package authentication

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure/cli"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/go-version"
	"github.com/manicminer/hamilton/environments"
)

type azureCLIProfile struct {
	// CLI "subscriptions" are really "accounts" that can represent either a subscription (with tenant) or _just_ a tenant
	account *cli.Subscription

	clientId       string
	environment    string
	subscriptionId string
	tenantId       string
	tenantOnly     bool

	azVersion version.Version
}

type azureCliTokenAuth struct {
	profile                      *azureCLIProfile
	servicePrincipalAuthDocsLink string
}

func (a azureCliTokenAuth) build(b Builder) (authMethod, error) {
	ver, err := populateAzVersion(b.TenantOnly)
	if err != nil {
		return nil, err
	}
	auth := azureCliTokenAuth{
		profile: &azureCLIProfile{
			subscriptionId: b.SubscriptionID,
			tenantId:       b.TenantID,
			tenantOnly:     b.TenantOnly,
			clientId:       "04b07795-8ddb-461a-bbee-02f9e1bf7b46", // fixed first party client id for Az CLI
			azVersion:      *ver,
		},
		servicePrincipalAuthDocsLink: b.ClientSecretDocsLink,
	}

	var acc *cli.Subscription
	if auth.profile.tenantOnly {
		var err error
		acc, err = obtainTenant(b.TenantID)
		if err != nil {
			return nil, fmt.Errorf("obtain tenant(%s) from Azure CLI: %+v", b.TenantID, err)
		}
		auth.profile.account = acc
	} else {
		var err error
		acc, err = obtainSubscription(b.SubscriptionID)
		if err != nil {
			return nil, fmt.Errorf("obtain subscription(%s) from Azure CLI: %+v", b.SubscriptionID, err)
		}
		auth.profile.account = acc
	}

	// Authenticating as a Service Principal doesn't return all of the information we need for authentication purposes
	// as such Service Principal authentication is supported using the specific auth method
	if acc.User == nil || !strings.EqualFold(acc.User.Type, "user") {
		return nil, fmt.Errorf(`Authenticating using the Azure CLI is only supported as a User (not a Service Principal).

To authenticate to Azure using a Service Principal, you can use the separate 'Authenticate using a Service Principal'
auth method - instructions for which can be found here: %s

Alternatively you can authenticate using the Azure CLI by using a User Account.`, auth.servicePrincipalAuthDocsLink)
	}

	// Populate fields
	if !b.TenantOnly && auth.profile.subscriptionId == "" {
		auth.profile.subscriptionId = acc.ID
	}
	if auth.profile.tenantId == "" {
		auth.profile.tenantId = acc.TenantID
	}
	// always pull the environment from the Azure CLI, since the Access Token's associated with it
	auth.profile.environment = normalizeEnvironmentName(acc.EnvironmentName)

	return auth, nil
}

func (a azureCliTokenAuth) isApplicable(b Builder) bool {
	return b.SupportsAzureCliToken
}

func (a azureCliTokenAuth) getADALToken(_ context.Context, _ autorest.Sender, oauthConfig *OAuthConfig, endpoint string) (autorest.Authorizer, error) {
	if oauthConfig.OAuth == nil {
		return nil, fmt.Errorf("getting Authorization Token for cli auth: an OAuth token wasn't configured correctly; please file a bug with more details")
	}

	// the Azure CLI appears to cache these, so to maintain compatibility with the interface this method is intentionally not on the pointer
	var token *cli.Token
	var err error
	if a.profile.tenantOnly {
		token, err = obtainAuthorizationToken(endpoint, "", a.profile.tenantId)
	} else {
		token, err = obtainAuthorizationToken(endpoint, a.profile.subscriptionId, "")
	}
	if err != nil {
		return nil, fmt.Errorf("obtaining Authorization Token from the Azure CLI: %s", err)
	}

	adalToken, err := token.ToADALToken()
	if err != nil {
		return nil, fmt.Errorf("converting Authorization Token to an ADAL Token: %s", err)
	}

	spt, err := adal.NewServicePrincipalTokenFromManualToken(*oauthConfig.OAuth, a.profile.clientId, endpoint, adalToken)
	if err != nil {
		return nil, err
	}

	var refreshFunc adal.TokenRefresh = func(ctx context.Context, resource string) (*adal.Token, error) {
		var token *cli.Token
		var err error
		if a.profile.tenantOnly {
			token, err = obtainAuthorizationToken(resource, "", a.profile.tenantId)
		} else {
			token, err = obtainAuthorizationToken(resource, a.profile.subscriptionId, "")
		}
		if err != nil {
			return nil, err
		}

		adalToken, err := token.ToADALToken()
		if err != nil {
			return nil, err
		}

		return &adalToken, nil
	}
	spt.SetCustomRefreshFunc(refreshFunc)

	auth := autorest.NewBearerAuthorizer(spt)
	return auth, nil
}

func (a azureCliTokenAuth) getMSALToken(ctx context.Context, _ environments.Api, sender autorest.Sender, oauthConfig *OAuthConfig, endpoint string) (autorest.Authorizer, error) {
	// token version is the decision of az-cli, so we'll pass through to the existing method for continuity
	return a.getADALToken(ctx, sender, oauthConfig, endpoint)
}

func (a azureCliTokenAuth) name() string {
	return "Obtaining a token from the Azure CLI"
}

func (a azureCliTokenAuth) populateConfig(c *Config) error {
	c.ClientID = a.profile.clientId
	c.TenantID = a.profile.tenantId
	c.Environment = a.profile.environment
	c.SubscriptionID = a.profile.subscriptionId

	c.GetAuthenticatedObjectID = func(ctx context.Context) (*string, error) {
		objectId, err := obtainAuthenticatedObjectID(a.profile.azVersion)
		if err != nil {
			return nil, err
		}

		return objectId, nil
	}

	return nil
}

func (a azureCliTokenAuth) validate() error {
	var err *multierror.Error

	errorMessageFmt := "A %s was not found in your Azure CLI Credentials.\n\nPlease login to the Azure CLI again via `az login`"

	if a.profile == nil {
		return fmt.Errorf("Azure CLI Profile is nil - this is an internal error and should be reported.")
	}

	if a.profile.clientId == "" {
		err = multierror.Append(err, fmt.Errorf(errorMessageFmt, "Client ID"))
	}

	if !a.profile.tenantOnly && a.profile.subscriptionId == "" {
		err = multierror.Append(err, fmt.Errorf(errorMessageFmt, "Subscription ID"))
	}

	if a.profile.tenantId == "" {
		err = multierror.Append(err, fmt.Errorf(errorMessageFmt, "Tenant ID"))
	}

	return err.ErrorOrNil()
}

func obtainAuthenticatedObjectID(azVersion version.Version) (*string, error) {
	// Since v2.37.0, CLI migrated the underlying API for `az ad` to using MS graph: https://github.com/Azure/azure-cli/pull/22432.
	// This causes different format for the output of `az ad signed-in-user show -o=json`.
	v2_37 := version.Must(version.NewVersion("2.37.0"))
	args := []string{"ad", "signed-in-user", "show", "-o=json"}
	if azVersion.LessThan(v2_37) {
		var json struct {
			ObjectId string `json:"objectId"`
		}
		if err := jsonUnmarshalAzCmd(&json, args...); err != nil {
			return nil, fmt.Errorf("parsing json result from the Azure CLI: %v", err)
		}
		return &json.ObjectId, nil
	}
	var json struct {
		Id string `json:"id"`
	}
	if err := jsonUnmarshalAzCmd(&json, args...); err != nil {
		return nil, fmt.Errorf("parsing json result from the Azure CLI: %v", err)
	}
	return &json.Id, nil
}

func populateAzVersion(tenantOnly bool) (*version.Version, error) {
	// Azure CLI v2.0.79 is the earliest version to have a `version` command
	var minimumVersion string
	if tenantOnly {
		// v2.0.81 introduced the `--tenant` option to the `account get-access-token` subcommand
		minimumVersion = "2.0.81"
	} else {
		minimumVersion = "2.0.79"
	}

	var cliVersion *struct {
		AzureCli          *string      `json:"azure-cli,omitempty"`
		AzureCliCore      *string      `json:"azure-cli-core,omitempty"`
		AzureCliTelemetry *string      `json:"azure-cli-telemetry,omitempty"`
		Extensions        *interface{} `json:"extensions,omitempty"`
	}
	err := jsonUnmarshalAzCmd(&cliVersion, "version", "-o=json")
	if err != nil {
		return nil, fmt.Errorf("please ensure you have installed Azure CLI version %s or newer. Error parsing json result from the Azure CLI: %v.", minimumVersion, err)
	}

	if cliVersion.AzureCli == nil {
		return nil, fmt.Errorf("could not detect Azure CLI version. Please ensure you have installed Azure CLI version %s or newer.", minimumVersion)
	}

	actual, err := version.NewVersion(*cliVersion.AzureCli)
	if err != nil {
		return nil, fmt.Errorf("could not parse detected Azure CLI version %q: %+v", *cliVersion.AzureCli, err)
	}

	supported := version.Must(version.NewVersion(minimumVersion))

	nextMajor := version.Must(version.NewVersion("3.0.0"))

	if nextMajor.LessThanOrEqual(actual) {
		return nil, fmt.Errorf(`Authenticating using the Azure CLI requires a version older than %[1]s but Terraform detected version %[3]s.

Please install v%[2]s or newer (but also older than %[1]s) and ensure the correct version is in your path.`, nextMajor.String(), supported.String(), actual.String())
	}

	if actual.LessThan(supported) {
		return nil, fmt.Errorf(`Authenticating using the Azure CLI requires version %[1]s but Terraform detected version %[2]s.

Please install v%[1]s or greater and ensure the correct version is in your path.`, supported.String(), actual.String())
	}

	return actual, nil
}

func obtainAuthorizationToken(endpoint string, subscriptionId string, tenantId string) (*cli.Token, error) {
	var token cli.Token
	var err error
	if tenantId != "" {
		err = jsonUnmarshalAzCmd(&token, "account", "get-access-token", "--resource", endpoint, "--tenant", tenantId, "-o=json")
	} else {
		err = jsonUnmarshalAzCmd(&token, "account", "get-access-token", "--resource", endpoint, "--subscription", subscriptionId, "-o=json")
	}
	if err != nil {
		return nil, fmt.Errorf("parsing json result from the Azure CLI: %v", err)
	}

	return &token, nil
}

// obtainSubscription returns a Subscription object of the specified subscriptionId.
// If the subscriptionId is empty, it selects the default subscription.
func obtainSubscription(subscriptionId string) (*cli.Subscription, error) {
	var acc cli.Subscription
	cmd := make([]string, 0)
	cmd = []string{"account", "show", "-o=json"}
	if subscriptionId != "" {
		cmd = append(cmd, "-s", subscriptionId)
	}
	err := jsonUnmarshalAzCmd(&acc, cmd...)
	if err != nil {
		return nil, fmt.Errorf("parsing json result from the Azure CLI: %v", err)
	}

	return &acc, nil
}

// obtainTenant returns a Subscription object having the specified tenantId.
// If the tenantId is empty, it selects the default subscription.
// This works with `az login --allow-no-subscriptions`
func obtainTenant(tenantId string) (*cli.Subscription, error) {
	var acc cli.Subscription
	if tenantId == "" {
		cmd := make([]string, 0)
		cmd = []string{"account", "show", "-o=json"}
		err := jsonUnmarshalAzCmd(&acc, cmd...)
		if err != nil {
			return nil, fmt.Errorf("parsing json result from the Azure CLI: %v", err)
		}
	} else {
		var accs []cli.Subscription
		cmd := make([]string, 0)
		cmd = []string{"account", "list", "-o=json"}
		err := jsonUnmarshalAzCmd(&accs, cmd...)
		if err != nil {
			return nil, fmt.Errorf("parsing json result from the Azure CLI: %v", err)
		}

		for _, a := range accs {
			if a.TenantID == tenantId {
				acc = a
				break
			}
		}

		if acc.TenantID == "" {
			return nil, fmt.Errorf("tenant %q was not found", tenantId)
		}
	}

	return &acc, nil
}

func jsonUnmarshalAzCmd(i interface{}, arg ...string) error {
	var stderr bytes.Buffer
	var stdout bytes.Buffer

	cmd := exec.Command("az", arg...)

	cmd.Stderr = &stderr
	cmd.Stdout = &stdout

	if err := cmd.Start(); err != nil {
		err := fmt.Errorf("launching Azure CLI: %+v", err)
		if stdErrStr := stderr.String(); stdErrStr != "" {
			err = fmt.Errorf("%s: %s", err, strings.TrimSpace(stdErrStr))
		}
		return err
	}

	if err := cmd.Wait(); err != nil {
		err := fmt.Errorf("waiting for the Azure CLI: %+v", err)
		if stdErrStr := stderr.String(); stdErrStr != "" {
			err = fmt.Errorf("%s: %s", err, strings.TrimSpace(stdErrStr))
		}
		return err
	}

	if err := json.Unmarshal([]byte(stdout.String()), &i); err != nil {
		return fmt.Errorf("unmarshaling the result of Azure CLI: %v", err)
	}

	return nil
}
