package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/disc/terraform-provider-pritunl/internal/pritunl"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &HostDataSource{}

func NewHostDataSource() datasource.DataSource {
	return &HostDataSource{}
}

type HostDataSource struct {
	client pritunl.Client
}

type HostDataSourceModel struct {
	ID                types.String `tfsdk:"id"`
	Hostname          types.String `tfsdk:"hostname"`
	Name              types.String `tfsdk:"name"`
	PublicAddr        types.String `tfsdk:"public_addr"`
	PublicAddr6       types.String `tfsdk:"public_addr6"`
	RoutedSubnet6     types.String `tfsdk:"routed_subnet6"`
	RoutedSubnet6WG   types.String `tfsdk:"routed_subnet6_wg"`
	LocalAddr         types.String `tfsdk:"local_addr"`
	LocalAddr6        types.String `tfsdk:"local_addr6"`
	AvailabilityGroup types.String `tfsdk:"availability_group"`
	LinkAddr          types.String `tfsdk:"link_addr"`
	SyncAddress       types.String `tfsdk:"sync_address"`
	Status            types.String `tfsdk:"status"`
}

func (d *HostDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_host"
}

func (d *HostDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to get information about the Pritunl hosts.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Host identifier",
				Computed:            true,
			},
			"hostname": schema.StringAttribute{
				MarkdownDescription: "Hostname",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of host",
				Computed:            true,
			},
			"public_addr": schema.StringAttribute{
				MarkdownDescription: "Public IP address or domain name of the host",
				Computed:            true,
			},
			"public_addr6": schema.StringAttribute{
				MarkdownDescription: "Public IPv6 address or domain name of the host",
				Computed:            true,
			},
			"routed_subnet6": schema.StringAttribute{
				MarkdownDescription: "IPv6 subnet that is routed to the host",
				Computed:            true,
			},
			"routed_subnet6_wg": schema.StringAttribute{
				MarkdownDescription: "IPv6 WG subnet that is routed to the host",
				Computed:            true,
			},
			"local_addr": schema.StringAttribute{
				MarkdownDescription: "Local network address for server",
				Computed:            true,
			},
			"local_addr6": schema.StringAttribute{
				MarkdownDescription: "Local IPv6 network address for server",
				Computed:            true,
			},
			"availability_group": schema.StringAttribute{
				MarkdownDescription: "Availability group for host. Replicated servers will only be replicated to a group of hosts in the same availability group",
				Computed:            true,
			},
			"link_addr": schema.StringAttribute{
				MarkdownDescription: "IP address or domain used when linked servers connect to a linked server on this host",
				Computed:            true,
			},
			"sync_address": schema.StringAttribute{
				MarkdownDescription: "IP address or domain used by users when syncing configuration. This is needed when using a load balancer.",
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Status of host",
				Computed:            true,
			},
		},
	}
}

func (d *HostDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *HostDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data HostDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	hostname := data.Hostname.ValueString()
	filterFunction := func(host pritunl.Host) bool {
		return host.Hostname == hostname
	}

	host, err := d.filterHosts(filterFunction)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Could not find host with hostname %s: %s", hostname, err))
		return
	}

	data.ID = types.StringValue(host.ID)
	data.Name = types.StringValue(host.Name)
	data.Hostname = types.StringValue(host.Hostname)
	data.PublicAddr = types.StringValue(host.PublicAddr)
	data.PublicAddr6 = types.StringValue(host.PublicAddr6)
	data.RoutedSubnet6 = types.StringValue(host.RoutedSubnet6)
	data.RoutedSubnet6WG = types.StringValue(host.RoutedSubnet6WG)
	data.LocalAddr = types.StringValue(host.LocalAddr)
	data.LocalAddr6 = types.StringValue(host.LocalAddr6)
	data.LinkAddr = types.StringValue(host.LinkAddr)
	data.SyncAddress = types.StringValue(host.SyncAddress)
	data.AvailabilityGroup = types.StringValue(host.AvailabilityGroup)
	data.Status = types.StringValue(host.Status)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *HostDataSource) filterHosts(test func(host pritunl.Host) bool) (pritunl.Host, error) {
	hosts, err := d.client.GetHosts()

	if err != nil {
		return pritunl.Host{}, err
	}

	for _, host := range hosts {
		if test(host) {
			return host, nil
		}
	}

	return pritunl.Host{}, errors.New("could not find a host with specified parameters")
}
