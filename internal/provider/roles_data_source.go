package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &rolesDataSource{}
	_ datasource.DataSourceWithConfigure = &rolesDataSource{}
)

type rolesDataSource struct {
	client *Client
}

type rolesDataSourceModel struct {
	ID    types.String `tfsdk:"id"`
	Name  types.String `tfsdk:"name"`
	Key   types.String `tfsdk:"key"`
	Roles types.List   `tfsdk:"roles"`
}

func NewRolesDataSource() datasource.DataSource {
	return &rolesDataSource{}
}

func (d *rolesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_roles"
}

func (d *rolesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches roles from Apiiro with optional filtering by name or key.",
		Attributes: map[string]schema.Attribute{
			"id":   schema.StringAttribute{Computed: true},
			"name": schema.StringAttribute{Optional: true},
			"key":  schema.StringAttribute{Optional: true},
			"roles": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
					"id":   schema.StringAttribute{Computed: true},
					"key":  schema.StringAttribute{Computed: true},
					"name": schema.StringAttribute{Computed: true},
				}},
			},
		},
	}
}

func (d *rolesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *Client, got: %T", req.ProviderData))
		return
	}
	d.client = client
}

func (d *rolesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state rolesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	nameFilter := strings.TrimSpace(state.Name.ValueString())
	keyFilter := strings.TrimSpace(state.Key.ValueString())
	tflog.Debug(ctx, "roles data source read requested", map[string]any{"name": nameFilter, "key": keyFilter})

	roles, err := d.client.listRoles(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Roles", err.Error())
		return
	}

	values := make([]attr.Value, 0, len(roles))
	for _, role := range roles {
		if nameFilter != "" && !strings.Contains(strings.ToLower(valueOrEmpty(role.Name)), strings.ToLower(nameFilter)) {
			continue
		}
		if keyFilter != "" && !strings.EqualFold(valueOrEmpty(role.Key), keyFilter) {
			continue
		}

		obj, diags := types.ObjectValue(map[string]attr.Type{"id": types.StringType, "key": types.StringType, "name": types.StringType}, map[string]attr.Value{
			"id":   nullableString(role.Key),
			"key":  nullableString(role.Key),
			"name": nullableString(role.Name),
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		values = append(values, obj)
	}

	roleList, diags := types.ListValue(types.ObjectType{AttrTypes: map[string]attr.Type{"id": types.StringType, "key": types.StringType, "name": types.StringType}}, values)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.ID = types.StringValue(fmt.Sprintf("roles:%s:%s", nameFilter, keyFilter))
	state.Roles = roleList
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
