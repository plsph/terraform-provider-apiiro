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
	_ datasource.DataSource              = &engagementsDataSource{}
	_ datasource.DataSourceWithConfigure = &engagementsDataSource{}
)

type engagementsDataSource struct {
	client *Client
}

type engagementsDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Type        types.String `tfsdk:"type"`
	Status      types.String `tfsdk:"status"`
	LeadKey     types.String `tfsdk:"lead_key"`
	Engagements types.List   `tfsdk:"engagements"`
}

func NewEngagementsDataSource() datasource.DataSource {
	return &engagementsDataSource{}
}

func (d *engagementsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_engagements"
}

func (d *engagementsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches engagements from Apiiro with optional filtering by type, status, or lead key.",
		Attributes: map[string]schema.Attribute{
			"id":       schema.StringAttribute{Computed: true},
			"name":     schema.StringAttribute{Optional: true},
			"type":     schema.StringAttribute{Optional: true},
			"status":   schema.StringAttribute{Optional: true},
			"lead_key": schema.StringAttribute{Optional: true},
			"engagements": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
					"id":           schema.StringAttribute{Computed: true},
					"key":          schema.StringAttribute{Computed: true},
					"name":         schema.StringAttribute{Computed: true},
					"status":       schema.StringAttribute{Computed: true},
					"type":         schema.StringAttribute{Computed: true},
					"external_url": schema.StringAttribute{Computed: true},
					"provider":     schema.StringAttribute{Computed: true},
					"reporter":     schema.StringAttribute{Computed: true},
					"summary":      schema.StringAttribute{Computed: true},
					"start_date":   schema.StringAttribute{Computed: true},
					"end_date":     schema.StringAttribute{Computed: true},
					"updated_at":   schema.StringAttribute{Computed: true},
				}},
			},
		},
	}
}

func (d *engagementsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *engagementsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state engagementsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	typeFilter := strings.TrimSpace(state.Type.ValueString())
	statusFilter := strings.TrimSpace(state.Status.ValueString())
	leadFilter := strings.TrimSpace(state.LeadKey.ValueString())
	tflog.Debug(ctx, "engagements data source read requested", map[string]any{"type": typeFilter, "status": statusFilter, "lead_key": leadFilter})

	engagements, err := d.client.listEngagements(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Engagements", err.Error())
		return
	}

	values := make([]attr.Value, 0, len(engagements))
	for _, engagement := range engagements {
		if typeFilter != "" && !strings.EqualFold(valueOrEmpty(engagement.Type), typeFilter) {
			continue
		}
		if statusFilter != "" && !strings.EqualFold(valueOrEmpty(engagement.Status), statusFilter) {
			continue
		}
		if leadFilter != "" && !strings.EqualFold(valueOrEmpty(engagement.EngagementLeadKey), leadFilter) {
			continue
		}

		obj, diags := types.ObjectValue(engagementsDataSourceObjectType(), map[string]attr.Value{
			"id":           nullableString(engagement.Key),
			"key":          nullableString(engagement.Key),
			"name":         nullableString(engagement.Name),
			"status":       nullableString(engagement.Status),
			"type":         nullableString(engagement.Type),
			"external_url": nullableString(engagement.ExternalURL),
			"provider":     nullableString(engagement.Provider),
			"reporter":     nullableString(engagement.Reporter),
			"summary":      nullableString(engagement.Summary),
			"start_date":   types.StringValue(engagement.StartDate),
			"end_date":     nullableString(engagement.EndDate),
			"updated_at":   nullableString(engagement.UpdatedAt),
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		values = append(values, obj)
	}

	engagementList, diags := types.ListValue(types.ObjectType{AttrTypes: engagementsDataSourceObjectType()}, values)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.ID = types.StringValue(fmt.Sprintf("engagements:%s:%s:%s", typeFilter, statusFilter, leadFilter))
	state.Engagements = engagementList
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func engagementsDataSourceObjectType() map[string]attr.Type {
	return map[string]attr.Type{
		"id":           types.StringType,
		"key":          types.StringType,
		"name":         types.StringType,
		"status":       types.StringType,
		"type":         types.StringType,
		"external_url": types.StringType,
		"provider":     types.StringType,
		"reporter":     types.StringType,
		"summary":      types.StringType,
		"start_date":   types.StringType,
		"end_date":     types.StringType,
		"updated_at":   types.StringType,
	}
}
