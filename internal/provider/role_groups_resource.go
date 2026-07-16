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
	_ resource.Resource                = &roleGroupsResource{}
	_ resource.ResourceWithConfigure   = &roleGroupsResource{}
	_ resource.ResourceWithImportState = &roleGroupsResource{}
)

type roleGroupsResource struct {
	client *Client
}

type roleGroupsResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	AdminIDs    types.List   `tfsdk:"admin_ids"`
	MemberIDs   types.List   `tfsdk:"member_ids"`
}

func NewRoleGroupsResource() resource.Resource {
	return &roleGroupsResource{}
}

func (r *roleGroupsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rolegroups"
}

func (r *roleGroupsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages Apiiro user groups.",
		Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Computed: true},
			"name":        schema.StringAttribute{Required: true},
			"description": schema.StringAttribute{Optional: true, Computed: true},
			"admin_ids":   schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType},
			"member_ids":  schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType},
		},
	}
}

func (r *roleGroupsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *roleGroupsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "role groups create requested")
	var plan roleGroupsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := roleGroupBody{
		AdminIDs:    listToStrings(ctx, plan.AdminIDs),
		Description: stringPtr(plan.Description),
		MemberIDs:   listToStrings(ctx, plan.MemberIDs),
		Name:        plan.Name.ValueString(),
	}
	key, err := r.client.createRoleGroup(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create Group", err.Error())
		return
	}
	fresh, err := r.client.getRoleGroup(ctx, key)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Group", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, r.toState(fresh))...)
}

func (r *roleGroupsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state roleGroupsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "role groups read requested", map[string]any{"id": state.ID.ValueString()})
	fresh, err := r.client.getRoleGroup(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Group", err.Error())
		return
	}
	if fresh == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, r.toState(fresh))...)
}

func (r *roleGroupsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan roleGroupsResourceModel
	var state roleGroupsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "role groups update requested", map[string]any{"id": state.ID.ValueString()})
	body := roleGroupBody{
		AdminIDs:    listToStrings(ctx, mergeList(plan.AdminIDs, state.AdminIDs)),
		Description: stringPtr(mergeString(plan.Description, state.Description)),
		MemberIDs:   listToStrings(ctx, mergeList(plan.MemberIDs, state.MemberIDs)),
		Name:        mergeString(plan.Name, state.Name).ValueString(),
	}
	if err := r.client.updateRoleGroup(ctx, state.ID.ValueString(), body); err != nil {
		resp.Diagnostics.AddError("Unable to Update Group", err.Error())
		return
	}
	fresh, err := r.client.getRoleGroup(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Group", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, r.toState(fresh))...)
}

func (r *roleGroupsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state roleGroupsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "role groups delete requested", map[string]any{"id": state.ID.ValueString()})
	if err := r.client.deleteRoleGroup(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Unable to Delete Group", err.Error())
	}
}

func (r *roleGroupsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *roleGroupsResource) toState(group *roleGroupBody) *roleGroupsResourceModel {
	if group == nil {
		return nil
	}
	return &roleGroupsResourceModel{
		ID:          nullableString(group.Key),
		Name:        types.StringValue(group.Name),
		Description: nullableString(group.Description),
		AdminIDs:    stringSliceToList(group.AdminIDs),
		MemberIDs:   stringSliceToList(group.MemberIDs),
	}
}
