package servicebus

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/preview/servicebus/mgmt/2021-06-01-preview/servicebus"
	"github.com/hashicorp/go-azure-helpers/lang/response"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/internal/clients"
	"github.com/hashicorp/terraform-provider-azurerm/internal/identity"
	"github.com/hashicorp/terraform-provider-azurerm/internal/location"
	msiParse "github.com/hashicorp/terraform-provider-azurerm/internal/services/msi/parse"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/servicebus/migration"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/servicebus/parse"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/servicebus/validate"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tags"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/suppress"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/internal/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

// Default Authorization Rule/Policy created by Azure, used to populate the
// default connection strings and keys
var (
	serviceBusNamespaceDefaultAuthorizationRule = "RootManageSharedAccessKey"
	serviceBusNamespaceResourceName             = "azurerm_servicebus_namespace"
)

func resourceServiceBusNamespace() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceServiceBusNamespaceCreateUpdate,
		Read:   resourceServiceBusNamespaceRead,
		Update: resourceServiceBusNamespaceCreateUpdate,
		Delete: resourceServiceBusNamespaceDelete,

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.NamespaceID(id)
			return err
		}),

		SchemaVersion: 1,
		StateUpgraders: pluginsdk.StateUpgrades(map[int]pluginsdk.StateUpgrade{
			0: migration.NamespaceV0ToV1{},
		}),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.NamespaceName,
			},

			"location": azure.SchemaLocation(),

			"resource_group_name": azure.SchemaResourceGroupName(),

			"identity": identity.SystemAssignedUserAssigned{}.Schema(),

			"sku": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(servicebus.SkuNameBasic),
					string(servicebus.SkuNameStandard),
					string(servicebus.SkuNamePremium),
				}, true),
				DiffSuppressFunc: suppress.CaseDifference,
			},

			"capacity": {
				Type:         pluginsdk.TypeInt,
				Optional:     true,
				Default:      0,
				ValidateFunc: validation.IntInSlice([]int{0, 1, 2, 4, 8, 16}),
			},

			"default_primary_connection_string": {
				Type:      pluginsdk.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"default_secondary_connection_string": {
				Type:      pluginsdk.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"default_primary_key": {
				Type:      pluginsdk.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"default_secondary_key": {
				Type:      pluginsdk.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"zone_redundant": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				ForceNew: true,
			},

			"tags": tags.Schema(),
		},
	}
}

func resourceServiceBusNamespaceCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ServiceBus.NamespacesClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	log.Printf("[INFO] preparing arguments for ServiceBus Namespace create/update.")

	location := azure.NormalizeLocation(d.Get("location").(string))
	sku := d.Get("sku").(string)
	t := d.Get("tags").(map[string]interface{})

	resourceId := parse.NewNamespaceID(subscriptionId, d.Get("resource_group_name").(string), d.Get("name").(string))
	if d.IsNewResource() {
		existing, err := client.Get(ctx, resourceId.ResourceGroup, resourceId.Name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing %s: %+v", resourceId, err)
			}
		}

		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurerm_servicebus_namespace", resourceId.ID())
		}
	}

	parameters := servicebus.SBNamespace{
		Location: &location,
		Sku: &servicebus.SBSku{
			Name: servicebus.SkuName(sku),
			Tier: servicebus.SkuTier(sku),
		},
		SBNamespaceProperties: &servicebus.SBNamespaceProperties{
			ZoneRedundant: utils.Bool(d.Get("zone_redundant").(bool)),
		},
		Tags: tags.Expand(t),
	}

	identity, err := expandServiceBusNamespaceIdentity(d.Get("identity").([]interface{}))
	if err != nil {
		return fmt.Errorf("expanding `identity`: %+v", err)
	}
	parameters.Identity = identity

	if capacity := d.Get("capacity"); capacity != nil {
		if !strings.EqualFold(sku, string(servicebus.SkuNamePremium)) && capacity.(int) > 0 {
			return fmt.Errorf("Service Bus SKU %q only supports `capacity` of 0", sku)
		}
		if strings.EqualFold(sku, string(servicebus.SkuNamePremium)) && capacity.(int) == 0 {
			return fmt.Errorf("Service Bus SKU %q only supports `capacity` of 1, 2, 4, 8 or 16", sku)
		}
		parameters.Sku.Capacity = utils.Int32(int32(capacity.(int)))
	}

	future, err := client.CreateOrUpdate(ctx, resourceId.ResourceGroup, resourceId.Name, parameters)
	if err != nil {
		return fmt.Errorf("creating/updating %s: %+v", resourceId, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for create/update of %s: %+v", resourceId, err)
	}

	d.SetId(resourceId.ID())
	return resourceServiceBusNamespaceRead(d, meta)
}

func resourceServiceBusNamespaceRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ServiceBus.NamespacesClient
	clientStable := meta.(*clients.Client).ServiceBus.NamespacesClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.NamespaceID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving %s: %+v", id, err)
	}

	d.Set("name", id.Name)
	d.Set("resource_group_name", id.ResourceGroup)
	d.Set("location", location.NormalizeNilable(resp.Location))

	identity, err := flattenServiceBusNamespaceIdentity(resp.Identity)
	if err != nil {
		return err
	}
	d.Set("identity", identity)

	if sku := resp.Sku; sku != nil {
		d.Set("sku", strings.ToLower(string(sku.Name)))
		d.Set("capacity", sku.Capacity)
	}

	if properties := resp.SBNamespaceProperties; properties != nil {
		d.Set("zone_redundant", properties.ZoneRedundant)
	}

	keys, err := clientStable.ListKeys(ctx, id.ResourceGroup, id.Name, serviceBusNamespaceDefaultAuthorizationRule)
	if err != nil {
		log.Printf("[WARN] listing default keys for %s: %+v", id, err)
	} else {
		d.Set("default_primary_connection_string", keys.PrimaryConnectionString)
		d.Set("default_secondary_connection_string", keys.SecondaryConnectionString)
		d.Set("default_primary_key", keys.PrimaryKey)
		d.Set("default_secondary_key", keys.SecondaryKey)
	}

	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceServiceBusNamespaceDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ServiceBus.NamespacesClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.NamespaceID(d.Id())
	if err != nil {
		return err
	}

	future, err := client.Delete(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		return fmt.Errorf("deleting %s: %+v", id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		if !response.WasNotFound(future.Response()) {
			return fmt.Errorf("waiting for deletion of %s: %+v", id, err)
		}
	}

	return nil
}

func expandServiceBusNamespaceIdentity(input []interface{}) (*servicebus.Identity, error) {
	if len(input) == 0 || input[0] == nil {
		return &servicebus.Identity{
			Type: servicebus.ManagedServiceIdentityTypeNone,
		}, nil
	}

	v := input[0].(map[string]interface{})

	config := &servicebus.Identity{
		Type: servicebus.ManagedServiceIdentityType(v["type"].(string)),
	}

	identityIds := v["identity_ids"].(*pluginsdk.Set).List()

	if len(identityIds) != 0 {
		if config.Type != servicebus.ManagedServiceIdentityTypeSystemAssignedUserAssigned && config.Type != servicebus.ManagedServiceIdentityTypeUserAssigned {
			return nil, fmt.Errorf("`identity_ids` can only be specified when `type` includes `UserAssigned`")
		}
		config.UserAssignedIdentities = map[string]*servicebus.UserAssignedIdentity{}
		for _, id := range identityIds {
			config.UserAssignedIdentities[id.(string)] = &servicebus.UserAssignedIdentity{}
		}
	}

	return config, nil
}

func flattenServiceBusNamespaceIdentity(input *servicebus.Identity) ([]interface{}, error) {
	if input == nil || input.Type == servicebus.ManagedServiceIdentityTypeNone {
		return []interface{}{}, nil
	}

	coalesce := func(input *string) string {
		if input == nil {
			return ""
		}

		return *input
	}

	var identityIds []string
	for id := range input.UserAssignedIdentities {
		parsedId, err := msiParse.UserAssignedIdentityIDInsensitively(id)
		if err != nil {
			return nil, err
		}
		identityIds = append(identityIds, parsedId.ID())
	}

	return []interface{}{
		map[string]interface{}{
			"type":         input.Type,
			"identity_ids": identityIds,
			"principal_id": coalesce(input.PrincipalID),
			"tenant_id":    coalesce(input.TenantID),
		},
	}, nil
}
