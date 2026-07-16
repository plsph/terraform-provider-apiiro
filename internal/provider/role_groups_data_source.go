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
	_ datasource.DataSource              = &roleGroupsDataSource{}
	_ datasource.DataSourceWithConfigure = &roleGroupsDataSource{}
)

type roleGroupsDataSource struct {
	client *Client
}

type roleGroupsDataSourceModel struct {
	ID     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Key    types.String `tfsdk:"key"`
	Groups types.List   `tfsdk:"groups"`
}

func NewRoleGroupsDataSource() datasource.DataSource {
	return &roleGroupsDataSource{}
}

func (d *roleGroupsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rolegroups"
}

func (d *roleGroupsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches Apiiro user groups with optional filtering by name or key.",
		Attributes: map[string]schema.Attribute{
			"id":   schema.StringAttribute{Computed: true},
			"name": schema.StringAttribute{Optional: true},
			"key":  schema.StringAttribute{Optional: true},
			"groups": schema.ListNestedAttribute{Computed: true, NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
				"id":          schema.StringAttribute{Computed: true},
				"key":         schema.StringAttribute{Computed: true},
				"name":        schema.StringAttribute{Computed: true},
				"description": schema.StringAttribute{Computed: true},
			}}},
		},
	}
}

func (d *roleGroupsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *roleGroupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state roleGroupsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	nameFilter := strings.TrimSpace(state.Name.ValueString())
	keyFilter := strings.TrimSpace(state.Key.ValueString())
	tflog.Debug(ctx, "role groups data source read requested", map[string]any{"name": nameFilter, "key": keyFilter})

	groups, err := d.client.listRoleGroups(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Groups", err.Error())
		return
	}

	values := make([]attr.Value, 0, len(groups))
	for _, group := range groups {
		if nameFilter != "" && !strings.Contains(strings.ToLower(valueOrEmpty(group.Name)), strings.ToLower(nameFilter)) {
			continue
		}
		if keyFilter != "" && !strings.EqualFold(valueOrEmpty(group.Key), keyFilter) {
			continue
		}
		obj, diags := types.ObjectValue(map[string]attr.Type{"id": types.StringType, "key": types.StringType, "name": types.StringType, "description": types.StringType}, map[string]attr.Value{
			"id":          nullableString(group.Key),
			"key":         nullableString(group.Key),
			"name":        nullableString(group.Name),
			"description": nullableString(group.Description),
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		values = append(values, obj)
	}

	groupList, diags := types.ListValue(types.ObjectType{AttrTypes: map[string]attr.Type{"id": types.StringType, "key": types.StringType, "name": types.StringType, "description": types.StringType}}, values)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.ID = types.StringValue(fmt.Sprintf("rolegroups:%s:%s", nameFilter, keyFilter))
	state.Groups = groupList
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
