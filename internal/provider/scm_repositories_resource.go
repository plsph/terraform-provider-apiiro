package provider

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &scmRepositoriesResource{}
	_ resource.ResourceWithConfigure   = &scmRepositoriesResource{}
	_ resource.ResourceWithImportState = &scmRepositoriesResource{}
)

type scmRepositoriesResource struct {
	client *Client
}

type scmRepositoriesResourceModel struct {
	ID                            types.String `tfsdk:"id"`
	ScmRepositoryKey              types.String `tfsdk:"scm_repository_key"`
	Monitored                     types.Bool   `tfsdk:"monitored"`
	MonitoredBranches             types.Set    `tfsdk:"monitored_branches"`
	Tags                          types.Map    `tfsdk:"tags"`
	Name                          types.String `tfsdk:"name"`
	Provider                      types.String `tfsdk:"provider"`
	MonitorStatus                 types.String `tfsdk:"monitor_status"`
	DefaultBranch                 types.String `tfsdk:"default_branch"`
	URL                           types.String `tfsdk:"url"`
	ServerURL                     types.String `tfsdk:"server_url"`
	HTTPCloneURL                  types.String `tfsdk:"http_clone_url"`
	SSHURL                        types.String `tfsdk:"ssh_url"`
	RepositoryExternalID          types.String `tfsdk:"repository_external_id"`
	ProjectID                     types.String `tfsdk:"project_id"`
	ProjectExternalID             types.String `tfsdk:"project_external_id"`
	ProjectURL                    types.String `tfsdk:"project_url"`
	CreatedAt                     types.String `tfsdk:"created_at"`
	LastMonitoringChangeTimestamp types.String `tfsdk:"last_monitoring_change_timestamp"`
	IsPublic                      types.Bool   `tfsdk:"is_public"`
	IsArchived                    types.Bool   `tfsdk:"is_archived"`
	IsIgnored                     types.Bool   `tfsdk:"is_ignored"`
	IgnoreReason                  types.String `tfsdk:"ignore_reason"`
	IgnoredBy                     types.String `tfsdk:"ignored_by"`
}

func NewScmRepositoriesResource() resource.Resource {
	return &scmRepositoriesResource{}
}

func (r *scmRepositoriesResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_scmrepositories"
}

func (r *scmRepositoriesResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages monitoring and tags for an existing Apiiro SCM repository.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"scm_repository_key": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Unique Apiiro SCM repository key.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"monitored": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether the repository should be monitored.",
			},
			"monitored_branches": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Specific monitored branches. If omitted and monitored is true, the default branch is monitored.",
			},
			"tags": schema.MapAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Repository tags managed through Apiiro tag API.",
			},
			"name":                             schema.StringAttribute{Computed: true},
			"provider":                         schema.StringAttribute{Computed: true},
			"monitor_status":                   schema.StringAttribute{Computed: true},
			"default_branch":                   schema.StringAttribute{Computed: true},
			"url":                              schema.StringAttribute{Computed: true},
			"server_url":                       schema.StringAttribute{Computed: true},
			"http_clone_url":                   schema.StringAttribute{Computed: true},
			"ssh_url":                          schema.StringAttribute{Computed: true},
			"repository_external_id":           schema.StringAttribute{Computed: true},
			"project_id":                       schema.StringAttribute{Computed: true},
			"project_external_id":              schema.StringAttribute{Computed: true},
			"project_url":                      schema.StringAttribute{Computed: true},
			"created_at":                       schema.StringAttribute{Computed: true},
			"last_monitoring_change_timestamp": schema.StringAttribute{Computed: true},
			"is_public":                        schema.BoolAttribute{Computed: true},
			"is_archived":                      schema.BoolAttribute{Computed: true},
			"is_ignored":                       schema.BoolAttribute{Computed: true},
			"ignore_reason":                    schema.StringAttribute{Computed: true},
			"ignored_by":                       schema.StringAttribute{Computed: true},
		},
	}
}

