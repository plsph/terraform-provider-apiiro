package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &applicationsResource{}
	_ resource.ResourceWithConfigure   = &applicationsResource{}
	_ resource.ResourceWithImportState = &applicationsResource{}
)

type applicationsResource struct {
	client *Client
}

type applicationsResourceModel struct {
	ID                     types.String `tfsdk:"id"`
	Name                   types.String `tfsdk:"name"`
	ApplicationType        types.String `tfsdk:"application_type"`
	ApplicationTypeOther   types.String `tfsdk:"application_type_other"`
	BusinessImpact         types.String `tfsdk:"business_impact"`
	BusinessUnit           types.String `tfsdk:"business_unit"`
	ComplianceRequirements types.List   `tfsdk:"compliance_requirements"`
	DeploymentLocation     types.String `tfsdk:"deployment_location"`
	Description            types.String `tfsdk:"description"`
	EstimatedRevenue       types.String `tfsdk:"estimated_revenue"`
	EstimatedUsersNumber   types.String `tfsdk:"estimated_users_number"`
	IsInternetFacing       types.Bool   `tfsdk:"is_internet_facing"`
	ApiGatewayUrls         types.List   `tfsdk:"api_gateway_urls"`
	ApiGroupKeys           types.List   `tfsdk:"api_group_keys"`
	EntryPoints            types.List   `tfsdk:"entry_points"`
	PointsOfContact        types.List   `tfsdk:"points_of_contact"`
	ProjectUrls            types.List   `tfsdk:"project_urls"`
	RepositoryUrls         types.List   `tfsdk:"repository_urls"`
	Tags                   types.Map    `tfsdk:"tags"`
	RiskScore              types.Int64  `tfsdk:"risk_score"`
}

func NewApplicationsResource() resource.Resource {
	return &applicationsResource{}
}

func (r *applicationsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_applications"
}

func (r *applicationsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages application configurations in Apiiro.",
		Attributes: map[string]schema.Attribute{
			"id":                      schema.StringAttribute{Computed: true},
			"name":                    schema.StringAttribute{Optional: true, Computed: true},
			"application_type":        schema.StringAttribute{Optional: true, Computed: true},
			"application_type_other":  schema.StringAttribute{Optional: true, Computed: true},
			"business_impact":         schema.StringAttribute{Optional: true, Computed: true},
			"business_unit":           schema.StringAttribute{Optional: true, Computed: true},
			"compliance_requirements": schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType},
			"deployment_location":     schema.StringAttribute{Optional: true, Computed: true},
			"description":             schema.StringAttribute{Optional: true, Computed: true},
			"estimated_revenue":       schema.StringAttribute{Optional: true, Computed: true},
			"estimated_users_number":  schema.StringAttribute{Optional: true, Computed: true},
			"is_internet_facing":      schema.BoolAttribute{Optional: true, Computed: true},
			"api_gateway_urls":        schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType},
			"api_group_keys":          schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType},
			"entry_points":            schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType},
			"points_of_contact": schema.ListNestedAttribute{Optional: true, Computed: true, NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
				"identity": schema.StringAttribute{Optional: true, Computed: true},
				"title":    schema.StringAttribute{Optional: true, Computed: true},
			}}},
			"project_urls":    schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType},
			"repository_urls": schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType},
			"tags":            schema.MapAttribute{Optional: true, Computed: true, ElementType: types.StringType},
			"risk_score":      schema.Int64Attribute{Computed: true},
		},
	}
}

func (r *applicationsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *applicationsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "applications create requested")
	var plan applicationsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := applicationBody{
		ApiGatewayUrls:         listToStrings(ctx, plan.ApiGatewayUrls),
		ApiGroupKeys:           listToStrings(ctx, plan.ApiGroupKeys),
		ApplicationType:        stringPtr(plan.ApplicationType),
		ApplicationTypeOther:   stringPtr(plan.ApplicationTypeOther),
		BusinessImpact:         stringPtr(plan.BusinessImpact),
		BusinessUnit:           stringPtr(plan.BusinessUnit),
		ComplianceRequirements: listToStrings(ctx, plan.ComplianceRequirements),
		DeploymentLocation:     stringPtr(plan.DeploymentLocation),
		Description:            stringPtr(plan.Description),
		EntryPoints:            listToStrings(ctx, plan.EntryPoints),
		EstimatedRevenue:       stringPtr(plan.EstimatedRevenue),
		EstimatedUsersNumber:   stringPtr(plan.EstimatedUsersNumber),
		IsInternetFacing:       plan.IsInternetFacing.ValueBool(),
		Name:                   stringPtr(plan.Name),
		PointsOfContact:        pointOfContactListToBodies(ctx, plan.PointsOfContact),
		ProjectUrls:            listToStrings(ctx, plan.ProjectUrls),
		RepositoryUrls:         listToStrings(ctx, plan.RepositoryUrls),
		Tags:                   tagsMapToBodies(mapFromTerraform(ctx, plan.Tags)),
	}

	key, err := r.client.createApplication(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create Application", err.Error())
		return
	}

	fresh, err := r.client.getApplication(ctx, key)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Application", err.Error())
		return
	}
	state := r.toState(ctx, fresh)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *applicationsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state applicationsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "applications read requested", map[string]any{"id": state.ID.ValueString()})

	fresh, err := r.client.getApplication(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Application", err.Error())
		return
	}
	if fresh == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, r.toState(ctx, fresh))...)
}

