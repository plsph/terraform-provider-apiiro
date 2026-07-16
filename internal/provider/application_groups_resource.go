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
	_ resource.Resource                = &applicationGroupsResource{}
	_ resource.ResourceWithConfigure   = &applicationGroupsResource{}
	_ resource.ResourceWithImportState = &applicationGroupsResource{}
)

type applicationGroupsResource struct {
	client *Client
}

type applicationGroupsResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Applications    types.List   `tfsdk:"applications"`
	PointsOfContact types.List   `tfsdk:"points_of_contact"`
	Tags            types.List   `tfsdk:"tags"`
}

func NewApplicationGroupsResource() resource.Resource {
	return &applicationGroupsResource{}
}

func (r *applicationGroupsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_applicationgroups"
}

func (r *applicationGroupsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages application groups in Apiiro.",
		Attributes: map[string]schema.Attribute{
			"id":           schema.StringAttribute{Computed: true},
			"name":         schema.StringAttribute{Optional: true, Computed: true},
			"applications": schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType},
			"points_of_contact": schema.ListNestedAttribute{Optional: true, Computed: true, NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
				"identity": schema.StringAttribute{Optional: true, Computed: true},
				"title":    schema.StringAttribute{Optional: true, Computed: true},
			}}},
			"tags": schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType},
		},
	}
}

func (r *applicationGroupsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *applicationGroupsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "application groups create requested")
	var plan applicationGroupsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := applicationGroupBody{
		Applications:    listToStrings(ctx, plan.Applications),
		Name:            stringPtr(plan.Name),
		PointsOfContact: pointOfContactListToBodies(ctx, plan.PointsOfContact),
		Tags:            listToStrings(ctx, plan.Tags),
	}

	fresh, err := r.client.createApplicationGroup(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create Application Group", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, r.toState(fresh))...)
}

func (r *applicationGroupsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state applicationGroupsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "application groups read requested", map[string]any{"id": state.ID.ValueString()})

	fresh, err := r.client.getApplicationGroup(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Application Group", err.Error())
		return
	}
	if fresh == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, r.toState(fresh))...)
}

func (r *applicationGroupsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan applicationGroupsResourceModel
	var state applicationGroupsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "application groups update requested", map[string]any{"id": state.ID.ValueString()})

	body := applicationGroupBody{
		Key:             stringPtr(state.ID),
		Applications:    listToStrings(ctx, mergeList(plan.Applications, state.Applications)),
		Name:            stringPtr(mergeString(plan.Name, state.Name)),
		PointsOfContact: pointOfContactListToBodies(ctx, mergeList(plan.PointsOfContact, state.PointsOfContact)),
		Tags:            listToStrings(ctx, mergeList(plan.Tags, state.Tags)),
	}

	fresh, err := r.client.updateApplicationGroup(ctx, state.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Update Application Group", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, r.toState(fresh))...)
}

func (r *applicationGroupsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state applicationGroupsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "application groups delete requested", map[string]any{"id": state.ID.ValueString()})
	if err := r.client.deleteApplicationGroup(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Unable to Delete Application Group", err.Error())
	}
}

func (r *applicationGroupsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *applicationGroupsResource) toState(group *applicationGroupBody) *applicationGroupsResourceModel {
	if group == nil {
		return nil
	}
	return &applicationGroupsResourceModel{
		ID:              nullableString(group.Key),
		Name:            nullableString(group.Name),
		Applications:    stringSliceToList(group.Applications),
		PointsOfContact: pointOfContactBodiesToList(group.PointsOfContact),
		Tags:            stringSliceToList(group.Tags),
	}
}
