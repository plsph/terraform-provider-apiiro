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
	_ datasource.DataSource              = &repositoriesDataSource{}
	_ datasource.DataSourceWithConfigure = &repositoriesDataSource{}
)

type repositoriesDataSource struct {
	client *Client
}

type repositoriesDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Key          types.String `tfsdk:"key"`
	ScmKey       types.String `tfsdk:"scm_repository_key"`
	ProviderName types.String `tfsdk:"provider_name"`
	Repositories types.List   `tfsdk:"repositories"`
}

func NewRepositoriesDataSource() datasource.DataSource {
	return &repositoriesDataSource{}
}

func (d *repositoriesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repositories"
}

func (d *repositoriesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches repositories from Apiiro v2 with optional filtering by name, key, SCM repository key, or provider.",
		Attributes: map[string]schema.Attribute{
			"id":                 schema.StringAttribute{Computed: true},
			"name":               schema.StringAttribute{Optional: true},
			"key":                schema.StringAttribute{Optional: true},
			"scm_repository_key": schema.StringAttribute{Optional: true},
			"provider_name":      schema.StringAttribute{Optional: true},
			"repositories": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
					"active_since":               schema.StringAttribute{Computed: true},
					"api_count":                  schema.Int64Attribute{Computed: true},
					"branch_name":                schema.StringAttribute{Computed: true},
					"business_impact":            schema.StringAttribute{Computed: true},
					"contributor_count":          schema.Int64Attribute{Computed: true},
					"dependency_count":           schema.Int64Attribute{Computed: true},
					"external_id":                schema.StringAttribute{Computed: true},
					"has_data_models":            schema.BoolAttribute{Computed: true},
					"has_external_dependencies":  schema.BoolAttribute{Computed: true},
					"has_payments_data":          schema.BoolAttribute{Computed: true},
					"has_phi_data":               schema.BoolAttribute{Computed: true},
					"has_pii_data":               schema.BoolAttribute{Computed: true},
					"has_sensitive_dependencies": schema.BoolAttribute{Computed: true},
					"insights":                   schema.ListAttribute{Computed: true, ElementType: types.StringType},
					"is_active":                  schema.BoolAttribute{Computed: true},
					"is_archived":                schema.BoolAttribute{Computed: true},
					"is_default_branch":          schema.BoolAttribute{Computed: true},
					"is_deployed":                schema.BoolAttribute{Computed: true},
					"is_internet_exposed":        schema.BoolAttribute{Computed: true},
					"is_public":                  schema.BoolAttribute{Computed: true},
					"key":                        schema.StringAttribute{Computed: true},
					"language_percentages":       schema.MapAttribute{Computed: true, ElementType: types.Float64Type},
					"languages":                  schema.ListAttribute{Computed: true, ElementType: types.StringType},
					"last_activity":              schema.StringAttribute{Computed: true},
					"licenses":                   schema.ListAttribute{Computed: true, ElementType: types.StringType},
					"modules": schema.ListNestedAttribute{Computed: true, NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
						"file_path": schema.StringAttribute{Computed: true},
						"name":      schema.StringAttribute{Computed: true},
					}}},
					"name":               schema.StringAttribute{Computed: true},
					"project_id":         schema.StringAttribute{Computed: true},
					"provider":           schema.StringAttribute{Computed: true},
					"risk_level":         schema.StringAttribute{Computed: true},
					"risk_score":         schema.Int64Attribute{Computed: true},
					"scm_repository_key": schema.StringAttribute{Computed: true},
					"server_url":         schema.StringAttribute{Computed: true},
					"url":                schema.StringAttribute{Computed: true},
				}},
			},
		},
	}
}