func (r *applicationsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan applicationsResourceModel
	var state applicationsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "applications update requested", map[string]any{"id": state.ID.ValueString()})

	body := applicationBody{
		Key:                    stringPtr(state.ID),
		ApiGatewayUrls:         listToStrings(ctx, mergeList(plan.ApiGatewayUrls, state.ApiGatewayUrls)),
		ApiGroupKeys:           listToStrings(ctx, mergeList(plan.ApiGroupKeys, state.ApiGroupKeys)),
		ApplicationType:        stringPtr(mergeString(plan.ApplicationType, state.ApplicationType)),
		ApplicationTypeOther:   stringPtr(mergeString(plan.ApplicationTypeOther, state.ApplicationTypeOther)),
		BusinessImpact:         stringPtr(mergeString(plan.BusinessImpact, state.BusinessImpact)),
		BusinessUnit:           stringPtr(mergeString(plan.BusinessUnit, state.BusinessUnit)),
		ComplianceRequirements: listToStrings(ctx, mergeList(plan.ComplianceRequirements, state.ComplianceRequirements)),
		DeploymentLocation:     stringPtr(mergeString(plan.DeploymentLocation, state.DeploymentLocation)),
		Description:            stringPtr(mergeString(plan.Description, state.Description)),
		EntryPoints:            listToStrings(ctx, mergeList(plan.EntryPoints, state.EntryPoints)),
		EstimatedRevenue:       stringPtr(mergeString(plan.EstimatedRevenue, state.EstimatedRevenue)),
		EstimatedUsersNumber:   stringPtr(mergeString(plan.EstimatedUsersNumber, state.EstimatedUsersNumber)),
		IsInternetFacing:       mergeBool(plan.IsInternetFacing, state.IsInternetFacing).ValueBool(),
		Name:                   stringPtr(mergeString(plan.Name, state.Name)),
		PointsOfContact:        pointOfContactListToBodies(ctx, mergeList(plan.PointsOfContact, state.PointsOfContact)),
		ProjectUrls:            listToStrings(ctx, mergeList(plan.ProjectUrls, state.ProjectUrls)),
		RepositoryUrls:         listToStrings(ctx, mergeList(plan.RepositoryUrls, state.RepositoryUrls)),
		Tags:                   tagsMapToBodies(mapFromTerraform(ctx, mergeMap(plan.Tags, state.Tags))),
	}

	if err := r.client.updateApplication(ctx, state.ID.ValueString(), body); err != nil {
		resp.Diagnostics.AddError("Unable to Update Application", err.Error())
		return
	}

	fresh, err := r.client.getApplication(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Application", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, r.toState(ctx, fresh))...)
}

func (r *applicationsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state applicationsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "applications delete requested", map[string]any{"id": state.ID.ValueString()})
	if err := r.client.deleteApplication(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Unable to Delete Application", err.Error())
	}
}

func (r *applicationsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *applicationsResource) toState(ctx context.Context, app *applicationBody) *applicationsResourceModel {
	if app == nil {
		return nil
	}
	return &applicationsResourceModel{
		ID:                     nullableString(app.Key),
		Name:                   nullableString(app.Name),
		ApplicationType:        nullableString(app.ApplicationType),
		ApplicationTypeOther:   nullableString(app.ApplicationTypeOther),
		BusinessImpact:         nullableString(app.BusinessImpact),
		BusinessUnit:           nullableString(app.BusinessUnit),
		ComplianceRequirements: stringSliceToList(app.ComplianceRequirements),
		DeploymentLocation:     nullableString(app.DeploymentLocation),
		Description:            nullableString(app.Description),
		EstimatedRevenue:       nullableString(app.EstimatedRevenue),
		EstimatedUsersNumber:   nullableString(app.EstimatedUsersNumber),
		IsInternetFacing:       types.BoolValue(app.IsInternetFacing),
		ApiGatewayUrls:         stringSliceToList(app.ApiGatewayUrls),
		ApiGroupKeys:           stringSliceToList(app.ApiGroupKeys),
		EntryPoints:            stringSliceToList(app.EntryPoints),
		PointsOfContact:        pointOfContactBodiesToList(app.PointsOfContact),
		ProjectUrls:            stringSliceToList(app.ProjectUrls),
		RepositoryUrls:         stringSliceToList(app.RepositoryUrls),
		Tags:                   mapToTerraform(tagsBodiesToMap(app.Tags)),
		RiskScore:              types.Int64Value(int64(valueOrZero(app.RiskScore))),
	}
}

func listToStrings(ctx context.Context, list types.List) []string {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}
	var values []string
	_ = list.ElementsAs(ctx, &values, false)
	clean := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			continue
		}
		clean = append(clean, value)
	}
	return clean
}

func stringPtr(value types.String) *string {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	text := value.ValueString()
	if strings.TrimSpace(text) == "" {
		return nil
	}
	return &text
}

func valueOrZero(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}