func (r *scmRepositoriesResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *Client, got: %T", req.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *scmRepositoriesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan scmRepositoriesResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.Monitored.IsNull() || plan.Monitored.IsUnknown() {
		plan.Monitored = types.BoolValue(true)
	}

	repoKey := plan.ScmRepositoryKey.ValueString()
	repo, err := r.client.getScmRepositoryByKey(repoKey)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Query SCM Repository", err.Error())
		return
	}
	if repo == nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("scm_repository_key"),
			"SCM Repository Not Found",
			"No SCM repository with the specified key exists.",
		)
		return
	}

	desiredBranches := setToStrings(ctx, plan.MonitoredBranches)
	if !plan.Monitored.ValueBool() && len(desiredBranches) > 0 {
		resp.Diagnostics.AddAttributeError(
			path.Root("monitored_branches"),
			"Invalid Branch Configuration",
			"`monitored_branches` cannot be set when `monitored` is false.",
		)
		return
	}

	if plan.Monitored.ValueBool() {
		if len(desiredBranches) == 0 {
			if err := r.client.monitorRepository(repoKey); err != nil && !isAlreadyMonitoredError(err) {
				resp.Diagnostics.AddError("Unable to Monitor SCM Repository", err.Error())
				return
			}
		} else {
			for _, branch := range desiredBranches {
				if err := r.client.monitorBranch(repoKey, branch); err != nil {
					resp.Diagnostics.AddError("Unable to Monitor SCM Repository Branch", err.Error())
					return
				}
			}
		}
	} else {
		if err := r.client.unmonitorRepository(repoKey); err != nil && !isAlreadyUnmonitoredError(err) {
			resp.Diagnostics.AddError("Unable to Unmonitor SCM Repository", err.Error())
			return
		}
	}

	planTags := mapFromTerraform(ctx, plan.Tags)
	for k, v := range planTags {
		if err := r.client.upsertRepositoryTag(repoKey, k, v); err != nil {
			resp.Diagnostics.AddError("Unable to Upsert Repository Tag", err.Error())
			return
		}
	}

	state, err := r.readState(ctx, repoKey)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read SCM Repository", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *scmRepositoriesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state scmRepositoriesResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	repoKey := state.ScmRepositoryKey.ValueString()
	fresh, err := r.readState(ctx, repoKey)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read SCM Repository", err.Error())
		return
	}
	if fresh == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, fresh)...)
}

func (r *scmRepositoriesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan scmRepositoriesResourceModel
	var state scmRepositoriesResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	repoKey := state.ScmRepositoryKey.ValueString()
	oldMonitored := state.Monitored.ValueBool()
	newMonitored := plan.Monitored.ValueBool()
	oldBranches := setToStrings(ctx, state.MonitoredBranches)
	newBranches := setToStrings(ctx, plan.MonitoredBranches)

	if !newMonitored && len(newBranches) > 0 {
		resp.Diagnostics.AddAttributeError(
			path.Root("monitored_branches"),
			"Invalid Branch Configuration",
			"`monitored_branches` cannot be set when `monitored` is false.",
		)
		return
	}

	if oldMonitored != newMonitored {
		if newMonitored {
			if err := r.client.monitorRepository(repoKey); err != nil && !isAlreadyMonitoredError(err) {
				resp.Diagnostics.AddError("Unable to Monitor SCM Repository", err.Error())
				return
			}
		} else {
			if err := r.client.unmonitorRepository(repoKey); err != nil && !isAlreadyUnmonitoredError(err) {
				resp.Diagnostics.AddError("Unable to Unmonitor SCM Repository", err.Error())
				return
			}
		}
	}

	if newMonitored {
		toUnmonitor := diff(oldBranches, newBranches)
		toMonitor := diff(newBranches, oldBranches)

		for _, branch := range toUnmonitor {
			if err := r.client.unmonitorBranch(repoKey, branch); err != nil {
				resp.Diagnostics.AddError("Unable to Unmonitor SCM Repository Branch", err.Error())
				return
			}
		}
		for _, branch := range toMonitor {
			if err := r.client.monitorBranch(repoKey, branch); err != nil {
				resp.Diagnostics.AddError("Unable to Monitor SCM Repository Branch", err.Error())
				return
			}
		}
	}

	oldTags := mapFromTerraform(ctx, state.Tags)
	newTags := mapFromTerraform(ctx, plan.Tags)
	removed := mapDiffKeys(oldTags, newTags)
	for _, key := range removed {
		if err := r.client.deleteRepositoryTag(repoKey, key); err != nil {
			resp.Diagnostics.AddError("Unable to Delete Repository Tag", err.Error())
			return
		}
	}
	for key, val := range newTags {
		if oldVal, ok := oldTags[key]; !ok || oldVal != val {
			if err := r.client.upsertRepositoryTag(repoKey, key, val); err != nil {
				resp.Diagnostics.AddError("Unable to Upsert Repository Tag", err.Error())
				return
			}
		}
	}

	fresh, err := r.readState(ctx, repoKey)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read SCM Repository", err.Error())
		return
	}
	if fresh == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, fresh)...)
}

func (r *scmRepositoriesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state scmRepositoriesResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	repoKey := state.ScmRepositoryKey.ValueString()
	for tagName := range mapFromTerraform(ctx, state.Tags) {
		if err := r.client.deleteRepositoryTag(repoKey, tagName); err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "no tag with the specified name") {
				continue
			}
			resp.Diagnostics.AddError("Unable to Delete Repository Tag", err.Error())
			return
		}
	}

	if state.Monitored.ValueBool() {
		if err := r.client.unmonitorRepository(repoKey); err != nil && !isAlreadyUnmonitoredError(err) {
			resp.Diagnostics.AddError("Unable to Unmonitor SCM Repository", err.Error())
			return
		}
	}
}

