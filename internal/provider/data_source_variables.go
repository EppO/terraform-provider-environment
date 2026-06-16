package environment

import (
	"context"
	"os"

	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceVariables() *schema.Resource {
	return &schema.Resource{
		DeprecationMessage: "The environment_variables data source is deprecated and " +
			"will be removed in v3.0. Use the get_env or get_env_map provider-defined " +
			"functions instead. See " +
			"https://registry.terraform.io/providers/EppO/environment/latest/docs/functions/get_env_map " +
			"for the migration guide.",
		ReadContext: dataSourceVariablesRead,
		Schema: map[string]*schema.Schema{
			"sensitive": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"filter": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"items": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceVariablesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sensitive := d.Get("sensitive").(bool)
	filter := d.Get("filter").(string)

	items, err := filterVariables(os.Environ(), sensitive, filter)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("items", items); err != nil {
		return diag.FromErr(err)
	}

	uuid, err := uuid.GenerateUUID()
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(uuid)

	return nil
}
