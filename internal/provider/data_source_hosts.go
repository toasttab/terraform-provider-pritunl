package provider

import (
	"context"
	"fmt"

	"github.com/disc/terraform-provider-pritunl/internal/pritunl"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure HostsDataSource implements datasource.DataSource interface at compile time
var _ datasource.DataSource = &HostsDataSource{}

func NewHostsDataSource() datasource.DataSource {
	return &HostsDataSource{}
}

type HostsDataSource struct {
	client pritunl.Client
}

type HostsDataSourceModel struct {
	Hosts []HostDataSourceModel `tfsdk:"hosts"`
}

func (d *HostsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hosts"
}

func (d *HostsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to get a list of the Pritunl hosts.",

		Attributes: map[string]schema.Attribute{
			"hosts": schema.ListNestedAttribute{
				MarkdownDescription: "A list of the Pritunl hosts resources.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: hostAttributes(),
				},
			},
		},
	}
}

func (d *HostsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(pritunl.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected pritunl.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *HostsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data HostsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	hosts, err := d.client.GetHosts()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Could not retrieve hosts: %s", err))
		return
	}

	var hostModels []HostDataSourceModel
	for _, host := range hosts {
		hostModel := HostDataSourceModel{
			ID:                types.StringValue(host.ID),
			Hostname:          types.StringValue(host.Hostname),
			Name:              types.StringValue(host.Name),
			PublicAddr:        types.StringValue(host.PublicAddr),
			PublicAddr6:       types.StringValue(host.PublicAddr6),
			RoutedSubnet6:     types.StringValue(host.RoutedSubnet6),
			RoutedSubnet6WG:   types.StringValue(host.RoutedSubnet6WG),
			LocalAddr:         types.StringValue(host.LocalAddr),
			LocalAddr6:        types.StringValue(host.LocalAddr6),
			AvailabilityGroup: types.StringValue(host.AvailabilityGroup),
			LinkAddr:          types.StringValue(host.LinkAddr),
			SyncAddress:       types.StringValue(host.SyncAddress),
			Status:            types.StringValue(host.Status),
		}
		hostModels = append(hostModels, hostModel)
	}

	data.Hosts = hostModels

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
