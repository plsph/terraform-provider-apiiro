package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &applicationsDataSource{}
	_ datasource.DataSourceWithConfigure = &applicationsDataSource{}
)

type applicationsDataSource struct {
	client *Client
}

type applicationsDataSourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	ApplicationType types.String `tfsdk:"application_type"`
	ApplicationKey  types.String `tfsdk:"application_key"`
	Applications    types.List   `tfsdk:"applications"`
}

func NewApplicationsDataSource() datasource.DataSource {
	return &applicationsDataSource{}
}

func (d *applicationsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_applications"
}

func (d *applicationsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches applications from Apiiro with optional filtering by name, key, or application type.",
		Attributes: map[string]schema.Attribute{
			"id":               schema.StringAttribute{Computed: true},
			"name":             schema.StringAttribute{Optional: true},
			"application_type": schema.StringAttribute{Optional: true},
			"application_key":  schema.StringAttribute{Optional: true},
			"applications": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
					"id":                     schema.StringAttribute{Computed: true},
					"key":                    schema.StringAttribute{Computed: true},
					"name":                   schema.StringAttribute{Computed: true},
					"application_type":       schema.StringAttribute{Computed: true},
					"business_impact":        schema.StringAttribute{Computed: true},
					"business_unit":          schema.StringAttribute{Computed: true},
					"description":            schema.StringAttribute{Computed: true},
					"estimated_revenue":      schema.StringAttribute{Computed: true},
					"estimated_users_number": schema.StringAttribute{Computed: true},
					"is_internet_facing":     schema.BoolAttribute{Computed: true},
					"api_gateway_urls":       schema.ListAttribute{Computed: true, ElementType: types.StringType},
					"api_group_keys":         schema.ListAttribute{Computed: true, ElementType: types.StringType},
					"project_urls":           schema.ListAttribute{Computed: true, ElementType: types.StringType},
					"repository_urls":        schema.ListAttribute{Computed: true, ElementType: types.StringType},
					"tags":                   schema.MapAttribute{Computed: true, ElementType: types.StringType},
				}},
			},
		},
	}
}

func (d *applicationsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *applicationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state applicationsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	nameFilter := strings.TrimSpace(state.Name.ValueString())
	typeFilter := strings.TrimSpace(state.ApplicationType.ValueString())
	keyFilter := strings.TrimSpace(state.ApplicationKey.ValueString())

	applications, err := d.client.listApplications()
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Applications", err.Error())
		return
	}

	values := make([]attr.Value, 0, len(applications))
	for _, app := range applications {
		if nameFilter != "" && !strings.Contains(strings.ToLower(valueOrEmpty(app.Name)), strings.ToLower(nameFilter)) {
			continue
		}
		if typeFilter != "" && !strings.EqualFold(valueOrEmpty(app.ApplicationType), typeFilter) {
			continue
		}
		if keyFilter != "" && !strings.EqualFold(valueOrEmpty(app.Key), keyFilter) {
			continue
		}

		obj, diags := types.ObjectValue(applicationsDataSourceObjectType(), map[string]attr.Value{
			"id":                     nullableString(app.Key),
			"key":                    nullableString(app.Key),
			"name":                   nullableString(app.Name),
			"application_type":       nullableString(app.ApplicationType),
			"business_impact":        nullableString(app.BusinessImpact),
			"business_unit":          nullableString(app.BusinessUnit),
			"description":            nullableString(app.Description),
			"estimated_revenue":      nullableString(app.EstimatedRevenue),
			"estimated_users_number": nullableString(app.EstimatedUsersNumber),
			"is_internet_facing":     types.BoolValue(app.IsInternetFacing),
			"api_gateway_urls":       stringSliceToList(app.ApiGatewayUrls),
			"api_group_keys":         stringSliceToList(app.ApiGroupKeys),
			"project_urls":           stringSliceToList(app.ProjectUrls),
			"repository_urls":        stringSliceToList(app.RepositoryUrls),
			"tags":                   mapToTerraform(tagsBodiesToMap(app.Tags)),
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		values = append(values, obj)
	}

	appList, diags := types.ListValue(types.ObjectType{AttrTypes: applicationsDataSourceObjectType()}, values)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.ID = types.StringValue(fmt.Sprintf("applications:%s:%s:%s", nameFilter, typeFilter, keyFilter))
	state.Applications = appList
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func applicationsDataSourceObjectType() map[string]attr.Type {
	return map[string]attr.Type{
		"id":                     types.StringType,
		"key":                    types.StringType,
		"name":                   types.StringType,
		"application_type":       types.StringType,
		"business_impact":        types.StringType,
		"business_unit":          types.StringType,
		"description":            types.StringType,
		"estimated_revenue":      types.StringType,
		"estimated_users_number": types.StringType,
		"is_internet_facing":     types.BoolType,
		"api_gateway_urls":       types.ListType{ElemType: types.StringType},
		"api_group_keys":         types.ListType{ElemType: types.StringType},
		"project_urls":           types.ListType{ElemType: types.StringType},
		"repository_urls":        types.ListType{ElemType: types.StringType},
		"tags":                   types.MapType{ElemType: types.StringType},
	}
}
