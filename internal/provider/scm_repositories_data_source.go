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
	_ datasource.DataSource              = &scmRepositoriesDataSource{}
	_ datasource.DataSourceWithConfigure = &scmRepositoriesDataSource{}
)

type scmRepositoriesDataSource struct {
	client *Client
}

type scmRepositoriesDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	RepositoryName   types.String `tfsdk:"repository_name"`
	ScmRepositoryKey types.String `tfsdk:"scm_repository_key"`
	Repositories     types.List   `tfsdk:"repositories"`
}

func NewScmRepositoriesDataSource() datasource.DataSource {
	return &scmRepositoriesDataSource{}
}

func (d *scmRepositoriesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_scmrepositories"
}

func (d *scmRepositoriesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches SCM repositories from Apiiro with optional filtering by name or key.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"repository_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional case-insensitive filter for SCM repository name.",
			},
			"scm_repository_key": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional exact filter for SCM repository key.",
			},
			"repositories": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of matching SCM repositories.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key":                              schema.StringAttribute{Computed: true},
						"id":                               schema.StringAttribute{Computed: true},
						"name":                             schema.StringAttribute{Computed: true},
						"provider":                         schema.StringAttribute{Computed: true},
						"server_url":                       schema.StringAttribute{Computed: true},
						"url":                              schema.StringAttribute{Computed: true},
						"project_url":                      schema.StringAttribute{Computed: true},
						"http_clone_url":                   schema.StringAttribute{Computed: true},
						"ssh_url":                          schema.StringAttribute{Computed: true},
						"project_id":                       schema.StringAttribute{Computed: true},
						"project_external_id":              schema.StringAttribute{Computed: true},
						"repository_external_id":           schema.StringAttribute{Computed: true},
						"default_branch":                   schema.StringAttribute{Computed: true},
						"monitor_status":                   schema.StringAttribute{Computed: true},
						"monitored_branches":               schema.ListAttribute{Computed: true, ElementType: types.StringType},
						"languages":                        schema.ListAttribute{Computed: true, ElementType: types.StringType},
						"associated_project_ids":           schema.ListAttribute{Computed: true, ElementType: types.StringType},
						"approximate_size_mb":              schema.Int64Attribute{Computed: true},
						"approximate_file_count":           schema.Int64Attribute{Computed: true},
						"is_public":                        schema.BoolAttribute{Computed: true},
						"is_archived":                      schema.BoolAttribute{Computed: true},
						"is_ignored":                       schema.BoolAttribute{Computed: true},
						"ignore_reason":                    schema.StringAttribute{Computed: true},
						"ignored_by":                       schema.StringAttribute{Computed: true},
						"created_at":                       schema.StringAttribute{Computed: true},
						"last_monitoring_change_timestamp": schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *scmRepositoriesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *Client, got: %T", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *scmRepositoriesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state scmRepositoriesDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	repositoryNameFilter := strings.TrimSpace(state.RepositoryName.ValueString())
	repositoryKeyFilter := strings.TrimSpace(state.ScmRepositoryKey.ValueString())

	repositories, err := d.client.listScmRepositories()
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read SCM Repositories", err.Error())
		return
	}

	values := make([]attr.Value, 0, len(repositories))
	for _, repo := range repositories {
		if repositoryNameFilter != "" {
			repoName := strings.TrimSpace(valueOrEmpty(repo.Name))
			if !strings.Contains(strings.ToLower(repoName), strings.ToLower(repositoryNameFilter)) {
				continue
			}
		}
		if repositoryKeyFilter != "" {
			if valueOrEmpty(repo.Key) != repositoryKeyFilter {
				continue
			}
		}

		obj, diags := types.ObjectValue(scmRepositoryDataSourceObjectType(), map[string]attr.Value{
			"key":                              nullableString(repo.Key),
			"id":                               nullableString(repo.ID),
			"name":                             nullableString(repo.Name),
			"provider":                         nullableString(repo.Provider),
			"server_url":                       nullableString(repo.ServerURL),
			"url":                              nullableString(repo.URL),
			"project_url":                      nullableString(repo.ProjectURL),
			"http_clone_url":                   nullableString(repo.HTTPCloneURL),
			"ssh_url":                          nullableString(repo.SSHURL),
			"project_id":                       nullableString(repo.ProjectID),
			"project_external_id":              nullableString(repo.ProjectExternalID),
			"repository_external_id":           nullableString(repo.RepositoryExternalID),
			"default_branch":                   nullableString(repo.DefaultBranch),
			"monitor_status":                   nullableString(repo.MonitorStatus),
			"monitored_branches":               stringSliceToList(repo.MonitoredBranches),
			"languages":                        stringSliceToList(repo.Languages),
			"associated_project_ids":           stringSliceToList(repo.AssociatedProjectIDs),
			"approximate_size_mb":              nullableInt(repo.ApproximateSizeMB),
			"approximate_file_count":           nullableInt(repo.ApproximateFileCount),
			"is_public":                        types.BoolValue(repo.IsPublic),
			"is_archived":                      types.BoolValue(repo.IsArchived),
			"is_ignored":                       types.BoolValue(repo.IsIgnored),
			"ignore_reason":                    nullableString(repo.IgnoreReason),
			"ignored_by":                       nullableString(repo.IgnoredBy),
			"created_at":                       nullableString(repo.CreatedAt),
			"last_monitoring_change_timestamp": nullableString(repo.LastMonitoringChangeTimestamp),
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		values = append(values, obj)
	}

	repoList, diags := types.ListValue(types.ObjectType{AttrTypes: scmRepositoryDataSourceObjectType()}, values)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.ID = types.StringValue(fmt.Sprintf("scmrepositories:%s:%s", repositoryNameFilter, repositoryKeyFilter))
	state.Repositories = repoList

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func scmRepositoryDataSourceObjectType() map[string]attr.Type {
	return map[string]attr.Type{
		"key":                              types.StringType,
		"id":                               types.StringType,
		"name":                             types.StringType,
		"provider":                         types.StringType,
		"server_url":                       types.StringType,
		"url":                              types.StringType,
		"project_url":                      types.StringType,
		"http_clone_url":                   types.StringType,
		"ssh_url":                          types.StringType,
		"project_id":                       types.StringType,
		"project_external_id":              types.StringType,
		"repository_external_id":           types.StringType,
		"default_branch":                   types.StringType,
		"monitor_status":                   types.StringType,
		"monitored_branches":               types.ListType{ElemType: types.StringType},
		"languages":                        types.ListType{ElemType: types.StringType},
		"associated_project_ids":           types.ListType{ElemType: types.StringType},
		"approximate_size_mb":              types.Int64Type,
		"approximate_file_count":           types.Int64Type,
		"is_public":                        types.BoolType,
		"is_archived":                      types.BoolType,
		"is_ignored":                       types.BoolType,
		"ignore_reason":                    types.StringType,
		"ignored_by":                       types.StringType,
		"created_at":                       types.StringType,
		"last_monitoring_change_timestamp": types.StringType,
	}
}

func stringSliceToList(values []string) types.List {
	if len(values) == 0 {
		return types.ListNull(types.StringType)
	}
	elements := make([]attr.Value, 0, len(values))
	for _, v := range values {
		elements = append(elements, types.StringValue(v))
	}
	return types.ListValueMust(types.StringType, elements)
}

func nullableInt(v *int) types.Int64 {
	if v == nil {
		return types.Int64Null()
	}
	return types.Int64Value(int64(*v))
}
