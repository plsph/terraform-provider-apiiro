package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &engagementsResource{}
	_ resource.ResourceWithConfigure   = &engagementsResource{}
	_ resource.ResourceWithImportState = &engagementsResource{}
)

type engagementsResource struct {
	client *Client
}

type engagementsResourceModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	ExternalURL       types.String `tfsdk:"external_url"`
	Type              types.String `tfsdk:"type"`
	Summary           types.String `tfsdk:"summary"`
	Scope             types.Object `tfsdk:"scope"`
	StartDate         types.String `tfsdk:"start_date"`
	EndDate           types.String `tfsdk:"end_date"`
	Status            types.String `tfsdk:"status"`
	Tags              types.Map    `tfsdk:"tags"`
	EngagementLeadKey types.String `tfsdk:"engagement_lead_key"`
	Provider          types.String `tfsdk:"provider_name"`
	Attachments       types.List   `tfsdk:"attachments"`
	Reporter          types.String `tfsdk:"reporter"`
}

func NewEngagementsResource() resource.Resource {
	return &engagementsResource{}
}

func (r *engagementsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_engagements"
}

func (r *engagementsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages engagements in Apiiro.",
		Attributes: map[string]schema.Attribute{
			"id":           schema.StringAttribute{Computed: true},
			"name":         schema.StringAttribute{Required: true},
			"external_url": schema.StringAttribute{Optional: true, Computed: true},
			"type":         schema.StringAttribute{Required: true},
			"summary":      schema.StringAttribute{Optional: true, Computed: true},
			"scope": schema.SingleNestedAttribute{Required: true, Attributes: map[string]schema.Attribute{
				"application_keys": schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType},
				"repository_keys":  schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType},
			}},
			"start_date":          schema.StringAttribute{Required: true},
			"end_date":            schema.StringAttribute{Optional: true, Computed: true},
			"status":              schema.StringAttribute{Required: true},
			"tags":                schema.MapAttribute{Optional: true, Computed: true, ElementType: types.StringType},
			"engagement_lead_key": schema.StringAttribute{Optional: true, Computed: true},
			"provider_name":       schema.StringAttribute{Optional: true, Computed: true},
			"attachments":         schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType},
			"reporter":            schema.StringAttribute{Optional: true, Computed: true},
		},
	}
}

func (r *engagementsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *engagementsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "engagements create requested")
	var plan engagementsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := engagementBody{
		Name:              stringPtr(plan.Name),
		ExternalURL:       stringPtr(plan.ExternalURL),
		Type:              stringPtr(plan.Type),
		Summary:           stringPtr(plan.Summary),
		Scope:             engagementScopeFromObject(plan.Scope),
		StartDate:         plan.StartDate.ValueString(),
		EndDate:           stringPtr(plan.EndDate),
		Status:            stringPtr(plan.Status),
		Tags:              mapFromTerraform(ctx, plan.Tags),
		EngagementLeadKey: stringPtr(plan.EngagementLeadKey),
		Provider:          stringPtr(plan.Provider),
		Attachments:       listToStrings(ctx, plan.Attachments),
		Reporter:          stringPtr(plan.Reporter),
	}
	created, err := r.client.createEngagement(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create Engagement", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, r.toState(created))...)
}

func (r *engagementsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state engagementsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "engagements read requested", map[string]any{"id": state.ID.ValueString()})

	fresh, err := r.client.getEngagement(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Engagement", err.Error())
		return
	}
	if fresh == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, r.toState(fresh))...)
}

