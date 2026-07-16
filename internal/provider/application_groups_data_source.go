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
	_ datasource.DataSource              = &applicationGroupsDataSource{}
	_ datasource.DataSourceWithConfigure = &applicationGroupsDataSource{}
)

type applicationGroupsDataSource struct {
	client *Client
}

type applicationGroupsDataSourceModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Key               types.String `tfsdk:"key"`
	ApplicationGroups types.List   `tfsdk:"application_groups"`
}

func NewApplicationGroupsDataSource() datasource.DataSource {
	return &applicationGroupsDataSource{}
}

func (d *applicationGroupsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_applicationgroups"
}

func (d *applicationGroupsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches application groups from Apiiro with optional filtering by name or key.",
		Attributes: map[string]schema.Attribute{
			"id":   schema.StringAttribute{Computed: true},
			"name": schema.StringAttribute{Optional: true},
			"key":  schema.StringAttribute{Optional: true},
			"application_groups": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
					"id":           schema.StringAttribute{Computed: true},
					"key":          schema.StringAttribute{Computed: true},
					"name":         schema.StringAttribute{Computed: true},
					"applications": schema.ListAttribute{Computed: true, ElementType: types.StringType},
					"points_of_contact": schema.ListNestedAttribute{Computed: true, NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
						"identity": schema.StringAttribute{Computed: true},
						"title":    schema.StringAttribute{Computed: true},
					}}},
					"tags": schema.ListAttribute{Computed: true, ElementType: types.StringType},
				}},
			},
		},
	}
}

func (d *applicationGroupsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *applicationGroupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state applicationGroupsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	nameFilter := strings.TrimSpace(state.Name.ValueString())
	keyFilter := strings.TrimSpace(state.Key.ValueString())
	tflog.Debug(ctx, "application groups data source read requested", map[string]any{"name": nameFilter, "key": keyFilter})

	apiFilters := map[string][]string{}
	if nameFilter != "" {
		apiFilters["Name"] = []string{nameFilter}
	}

	groups, err := d.client.listApplicationGroups(ctx, apiFilters)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Application Groups", err.Error())
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

		obj, diags := types.ObjectValue(applicationGroupsDataSourceObjectType(), map[string]attr.Value{
			"id":                nullableString(group.Key),
			"key":               nullableString(group.Key),
			"name":              nullableString(group.Name),
			"applications":      stringSliceToList(group.Applications),
			"points_of_contact": pointOfContactBodiesToList(group.PointsOfContact),
			"tags":              stringSliceToList(group.Tags),
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		values = append(values, obj)
	}

	groupList, diags := types.ListValue(types.ObjectType{AttrTypes: applicationGroupsDataSourceObjectType()}, values)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.ID = types.StringValue(fmt.Sprintf("applicationgroups:%s:%s", nameFilter, keyFilter))
	state.ApplicationGroups = groupList
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func applicationGroupsDataSourceObjectType() map[string]attr.Type {
	return map[string]attr.Type{
		"id":                types.StringType,
		"key":               types.StringType,
		"name":              types.StringType,
		"applications":      types.ListType{ElemType: types.StringType},
		"points_of_contact": types.ListType{ElemType: types.ObjectType{AttrTypes: pointOfContactObjectType()}},
		"tags":              types.ListType{ElemType: types.StringType},
	}
}
