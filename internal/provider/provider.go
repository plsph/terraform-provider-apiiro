package provider

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &ApiiroProvider{}

type ApiiroProvider struct {
	version string
}

type ApiiroProviderModel struct {
	BaseURL types.String `tfsdk:"base_url"`
	Token   types.String `tfsdk:"token"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ApiiroProvider{version: version}
	}
}

func (p *ApiiroProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "apiiro"
	resp.Version = p.version
}

func (p *ApiiroProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Terraform provider for Apiiro.",
		Attributes: map[string]schema.Attribute{
			"base_url": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Apiiro API base URL. Defaults to `https://app.apiiro.com` or environment variable `APIIRO_BASE_URL`.",
			},
			"token": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Apiiro access token. Can also be set via environment variable `APIIRO_TOKEN`.",
			},
		},
	}
}

func (p *ApiiroProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data ApiiroProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	baseURL := data.BaseURL.ValueString()
	if data.BaseURL.IsNull() || data.BaseURL.IsUnknown() || strings.TrimSpace(baseURL) == "" {
		if v := os.Getenv("APIIRO_BASE_URL"); strings.TrimSpace(v) != "" {
			baseURL = v
		} else {
			baseURL = "https://app.apiiro.com"
		}
	}

	token := data.Token.ValueString()
	if data.Token.IsNull() || data.Token.IsUnknown() || strings.TrimSpace(token) == "" {
		token = os.Getenv("APIIRO_TOKEN")
	}

	if strings.TrimSpace(token) == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Missing Apiiro Access Token",
			"The provider cannot create the Apiiro client because token is empty. Set `token` in configuration or `APIIRO_TOKEN` environment variable.",
		)
		return
	}

	client, err := NewClient(baseURL, token)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create Apiiro Client", fmt.Sprintf("Failed to initialize Apiiro client: %s", err))
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *ApiiroProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewScmRepositoriesResource,
	}
}

func (p *ApiiroProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewScmRepositoriesDataSource,
	}
}
