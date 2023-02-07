package keyvault

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-provider-azurerm/internal/clients"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/keyvault/parse"
	keyVaultValidate "github.com/hashicorp/terraform-provider-azurerm/internal/services/keyvault/validate"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func dataSourceKeyVaultCertificates() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Read: dataSourceKeyVaultCertificatesRead,

		Timeouts: &pluginsdk.ResourceTimeout{
			Read: pluginsdk.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"key_vault_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: keyVaultValidate.VaultID,
			},

			"names": {
				Type:     pluginsdk.TypeList,
				Computed: true,
				Elem: &pluginsdk.Schema{
					Type: pluginsdk.TypeString,
				},
			},

			"include_pending": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				Default:  true,
			},
		},
	}
}

func dataSourceKeyVaultCertificatesRead(d *pluginsdk.ResourceData, meta interface{}) error {
	keyVaultsClient := meta.(*clients.Client).KeyVault
	client := meta.(*clients.Client).KeyVault.ManagementClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	keyVaultId, err := parse.VaultID(d.Get("key_vault_id").(string))
	if err != nil {
		return err
	}

	includePending := d.Get("include_pending").(bool)

	keyVaultBaseUri, err := keyVaultsClient.BaseUriForKeyVault(ctx, *keyVaultId)
	if err != nil {
		return fmt.Errorf("fetching base vault url from id %q: %+v", *keyVaultId, err)
	}

	certificateList, err := client.GetCertificatesComplete(ctx, *keyVaultBaseUri, utils.Int32(25), &includePending)
	if err != nil {
		return fmt.Errorf("retrieving %s: %+v", *keyVaultId, err)
	}

	d.SetId(keyVaultId.ID())

	var names []string
	if certificateList.Response().Value != nil {
		for certificateList.NotDone() {
			for _, v := range *certificateList.Response().Value {
				nestedItem, err := parse.ParseOptionallyVersionedNestedItemID(*v.ID)
				if err != nil {
					return err
				}
				names = append(names, nestedItem.Name)
				err = certificateList.NextWithContext(ctx)
				if err != nil {
					return fmt.Errorf("retrieving next page of Certificates from %s: %+v", *keyVaultId, err)
				}
			}
		}
	}

	d.Set("names", names)
	d.Set("key_vault_id", keyVaultId.ID())

	return nil
}
