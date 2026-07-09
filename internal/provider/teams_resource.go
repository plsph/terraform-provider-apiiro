package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &teamsResource{}
	_ resource.ResourceWithConfigure   = &teamsResource{}
	_ resource.ResourceWithImportState = &teamsResource{}
)

type teamsResource struct {
	client *Client
}

type teamsResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Description     types.String `tfsdk:"description"`
	ParentKey       types.String `tfsdk:"parent_key"`
	Applications    types.List   `tfsdk:"applications"`
	PointsOfContact types.List   `tfsdk:"points_of_contact"`
	ProjectUrls     types.List   `tfsdk:"project_urls"`
	RepositoryUrls  types.List   `tfsdk:"repository_urls"`
	ApplicationTags types.Map    `tfsdk:"application_tags"`
	Tags            types.Map    `tfsdk:"tags"`
	RiskScore       types.Int64  `tfsdk:"risk_score"`
}

func NewTeamsResource() resource.Resource {
	return &teamsResource{}
}

func (r *teamsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_teams"
}

func (r *teamsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages organizational teams in Apiiro.",
		Attributes: map[string]schema.Attribute{
			"id":           schema.StringAttribute{Computed: true},
			"name":         schema.StringAttribute{Optional: true, Computed: true},
			"description":  schema.StringAttribute{Optional: true, Computed: true},
			"parent_key":   schema.StringAttribute{Optional: true, Computed: true},
			"applications": schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType},
			"points_of_contact": schema.ListNestedAttribute{Optional: true, Computed: true, NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
				"identity": schema.StringAttribute{Optional: true, Computed: true},
				"title":    schema.StringAttribute{Optional: true, Computed: true},
			}}},
			"project_urls":     schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType},
			"repository_urls":  schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType},
			"application_tags": schema.MapAttribute{Optional: true, Computed: true, ElementType: types.StringType},
			"tags":             schema.MapAttribute{Optional: true, Computed: true, ElementType: types.StringType},
			"risk_score":       schema.Int64Attribute{Computed: true},
		},
	}
}

func (r *teamsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *Client, got: %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *teamsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan teamsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := orgTeamBody{
		Applications:    listToStrings(ctx, plan.Applications),
		ApplicationTags: tagsMapToBodies(mapFromTerraform(ctx, plan.ApplicationTags)),
		Description:     stringPtr(plan.Description),
		Name:            stringPtr(plan.Name),
		ParentKey:       stringPtr(plan.ParentKey),
		PointsOfContact: pointOfContactListToBodies(ctx, plan.PointsOfContact),
		ProjectUrls:     listToStrings(ctx, plan.ProjectUrls),
		RepositoryUrls:  listToStrings(ctx, plan.RepositoryUrls),
		Tags:            tagsMapToBodies(mapFromTerraform(ctx, plan.Tags)),
	}

	key, err := r.client.createTeam(body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create Team", err.Error())
		return
	}

	fresh, err := r.client.getTeam(key)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Team", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, r.toState(fresh))...)
}

func (r *teamsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state teamsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	fresh, err := r.client.getTeam(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Team", err.Error())
		return
	}
	if fresh == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, r.toState(fresh))...)
}

func (r *teamsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan teamsResourceModel
	var state teamsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := orgTeamBody{
		Key:             stringPtr(state.ID),
		Applications:    listToStrings(ctx, mergeList(plan.Applications, state.Applications)),
		ApplicationTags: tagsMapToBodies(mapFromTerraform(ctx, mergeMap(plan.ApplicationTags, state.ApplicationTags))),
		Description:     stringPtr(mergeString(plan.Description, state.Description)),
		Name:            stringPtr(mergeString(plan.Name, state.Name)),
		ParentKey:       stringPtr(mergeString(plan.ParentKey, state.ParentKey)),
		PointsOfContact: pointOfContactListToBodies(ctx, mergeList(plan.PointsOfContact, state.PointsOfContact)),
		ProjectUrls:     listToStrings(ctx, mergeList(plan.ProjectUrls, state.ProjectUrls)),
		RepositoryUrls:  listToStrings(ctx, mergeList(plan.RepositoryUrls, state.RepositoryUrls)),
		Tags:            tagsMapToBodies(mapFromTerraform(ctx, mergeMap(plan.Tags, state.Tags))),
	}

	key, err := r.client.updateTeam(state.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Update Team", err.Error())
		return
	}

	fresh, err := r.client.getTeam(key)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Team", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, r.toState(fresh))...)
}

func (r *teamsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state teamsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.deleteTeam(state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Unable to Delete Team", err.Error())
	}
}

func (r *teamsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *teamsResource) toState(team *orgTeamBody) *teamsResourceModel {
	if team == nil {
		return nil
	}
	return &teamsResourceModel{
		ID:              nullableString(team.Key),
		Name:            nullableString(team.Name),
		Description:     nullableString(team.Description),
		ParentKey:       nullableString(team.ParentKey),
		Applications:    stringSliceToList(team.Applications),
		PointsOfContact: pointOfContactBodiesToList(team.PointsOfContact),
		ProjectUrls:     stringSliceToList(team.ProjectUrls),
		RepositoryUrls:  stringSliceToList(team.RepositoryUrls),
		ApplicationTags: mapToTerraform(tagsBodiesToMap(team.ApplicationTags)),
		Tags:            mapToTerraform(tagsBodiesToMap(team.Tags)),
		RiskScore:       types.Int64Value(int64(valueOrZero(team.RiskScore))),
	}
}