func (r *scmRepositoriesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("scm_repository_key"), req, resp)
}

func (r *scmRepositoriesResource) readState(ctx context.Context, repoKey string) (*scmRepositoriesResourceModel, error) {
	repo, err := r.client.getScmRepositoryByKey(repoKey)
	if err != nil {
		return nil, err
	}
	if repo == nil {
		return nil, nil
	}

	tagsMap := make(map[string]string)
	tags, err := r.client.listRepositoryTags(repoKey)
	if err != nil {
		if !strings.Contains(strings.ToLower(err.Error()), "404") {
			return nil, err
		}
	} else {
		for _, tag := range tags {
			if strings.TrimSpace(tag.Name) != "" {
				tagsMap[tag.Name] = tag.Value
			}
		}
	}

	branches := append([]string{}, repo.MonitoredBranches...)
	sort.Strings(branches)

	state := &scmRepositoriesResourceModel{
		ID:                            types.StringValue(repoKey),
		ScmRepositoryKey:              types.StringValue(repoKey),
		Monitored:                     types.BoolValue(strings.EqualFold(valueOrEmpty(repo.MonitorStatus), "Monitored")),
		MonitoredBranches:             stringsToSet(branches),
		Tags:                          mapToTerraform(tagsMap),
		Name:                          nullableString(repo.Name),
		Provider:                      nullableString(repo.Provider),
		MonitorStatus:                 nullableString(repo.MonitorStatus),
		DefaultBranch:                 nullableString(repo.DefaultBranch),
		URL:                           nullableString(repo.URL),
		ServerURL:                     nullableString(repo.ServerURL),
		HTTPCloneURL:                  nullableString(repo.HTTPCloneURL),
		SSHURL:                        nullableString(repo.SSHURL),
		RepositoryExternalID:          nullableString(repo.RepositoryExternalID),
		ProjectID:                     nullableString(repo.ProjectID),
		ProjectExternalID:             nullableString(repo.ProjectExternalID),
		ProjectURL:                    nullableString(repo.ProjectURL),
		CreatedAt:                     nullableString(repo.CreatedAt),
		LastMonitoringChangeTimestamp: nullableString(repo.LastMonitoringChangeTimestamp),
		IsPublic:                      types.BoolValue(repo.IsPublic),
		IsArchived:                    types.BoolValue(repo.IsArchived),
		IsIgnored:                     types.BoolValue(repo.IsIgnored),
		IgnoreReason:                  nullableString(repo.IgnoreReason),
		IgnoredBy:                     nullableString(repo.IgnoredBy),
	}
	return state, nil
}

func nullableString(v *string) types.String {
	if v == nil || strings.TrimSpace(*v) == "" {
		return types.StringNull()
	}
	return types.StringValue(*v)
}

func valueOrEmpty(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func setToStrings(ctx context.Context, setVal types.Set) []string {
	if setVal.IsNull() || setVal.IsUnknown() {
		return nil
	}
	var values []string
	_ = setVal.ElementsAs(ctx, &values, false)
	clean := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		clean = append(clean, v)
	}
	sort.Strings(clean)
	return clean
}

func stringsToSet(values []string) types.Set {
	elements := make([]types.String, 0, len(values))
	for _, v := range values {
		elements = append(elements, types.StringValue(v))
	}
	setValue, diags := types.SetValueFrom(context.Background(), types.StringType, elements)
	if diags.HasError() {
		return types.SetNull(types.StringType)
	}
	return setValue
}

func mapFromTerraform(ctx context.Context, input types.Map) map[string]string {
	if input.IsNull() || input.IsUnknown() {
		return map[string]string{}
	}
	out := map[string]string{}
	_ = input.ElementsAs(ctx, &out, false)
	return out
}

func mapToTerraform(in map[string]string) types.Map {
	if len(in) == 0 {
		return types.MapValueMust(types.StringType, map[string]attr.Value{})
	}
	values := make(map[string]attr.Value, len(in))
	for k, v := range in {
		values[k] = types.StringValue(v)
	}
	return types.MapValueMust(types.StringType, values)
}

func diff(a, b []string) []string {
	base := map[string]struct{}{}
	for _, v := range b {
		base[v] = struct{}{}
	}
	out := make([]string, 0)
	for _, v := range a {
		if _, ok := base[v]; !ok {
			out = append(out, v)
		}
	}
	sort.Strings(out)
	return out
}

func mapDiffKeys(oldMap, newMap map[string]string) []string {
	keys := make([]string, 0)
	for key := range oldMap {
		if _, ok := newMap[key]; !ok {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys
}

func isAlreadyMonitoredError(err error) bool {
	errText := strings.ToLower(err.Error())
	return strings.Contains(errText, "already being monitored")
}

func isAlreadyUnmonitoredError(err error) bool {
	errText := strings.ToLower(err.Error())
	return strings.Contains(errText, "not currently monitored")
}
