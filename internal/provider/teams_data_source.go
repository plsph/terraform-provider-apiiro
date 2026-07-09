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
	_ datasource.DataSource              = &teamsDataSource{}
	_ datasource.DataSourceWithConfigure = &teamsDataSource{}
)

type teamsDataSource struct {
	client *Client
}

type teamsDataSourceModel struct {
	ID    types.String `tfsdk:"id"`
	Name  types.String `tfsdk:"name"`
	Key   types.String `tfsdk:"key"`
	Teams types.List   `tfsdk:"teams"`
}

func NewTeamsDataSource() datasource.DataSource {
	return &teamsDataSource{}
}

func (d *teamsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_teams"
}

func (d *teamsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches organizational teams from Apiiro with optional filtering by name or key.",
		Attributes: map[string]schema.Attribute{
			"id":   schema.StringAttribute{Computed: true},
			"name": schema.StringAttribute{Optional: true},
			"key":  schema.StringAttribute{Optional: true},
			"teams": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
					"id":               schema.StringAttribute{Computed: true},
					"key":              schema.StringAttribute{Computed: true},
					"name":             schema.StringAttribute{Computed: true},
					"description":      schema.StringAttribute{Computed: true},
					"parent_key":       schema.StringAttribute{Computed: true},
					"applications":     schema.ListAttribute{Computed: true, ElementType: types.StringType},
					"project_urls":     schema.ListAttribute{Computed: true, ElementType: types.StringType},
					"repository_urls":  schema.ListAttribute{Computed: true, ElementType: types.StringType},
					"application_tags": schema.MapAttribute{Computed: true, ElementType: types.StringType},
					"tags":             schema.MapAttribute{Computed: true, ElementType: types.StringType},
				}},
			},
		},
	}
}

func (d *teamsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *teamsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state teamsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	nameFilter := strings.TrimSpace(state.Name.ValueString())
	keyFilter := strings.TrimSpace(state.Key.ValueString())

	teams, err := d.client.listTeams()
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Teams", err.Error())
		return
	}

	values := make([]attr.Value, 0, len(teams))
	for _, team := range teams {
		if nameFilter != "" && !strings.Contains(strings.ToLower(valueOrEmpty(team.Name)), strings.ToLower(nameFilter)) {
			continue
		}
		if keyFilter != "" && !strings.EqualFold(valueOrEmpty(team.Key), keyFilter) {
			continue
		}

		obj, diags := types.ObjectValue(teamsDataSourceObjectType(), map[string]attr.Value{
			"id":               nullableString(team.Key),
			"key":              nullableString(team.Key),
			"name":             nullableString(team.Name),
			"description":      nullableString(team.Description),
			"parent_key":       nullableString(team.ParentKey),
			"applications":     stringSliceToList(team.Applications),
			"project_urls":     stringSliceToList(team.ProjectUrls),
			"repository_urls":  stringSliceToList(team.RepositoryUrls),
			"application_tags": mapToTerraform(tagsBodiesToMap(team.ApplicationTags)),
			"tags":             mapToTerraform(tagsBodiesToMap(team.Tags)),
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		values = append(values, obj)
	}

	teamList, diags := types.ListValue(types.ObjectType{AttrTypes: teamsDataSourceObjectType()}, values)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.ID = types.StringValue(fmt.Sprintf("teams:%s:%s", nameFilter, keyFilter))
	state.Teams = teamList
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func teamsDataSourceObjectType() map[string]attr.Type {
	return map[string]attr.Type{
		"id":               types.StringType,
		"key":              types.StringType,
		"name":             types.StringType,
		"description":      types.StringType,
		"parent_key":       types.StringType,
		"applications":     types.ListType{ElemType: types.StringType},
		"project_urls":     types.ListType{ElemType: types.StringType},
		"repository_urls":  types.ListType{ElemType: types.StringType},
		"application_tags": types.MapType{ElemType: types.StringType},
		"tags":             types.MapType{ElemType: types.StringType},
	}
}