func (r *engagementsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan engagementsResourceModel
	var state engagementsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "engagements update requested", map[string]any{"id": state.ID.ValueString()})

	body := engagementBody{
		Name:              stringPtr(mergeString(plan.Name, state.Name)),
		ExternalURL:       stringPtr(mergeString(plan.ExternalURL, state.ExternalURL)),
		Type:              stringPtr(mergeString(plan.Type, state.Type)),
		Summary:           stringPtr(mergeString(plan.Summary, state.Summary)),
		Scope:             engagementScopeFromObject(mergeObject(plan.Scope, state.Scope)),
		StartDate:         mergeString(plan.StartDate, state.StartDate).ValueString(),
		EndDate:           stringPtr(mergeString(plan.EndDate, state.EndDate)),
		Status:            stringPtr(mergeString(plan.Status, state.Status)),
		Tags:              mapFromTerraform(ctx, mergeMap(plan.Tags, state.Tags)),
		EngagementLeadKey: stringPtr(mergeString(plan.EngagementLeadKey, state.EngagementLeadKey)),
		Provider:          stringPtr(mergeString(plan.Provider, state.Provider)),
		Attachments:       listToStrings(ctx, mergeList(plan.Attachments, state.Attachments)),
		Reporter:          stringPtr(mergeString(plan.Reporter, state.Reporter)),
	}
	updated, err := r.client.updateEngagement(ctx, state.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Update Engagement", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, r.toState(updated))...)
}

func (r *engagementsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state engagementsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "engagements delete requested", map[string]any{"id": state.ID.ValueString()})
	if err := r.client.deleteEngagement(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Unable to Delete Engagement", err.Error())
	}
}

func (r *engagementsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *engagementsResource) toState(engagement *engagementBody) *engagementsResourceModel {
	if engagement == nil {
		return nil
	}
	return &engagementsResourceModel{
		ID:                nullableString(engagement.Key),
		Name:              nullableString(engagement.Name),
		ExternalURL:       nullableString(engagement.ExternalURL),
		Type:              nullableString(engagement.Type),
		Summary:           nullableString(engagement.Summary),
		Scope:             engagementScopeToObject(engagement.Scope),
		StartDate:         types.StringValue(engagement.StartDate),
		EndDate:           nullableString(engagement.EndDate),
		Status:            nullableString(engagement.Status),
		Tags:              mapToTerraform(engagement.Tags),
		EngagementLeadKey: nullableString(engagement.EngagementLeadKey),
		Provider:          nullableString(engagement.Provider),
		Attachments:       stringSliceToList(engagement.Attachments),
		Reporter:          nullableString(engagement.Reporter),
	}
}

func engagementScopeFromObject(scope types.Object) *engagementScopeBody {
	if scope.IsNull() || scope.IsUnknown() {
		return nil
	}
	attrs := scope.Attributes()
	applicationKeys := types.ListNull(types.StringType)
	if value, ok := attrs["application_keys"]; ok {
		if listValue, ok := value.(types.List); ok {
			applicationKeys = listValue
		}
	}
	repositoryKeys := types.ListNull(types.StringType)
	if value, ok := attrs["repository_keys"]; ok {
		if listValue, ok := value.(types.List); ok {
			repositoryKeys = listValue
		}
	}

	return &engagementScopeBody{
		ApplicationKeys: listToStrings(context.Background(), applicationKeys),
		RepositoryKeys:  listToStrings(context.Background(), repositoryKeys),
	}
}

func engagementScopeToObject(scope *engagementScopeBody) types.Object {
	objectType := map[string]attr.Type{
		"application_keys": types.ListType{ElemType: types.StringType},
		"repository_keys":  types.ListType{ElemType: types.StringType},
	}
	if scope == nil {
		return types.ObjectNull(objectType)
	}
	value, diags := types.ObjectValue(objectType, map[string]attr.Value{
		"application_keys": stringSliceToList(scope.ApplicationKeys),
		"repository_keys":  stringSliceToList(scope.RepositoryKeys),
	})
	if diags.HasError() {
		return types.ObjectNull(objectType)
	}
	return value
}

func mergeObject(plan, state types.Object) types.Object {
	if plan.IsNull() || plan.IsUnknown() {
		return state
	}
	return plan
}
