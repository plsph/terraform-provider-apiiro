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
	_ datasource.DataSource              = &applicationProfilesDataSource{}
	_ datasource.DataSourceWithConfigure = &applicationProfilesDataSource{}
)

type applicationProfilesDataSource struct {
	client *Client
}

type applicationProfilesDataSourceModel struct {
	ID                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	ApplicationKey     types.String `tfsdk:"application_key"`
	ApplicationType    types.String `tfsdk:"application_type"`
	RiskLevel          types.String `tfsdk:"risk_level"`
	InternetExposed    types.Bool   `tfsdk:"internet_exposed"`
	ApplicationProfile types.List   `tfsdk:"application_profiles"`
}

func NewApplicationProfilesDataSource() datasource.DataSource {
	return &applicationProfilesDataSource{}
}

func (d *applicationProfilesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_applicationprofiles"
}

func (d *applicationProfilesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches application profiles from Apiiro with optional filtering by key, name, type, risk level, and internet exposure.",
		Attributes: map[string]schema.Attribute{
			"id":               schema.StringAttribute{Computed: true},
			"name":             schema.StringAttribute{Optional: true},
			"application_key":  schema.StringAttribute{Optional: true},
			"application_type": schema.StringAttribute{Optional: true},
			"risk_level":       schema.StringAttribute{Optional: true},
			"internet_exposed": schema.BoolAttribute{Optional: true},
			"application_profiles": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
					"id":                  schema.StringAttribute{Computed: true},
					"key":                 schema.StringAttribute{Computed: true},
					"name":                schema.StringAttribute{Computed: true},
					"application_type":    schema.StringAttribute{Computed: true},
					"business_impact":     schema.StringAttribute{Computed: true},
					"business_unit":       schema.StringAttribute{Computed: true},
					"description":         schema.StringAttribute{Computed: true},
					"risk_level":          schema.StringAttribute{Computed: true},
					"risk_score":          schema.Int64Attribute{Computed: true},
					"is_active":           schema.BoolAttribute{Computed: true},
					"is_deployed":         schema.BoolAttribute{Computed: true},
					"is_internet_exposed": schema.BoolAttribute{Computed: true},
					"is_public":           schema.BoolAttribute{Computed: true},
					"is_user_facing":      schema.BoolAttribute{Computed: true},
					"languages":           schema.ListAttribute{Computed: true, ElementType: types.StringType},
					"licenses":            schema.ListAttribute{Computed: true, ElementType: types.StringType},
					"tags":                schema.MapAttribute{Computed: true, ElementType: types.StringType},
				}},
			},
		},
	}
}

func (d *applicationProfilesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *applicationProfilesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state applicationProfilesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	nameFilter := strings.TrimSpace(state.Name.ValueString())
	keyFilter := strings.TrimSpace(state.ApplicationKey.ValueString())
	typeFilter := strings.TrimSpace(state.ApplicationType.ValueString())
	riskLevelFilter := strings.TrimSpace(state.RiskLevel.ValueString())
	tflog.Debug(ctx, "application profiles data source read requested", map[string]any{"name": nameFilter, "application_key": keyFilter, "application_type": typeFilter, "risk_level": riskLevelFilter})

	profiles := make([]applicationProfileBody, 0)
	if keyFilter != "" {
		profile, err := d.client.getApplicationProfile(ctx, keyFilter)
		if err != nil {
			resp.Diagnostics.AddError("Unable to Read Application Profile", err.Error())
			return
		}
		if profile != nil {
			profiles = append(profiles, *profile)
		}
	} else {
		apiFilters := map[string][]string{}
		if nameFilter != "" {
			apiFilters["Name"] = []string{nameFilter}
		}

		list, err := d.client.listApplicationProfiles(ctx, apiFilters)
		if err != nil {
			resp.Diagnostics.AddError("Unable to Read Application Profiles", err.Error())
			return
		}
		profiles = list
	}

	values := make([]attr.Value, 0, len(profiles))
	for _, profile := range profiles {
		if nameFilter != "" && !strings.Contains(strings.ToLower(valueOrEmpty(profile.Name)), strings.ToLower(nameFilter)) {
			continue
		}
		if typeFilter != "" && !strings.EqualFold(valueOrEmpty(profile.ApplicationType), typeFilter) {
			continue
		}
		if riskLevelFilter != "" && !strings.EqualFold(valueOrEmpty(profile.RiskLevel), riskLevelFilter) {
			continue
		}
		if !state.InternetExposed.IsNull() && !state.InternetExposed.IsUnknown() && state.InternetExposed.ValueBool() != profile.IsInternetExposed {
			continue
		}

		obj, diags := types.ObjectValue(applicationProfilesDataSourceObjectType(), map[string]attr.Value{
			"id":                  nullableString(profile.Key),
			"key":                 nullableString(profile.Key),
			"name":                nullableString(profile.Name),
			"application_type":    nullableString(profile.ApplicationType),
			"business_impact":     nullableString(profile.BusinessImpact),
			"business_unit":       nullableString(profile.BusinessUnit),
			"description":         nullableString(profile.Description),
			"risk_level":          nullableString(profile.RiskLevel),
			"risk_score":          nullableInt(profile.RiskScore),
			"is_active":           types.BoolValue(profile.IsActive),
			"is_deployed":         types.BoolValue(profile.IsDeployed),
			"is_internet_exposed": types.BoolValue(profile.IsInternetExposed),
			"is_public":           types.BoolValue(profile.IsPublic),
			"is_user_facing":      types.BoolValue(profile.IsUserFacing),
			"languages":           stringSliceToList(profile.Languages),
			"licenses":            stringSliceToList(profile.Licenses),
			"tags":                mapToTerraform(applicationProfileTagsToMap(profile.Tags)),
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		values = append(values, obj)
	}

	profileList, diags := types.ListValue(types.ObjectType{AttrTypes: applicationProfilesDataSourceObjectType()}, values)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.ID = types.StringValue(fmt.Sprintf("applicationprofiles:%s:%s:%s:%s:%t", nameFilter, keyFilter, typeFilter, riskLevelFilter, state.InternetExposed.ValueBool()))
	state.ApplicationProfile = profileList
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func applicationProfilesDataSourceObjectType() map[string]attr.Type {
	return map[string]attr.Type{
		"id":                  types.StringType,
		"key":                 types.StringType,
		"name":                types.StringType,
		"application_type":    types.StringType,
		"business_impact":     types.StringType,
		"business_unit":       types.StringType,
		"description":         types.StringType,
		"risk_level":          types.StringType,
		"risk_score":          types.Int64Type,
		"is_active":           types.BoolType,
		"is_deployed":         types.BoolType,
		"is_internet_exposed": types.BoolType,
		"is_public":           types.BoolType,
		"is_user_facing":      types.BoolType,
		"languages":           types.ListType{ElemType: types.StringType},
		"licenses":            types.ListType{ElemType: types.StringType},
		"tags":                types.MapType{ElemType: types.StringType},
	}
}

func applicationProfileTagsToMap(values []tagBodyResponse) map[string]string {
	if len(values) == 0 {
		return map[string]string{}
	}
	result := make(map[string]string, len(values))
	for _, value := range values {
		name := strings.TrimSpace(valueOrEmpty(value.Name))
		if name == "" {
			continue
		}
		result[name] = valueOrEmpty(value.Value)
	}
	return result
}
