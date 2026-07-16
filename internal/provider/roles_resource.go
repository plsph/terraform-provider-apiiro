package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &rolesResource{}
	_ resource.ResourceWithConfigure   = &rolesResource{}
	_ resource.ResourceWithImportState = &rolesResource{}
)

type rolesResource struct {
	client *Client
}

type rolesResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Description     types.String `tfsdk:"description"`
	ApiiroGroupIDs  types.List   `tfsdk:"apiiro_group_ids"`
	IdpGroupIDs     types.List   `tfsdk:"idp_group_ids"`
	Permissions     types.Map    `tfsdk:"permissions"`
	ApplicationKeys types.List   `tfsdk:"application_keys"`
	OrgTeamKeys     types.List   `tfsdk:"org_team_keys"`
	RepositoryUrls  types.List   `tfsdk:"repository_urls"`
}

func NewRolesResource() resource.Resource {
	return &rolesResource{}
}

func (r *rolesResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_roles"
}

func (r *rolesResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages roles in Apiiro.",
		Attributes: map[string]schema.Attribute{
			"id":               schema.StringAttribute{Computed: true},
			"name":             schema.StringAttribute{Required: true},
			"description":      schema.StringAttribute{Optional: true, Computed: true},
			"apiiro_group_ids": schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType},
			"idp_group_ids":    schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType},
			"permissions":      schema.MapAttribute{Optional: true, Computed: true, ElementType: types.StringType},
			"application_keys": schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType},
			"org_team_keys":    schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType},
			"repository_urls":  schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType},
		},
	}
}

func (r *rolesResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *rolesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "roles create requested")
	var plan rolesResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := roleBody{
		ApiiroGroupIds: listToStrings(ctx, plan.ApiiroGroupIDs),
		Description:    stringPtr(plan.Description),
		IdpGroupIds:    listToStrings(ctx, plan.IdpGroupIDs),
		Name:           plan.Name.ValueString(),
		Permissions:    mapFromTerraform(ctx, plan.Permissions),
		Scope: &roleScopeBody{
			ApplicationKeys: listToStrings(ctx, plan.ApplicationKeys),
			OrgTeamKeys:     listToStrings(ctx, plan.OrgTeamKeys),
			RepositoryUrls:  listToStrings(ctx, plan.RepositoryUrls),
		},
	}

	key, err := r.client.createRole(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create Role", err.Error())
		return
	}

	state, err := r.readState(key, &body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Role", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *rolesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state rolesResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "roles read requested", map[string]any{"id": state.ID.ValueString()})

	fresh, err := r.client.getRole(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Role", err.Error())
		return
	}
	if fresh == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	readState, err := r.readState(state.ID.ValueString(), fresh)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Role", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, readState)...)
}

func (r *rolesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan rolesResourceModel
	var state rolesResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "roles update requested", map[string]any{"id": state.ID.ValueString()})

	if err := r.client.deleteRole(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Unable to Replace Role", err.Error())
		return
	}

	body := roleBody{
		ApiiroGroupIds: listToStrings(ctx, plan.ApiiroGroupIDs),
		Description:    stringPtr(plan.Description),
		IdpGroupIds:    listToStrings(ctx, plan.IdpGroupIDs),
		Name:           plan.Name.ValueString(),
		Permissions:    mapFromTerraform(ctx, plan.Permissions),
		Scope: &roleScopeBody{
			ApplicationKeys: listToStrings(ctx, plan.ApplicationKeys),
			OrgTeamKeys:     listToStrings(ctx, plan.OrgTeamKeys),
			RepositoryUrls:  listToStrings(ctx, plan.RepositoryUrls),
		},
	}

	key, err := r.client.createRole(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Recreate Role", err.Error())
		return
	}
	readState, err := r.readState(key, &body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Role", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, readState)...)
}

func (r *rolesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state rolesResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "roles delete requested", map[string]any{"id": state.ID.ValueString()})
	if err := r.client.deleteRole(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Unable to Delete Role", err.Error())
	}
}

func (r *rolesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *rolesResource) readState(key string, role *roleBody) (*rolesResourceModel, error) {
	if role == nil {
		return nil, nil
	}
	return &rolesResourceModel{
		ID:              types.StringValue(key),
		Name:            types.StringValue(role.Name),
		Description:     nullableString(role.Description),
		ApiiroGroupIDs:  stringSliceToList(role.ApiiroGroupIds),
		IdpGroupIDs:     stringSliceToList(role.IdpGroupIds),
		Permissions:     mapToTerraform(role.Permissions),
		ApplicationKeys: stringSliceToList(roleScopeValue(role.Scope).ApplicationKeys),
		OrgTeamKeys:     stringSliceToList(roleScopeValue(role.Scope).OrgTeamKeys),
		RepositoryUrls:  stringSliceToList(roleScopeValue(role.Scope).RepositoryUrls),
	}, nil
}

func roleScopeValue(scope *roleScopeBody) roleScopeBody {
	if scope == nil {
		return roleScopeBody{}
	}
	return *scope
}