func (d *repositoriesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *repositoriesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state repositoriesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	nameFilter := strings.TrimSpace(state.Name.ValueString())
	keyFilter := strings.TrimSpace(state.Key.ValueString())
	scmKeyFilter := strings.TrimSpace(state.ScmKey.ValueString())
	providerFilter := strings.TrimSpace(state.ProviderName.ValueString())
	tflog.Debug(ctx, "repositories data source read requested", map[string]any{"name": nameFilter, "key": keyFilter, "scm_repository_key": scmKeyFilter, "provider_name": providerFilter})

	repositories, err := d.client.listRepositoriesV2(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Repositories", err.Error())
		return
	}

	values := make([]attr.Value, 0, len(repositories))
	for _, repository := range repositories {
		if nameFilter != "" && !strings.Contains(strings.ToLower(valueOrEmpty(repository.Name)), strings.ToLower(nameFilter)) {
			continue
		}
		if keyFilter != "" && !strings.EqualFold(valueOrEmpty(repository.Key), keyFilter) {
			continue
		}
		if scmKeyFilter != "" && !strings.EqualFold(valueOrEmpty(repository.ScmRepositoryKey), scmKeyFilter) {
			continue
		}
		if providerFilter != "" && !strings.EqualFold(valueOrEmpty(repository.Provider), providerFilter) {
			continue
		}

		obj, diags := types.ObjectValue(repositoriesDataSourceObjectType(), map[string]attr.Value{
			"active_since":               nullableString(repository.ActiveSince),
			"api_count":                  nullableInt64(repository.ApiCount),
			"branch_name":                nullableString(repository.BranchName),
			"business_impact":            nullableString(repository.BusinessImpact),
			"contributor_count":          nullableInt64(repository.ContributorCount),
			"dependency_count":           nullableInt64(repository.DependencyCount),
			"external_id":                nullableString(repository.ExternalID),
			"has_data_models":            nullableBool(repository.HasDataModels),
			"has_external_dependencies":  nullableBool(repository.HasExternalDependencies),
			"has_payments_data":          nullableBool(repository.HasPaymentsData),
			"has_phi_data":               nullableBool(repository.HasPhiData),
			"has_pii_data":               nullableBool(repository.HasPiiData),
			"has_sensitive_dependencies": nullableBool(repository.HasSensitiveDependencies),
			"insights":                   stringSliceToList(repository.Insights),
			"is_active":                  nullableBool(repository.IsActive),
			"is_archived":                types.BoolValue(repository.IsArchived),
			"is_default_branch":          types.BoolValue(repository.IsDefaultBranch),
			"is_deployed":                nullableBool(repository.IsDeployed),
			"is_internet_exposed":        nullableBool(repository.IsInternetExposed),
			"is_public":                  types.BoolValue(repository.IsPublic),
			"key":                        nullableString(repository.Key),
			"language_percentages":       floatMapToTerraform(repository.LanguagePercentages),
			"languages":                  stringSliceToList(repository.Languages),
			"last_activity":              nullableString(repository.LastActivity),
			"licenses":                   stringSliceToList(repository.Licenses),
			"modules":                    moduleBodiesToList(repository.Modules),
			"name":                       nullableString(repository.Name),
			"project_id":                 nullableString(repository.ProjectID),
			"provider":                   nullableString(repository.Provider),
			"risk_level":                 nullableString(repository.RiskLevel),
			"risk_score":                 types.Int64Value(int64(repository.RiskScore)),
			"scm_repository_key":         nullableString(repository.ScmRepositoryKey),
			"server_url":                 nullableString(repository.ServerURL),
			"url":                        nullableString(repository.URL),
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		values = append(values, obj)
	}

	repoList, diags := types.ListValue(types.ObjectType{AttrTypes: repositoriesDataSourceObjectType()}, values)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.ID = types.StringValue(fmt.Sprintf("repositories:%s:%s:%s:%s", nameFilter, keyFilter, scmKeyFilter, providerFilter))
	state.Repositories = repoList
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func repositoriesDataSourceObjectType() map[string]attr.Type {
	return map[string]attr.Type{
		"active_since":               types.StringType,
		"api_count":                  types.Int64Type,
		"branch_name":                types.StringType,
		"business_impact":            types.StringType,
		"contributor_count":          types.Int64Type,
		"dependency_count":           types.Int64Type,
		"external_id":                types.StringType,
		"has_data_models":            types.BoolType,
		"has_external_dependencies":  types.BoolType,
		"has_payments_data":          types.BoolType,
		"has_phi_data":               types.BoolType,
		"has_pii_data":               types.BoolType,
		"has_sensitive_dependencies": types.BoolType,
		"insights":                   types.ListType{ElemType: types.StringType},
		"is_active":                  types.BoolType,
		"is_archived":                types.BoolType,
		"is_default_branch":          types.BoolType,
		"is_deployed":                types.BoolType,
		"is_internet_exposed":        types.BoolType,
		"is_public":                  types.BoolType,
		"key":                        types.StringType,
		"language_percentages":       types.MapType{ElemType: types.Float64Type},
		"languages":                  types.ListType{ElemType: types.StringType},
		"last_activity":              types.StringType,
		"licenses":                   types.ListType{ElemType: types.StringType},
		"modules":                    types.ListType{ElemType: types.ObjectType{AttrTypes: repositoryModuleObjectType()}},
		"name":                       types.StringType,
		"project_id":                 types.StringType,
		"provider":                   types.StringType,
		"risk_level":                 types.StringType,
		"risk_score":                 types.Int64Type,
		"scm_repository_key":         types.StringType,
		"server_url":                 types.StringType,
		"url":                        types.StringType,
	}
}

func repositoryModuleObjectType() map[string]attr.Type {
	return map[string]attr.Type{
		"file_path": types.StringType,
		"name":      types.StringType,
	}
}

func moduleBodiesToList(values []moduleBody) types.List {
	if len(values) == 0 {
		return types.ListNull(types.ObjectType{AttrTypes: repositoryModuleObjectType()})
	}
	elements := make([]attr.Value, 0, len(values))
	for _, value := range values {
		obj, diags := types.ObjectValue(repositoryModuleObjectType(), map[string]attr.Value{
			"file_path": nullableString(value.FilePath),
			"name":      nullableString(value.Name),
		})
		if diags.HasError() {
			return types.ListNull(types.ObjectType{AttrTypes: repositoryModuleObjectType()})
		}
		elements = append(elements, obj)
	}
	return types.ListValueMust(types.ObjectType{AttrTypes: repositoryModuleObjectType()}, elements)
}

func floatMapToTerraform(values map[string]float64) types.Map {
	if len(values) == 0 {
		return types.MapNull(types.Float64Type)
	}
	elements := make(map[string]attr.Value, len(values))
	for key, value := range values {
		elements[key] = types.Float64Value(value)
	}
	result, _ := types.MapValue(types.Float64Type, elements)
	return result
}

func nullableInt64(value *int64) types.Int64 {
	if value == nil {
		return types.Int64Null()
	}
	return types.Int64Value(*value)
}

func nullableBool(value *bool) types.Bool {
	if value == nil {
		return types.BoolNull()
	}
	return types.BoolValue(*value)
}
