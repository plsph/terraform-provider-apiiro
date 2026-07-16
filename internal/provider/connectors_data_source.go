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
	_ datasource.DataSource              = &connectorsDataSource{}
	_ datasource.DataSourceWithConfigure = &connectorsDataSource{}
)

type connectorsDataSource struct {
	client *Client
}

type connectorsDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	ConnectorID types.String `tfsdk:"connector_id"`
	Provider    types.String `tfsdk:"provider_name"`
	Connectors  types.List   `tfsdk:"connectors"`
}

func NewConnectorsDataSource() datasource.DataSource {
	return &connectorsDataSource{}
}

func (d *connectorsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connectors"
}

func (d *connectorsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches connectors from Apiiro with optional filtering by connector ID or provider name.",
		Attributes: map[string]schema.Attribute{
			"id":            schema.StringAttribute{Computed: true},
			"connector_id":  schema.StringAttribute{Optional: true},
			"provider_name": schema.StringAttribute{Optional: true},
			"connectors": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
					"id":                    schema.StringAttribute{Computed: true},
					"provider":              schema.StringAttribute{Computed: true},
					"url":                   schema.StringAttribute{Computed: true},
					"token_expiration_date": schema.StringAttribute{Computed: true},
				}},
			},
		},
	}
}

func (d *connectorsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *connectorsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state connectorsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	idFilter := strings.TrimSpace(state.ConnectorID.ValueString())
	providerFilter := strings.TrimSpace(state.Provider.ValueString())
	tflog.Debug(ctx, "connectors data source read requested", map[string]any{"connector_id": idFilter, "provider_name": providerFilter})

	connectors := make([]connectorBody, 0)
	if idFilter != "" {
		connector, err := d.client.getConnector(ctx, idFilter)
		if err != nil {
			resp.Diagnostics.AddError("Unable to Read Connector", err.Error())
			return
		}
		if connector != nil {
			connectors = append(connectors, *connector)
		}
	} else {
		list, err := d.client.listConnectors(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Unable to Read Connectors", err.Error())
			return
		}
		connectors = list
	}

	values := make([]attr.Value, 0, len(connectors))
	for _, connector := range connectors {
		if providerFilter != "" && !strings.EqualFold(valueOrEmpty(connector.Provider), providerFilter) {
			continue
		}

		obj, diags := types.ObjectValue(connectorsDataSourceObjectType(), map[string]attr.Value{
			"id":                    nullableString(connector.ID),
			"provider":              nullableString(connector.Provider),
			"url":                   nullableString(connector.URL),
			"token_expiration_date": nullableString(connector.TokenExpirationDate),
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		values = append(values, obj)
	}

	connectorList, diags := types.ListValue(types.ObjectType{AttrTypes: connectorsDataSourceObjectType()}, values)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.ID = types.StringValue(fmt.Sprintf("connectors:%s:%s", idFilter, providerFilter))
	state.Connectors = connectorList
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func connectorsDataSourceObjectType() map[string]attr.Type {
	return map[string]attr.Type{
		"id":                    types.StringType,
		"provider":              types.StringType,
		"url":                   types.StringType,
		"token_expiration_date": types.StringType,
	}
}
