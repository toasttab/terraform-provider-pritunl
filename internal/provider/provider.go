package provider

import (
	"context"
	"os"

	"github.com/disc/terraform-provider-pritunl/internal/pritunl"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &pritunlProvider{}

type pritunlProvider struct {
	version string
}

type pritunlProviderModel struct {
	URL             types.String `tfsdk:"url"`
	Token           types.String `tfsdk:"token"`
	Secret          types.String `tfsdk:"secret"`
	Insecure        types.Bool   `tfsdk:"insecure"`
	ConnectionCheck types.Bool   `tfsdk:"connection_check"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &pritunlProvider{
			version: version,
		}
	}
}

func (p *pritunlProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "pritunl"
	resp.Version = p.version
}

func (p *pritunlProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Required:    true,
				Description: "The URL of the Pritunl server.",
			},
			"token": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "The API token for the Pritunl server.",
			},
			"secret": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "The API secret for the Pritunl server.",
			},
			"insecure": schema.BoolAttribute{
				Required:    true,
				Description: "Whether to skip TLS verification.",
			},
			"connection_check": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether to check the connection to the Pritunl server.",
			},
		},
	}
}

func (p *pritunlProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config pritunlProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.URL.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("url"),
			"Unknown Pritunl URL",
			"The provider cannot create the Pritunl API client as there is an unknown configuration value for the Pritunl URL. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the PRITUNL_URL environment variable.",
		)
	}

	if config.Token.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Unknown Pritunl Token",
			"The provider cannot create the Pritunl API client as there is an unknown configuration value for the Pritunl token. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the PRITUNL_TOKEN environment variable.",
		)
	}

	if config.Secret.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("secret"),
			"Unknown Pritunl Secret",
			"The provider cannot create the Pritunl API client as there is an unknown configuration value for the Pritunl secret. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the PRITUNL_SECRET environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	url := os.Getenv("PRITUNL_URL")
	token := os.Getenv("PRITUNL_TOKEN")
	secret := os.Getenv("PRITUNL_SECRET")
	insecure := os.Getenv("PRITUNL_INSECURE") == "true"
	connectionCheck := os.Getenv("PRITUNL_CONNECTION_CHECK") != "false"

	if !config.URL.IsNull() {
		url = config.URL.ValueString()
	}

	if !config.Token.IsNull() {
		token = config.Token.ValueString()
	}

	if !config.Secret.IsNull() {
		secret = config.Secret.ValueString()
	}

	if !config.Insecure.IsNull() {
		insecure = config.Insecure.ValueBool()
	}

	if !config.ConnectionCheck.IsNull() {
		connectionCheck = config.ConnectionCheck.ValueBool()
	}

	if url == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("url"),
			"Missing Pritunl URL",
			"The provider cannot create the Pritunl API client as there is a missing or empty value for the Pritunl URL. "+
				"Set the url value in the configuration or use the PRITUNL_URL environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if token == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Missing Pritunl Token",
			"The provider cannot create the Pritunl API client as there is a missing or empty value for the Pritunl token. "+
				"Set the token value in the configuration or use the PRITUNL_TOKEN environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if secret == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("secret"),
			"Missing Pritunl Secret",
			"The provider cannot create the Pritunl API client as there is a missing or empty value for the Pritunl secret. "+
				"Set the secret value in the configuration or use the PRITUNL_SECRET environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	client := pritunl.NewClient(url, token, secret, insecure)

	if connectionCheck {
		err := client.TestApiCall()
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Create Pritunl API Client",
				"An unexpected error occurred when creating the Pritunl API client. "+
					"If the error is not clear, please contact the provider developers.\n\n"+
					"Pritunl Client Error: "+err.Error(),
			)
			return
		}
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *pritunlProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewOrganizationResource,
		NewServerResource,
		NewUserResource,
	}
}

func (p *pritunlProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewHostDataSource,
		NewHostsDataSource,
	}
}
