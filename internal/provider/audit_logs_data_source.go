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
	_ datasource.DataSource              = &auditLogsDataSource{}
	_ datasource.DataSourceWithConfigure = &auditLogsDataSource{}
)

type auditLogsDataSource struct {
	client *Client
}

type auditLogsDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	EventType types.String `tfsdk:"event_type"`
	Status    types.String `tfsdk:"status"`
	User      types.String `tfsdk:"user"`
	UserType  types.String `tfsdk:"user_type"`
	Logs      types.List   `tfsdk:"logs"`
}

func NewAuditLogsDataSource() datasource.DataSource {
	return &auditLogsDataSource{}
}

func (d *auditLogsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_auditlogs"
}

func (d *auditLogsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches audit logs from Apiiro with optional filtering by event type, status, user, or user type.",
		Attributes: map[string]schema.Attribute{
			"id":         schema.StringAttribute{Computed: true},
			"event_type": schema.StringAttribute{Optional: true},
			"status":     schema.StringAttribute{Optional: true},
			"user":       schema.StringAttribute{Optional: true},
			"user_type":  schema.StringAttribute{Optional: true},
			"logs": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
					"id":                   schema.StringAttribute{Computed: true},
					"event_type":           schema.StringAttribute{Computed: true},
					"event_description":    schema.StringAttribute{Computed: true},
					"status":               schema.StringAttribute{Computed: true},
					"time":                 schema.StringAttribute{Computed: true},
					"user":                 schema.StringAttribute{Computed: true},
					"user_type":            schema.StringAttribute{Computed: true},
					"source_ip_address":    schema.StringAttribute{Computed: true},
					"impacted_entity_id":   schema.StringAttribute{Computed: true},
					"impacted_entity_type": schema.StringAttribute{Computed: true},
					"impacted_entity_url":  schema.StringAttribute{Computed: true},
					"error_description":    schema.StringAttribute{Computed: true},
					"timezone":             schema.StringAttribute{Computed: true},
					"user_agent":           schema.StringAttribute{Computed: true},
				}},
			},
		},
	}
}

func (d *auditLogsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *auditLogsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state auditLogsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	eventTypeFilter := strings.TrimSpace(state.EventType.ValueString())
	statusFilter := strings.TrimSpace(state.Status.ValueString())
	userFilter := strings.TrimSpace(state.User.ValueString())
	userTypeFilter := strings.TrimSpace(state.UserType.ValueString())
	tflog.Debug(ctx, "audit logs data source read requested", map[string]any{"event_type": eventTypeFilter, "status": statusFilter, "user": userFilter, "user_type": userTypeFilter})

	apiFilters := map[string][]string{}
	if eventTypeFilter != "" {
		apiFilters["EventType"] = []string{eventTypeFilter}
	}
	if statusFilter != "" {
		apiFilters["Status"] = []string{statusFilter}
	}
	if userFilter != "" {
		apiFilters["User"] = []string{userFilter}
	}
	if userTypeFilter != "" {
		apiFilters["UserType"] = []string{userTypeFilter}
	}

	logs, err := d.client.listAuditLogs(ctx, apiFilters)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Audit Logs", err.Error())
		return
	}

	values := make([]attr.Value, 0, len(logs))
	for _, log := range logs {
		if eventTypeFilter != "" && !strings.EqualFold(valueOrEmpty(log.EventType), eventTypeFilter) {
			continue
		}
		if statusFilter != "" && !strings.EqualFold(valueOrEmpty(log.Status), statusFilter) {
			continue
		}
		if userFilter != "" && !strings.Contains(strings.ToLower(valueOrEmpty(log.User)), strings.ToLower(userFilter)) {
			continue
		}
		if userTypeFilter != "" && !strings.EqualFold(valueOrEmpty(log.UserType), userTypeFilter) {
			continue
		}

		obj, diags := types.ObjectValue(auditLogsDataSourceObjectType(), map[string]attr.Value{
			"id":                   nullableString(log.Key),
			"event_type":           nullableString(log.EventType),
			"event_description":    nullableString(log.EventDescription),
			"status":               nullableString(log.Status),
			"time":                 nullableString(log.Time),
			"user":                 nullableString(log.User),
			"user_type":            nullableString(log.UserType),
			"source_ip_address":    nullableString(log.SourceIPAddress),
			"impacted_entity_id":   nullableString(log.ImpactedEntityID),
			"impacted_entity_type": nullableString(log.ImpactedEntityType),
			"impacted_entity_url":  nullableString(log.ImpactedEntityURL),
			"error_description":    nullableString(log.ErrorDescription),
			"timezone":             nullableString(log.Timezone),
			"user_agent":           nullableString(log.UserAgent),
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		values = append(values, obj)
	}

	logList, diags := types.ListValue(types.ObjectType{AttrTypes: auditLogsDataSourceObjectType()}, values)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.ID = types.StringValue(fmt.Sprintf("auditlogs:%s:%s:%s:%s", eventTypeFilter, statusFilter, userFilter, userTypeFilter))
	state.Logs = logList
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func auditLogsDataSourceObjectType() map[string]attr.Type {
	return map[string]attr.Type{
		"id":                   types.StringType,
		"event_type":           types.StringType,
		"event_description":    types.StringType,
		"status":               types.StringType,
		"time":                 types.StringType,
		"user":                 types.StringType,
		"user_type":            types.StringType,
		"source_ip_address":    types.StringType,
		"impacted_entity_id":   types.StringType,
		"impacted_entity_type": types.StringType,
		"impacted_entity_url":  types.StringType,
		"error_description":    types.StringType,
		"timezone":             types.StringType,
		"user_agent":           types.StringType,
	}
}
