package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/disc/terraform-provider-pritunl/internal/pritunl"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ServerResource{}
var _ resource.ResourceWithImportState = &ServerResource{}

func NewServerResource() resource.Resource {
	return &ServerResource{}
}

type ServerResource struct {
	client pritunl.Client
}

type ServerResourceModel struct {
	ID                types.String   `tfsdk:"id"`
	Name              types.String   `tfsdk:"name"`
	Protocol          types.String   `tfsdk:"protocol"`
	Cipher            types.String   `tfsdk:"cipher"`
	Hash              types.String   `tfsdk:"hash"`
	Port              types.Int64    `tfsdk:"port"`
	Network           types.String   `tfsdk:"network"`
	WG                types.Bool     `tfsdk:"wg"`
	PortWG            types.Int64    `tfsdk:"port_wg"`
	NetworkWG         types.String   `tfsdk:"network_wg"`
	NetworkMode       types.String   `tfsdk:"network_mode"`
	NetworkStart      types.String   `tfsdk:"network_start"`
	NetworkEnd        types.String   `tfsdk:"network_end"`
	RestrictRoutes    types.Bool     `tfsdk:"restrict_routes"`
	IPv6              types.Bool     `tfsdk:"ipv6"`
	IPv6Firewall      types.Bool     `tfsdk:"ipv6_firewall"`
	BindAddress       types.String   `tfsdk:"bind_address"`
	DHParamBits       types.Int64    `tfsdk:"dh_param_bits"`
	Groups            types.List     `tfsdk:"groups"`
	MultiDevice       types.Bool     `tfsdk:"multi_device"`
	DNSServers        types.List     `tfsdk:"dns_servers"`
	SearchDomain      types.String   `tfsdk:"search_domain"`
	OTPAuth           types.Bool     `tfsdk:"otp_auth"`
	LZOCompression    types.Bool     `tfsdk:"lzo_compression"`
	InterClient       types.Bool     `tfsdk:"inter_client"`
	PingInterval      types.Int64    `tfsdk:"ping_interval"`
	PingTimeout       types.Int64    `tfsdk:"ping_timeout"`
	LinkPingInterval  types.Int64    `tfsdk:"link_ping_interval"`
	LinkPingTimeout   types.Int64    `tfsdk:"link_ping_timeout"`
	InactiveTimeout   types.Int64    `tfsdk:"inactive_timeout"`
	SessionTimeout    types.Int64    `tfsdk:"session_timeout"`
	AllowedDevices    types.String   `tfsdk:"allowed_devices"`
	MaxClients        types.Int64    `tfsdk:"max_clients"`
	MaxDevices        types.Int64    `tfsdk:"max_devices"`
	ReplicaCount      types.Int64    `tfsdk:"replica_count"`
	VXLAN             types.Bool     `tfsdk:"vxlan"`
	DNSMapping        types.Bool     `tfsdk:"dns_mapping"`
	Debug             types.Bool     `tfsdk:"debug"`
	RouteDNS          types.Bool     `tfsdk:"route_dns"`
	Routes            types.List     `tfsdk:"routes"`
}

func (r *ServerResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server"
}

func (r *ServerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The server resource allows managing information about a particular Pritunl server.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Server identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the server.",
				Required:            true,
			},
			"protocol": schema.StringAttribute{
				MarkdownDescription: "The protocol of the server.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("udp"),
				Validators: []validator.String{
					stringvalidator.OneOf("udp", "tcp"),
				},
			},
			"cipher": schema.StringAttribute{
				MarkdownDescription: "The cipher of the server.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("aes128"),
				Validators: []validator.String{
					stringvalidator.OneOf("none", "bf128", "bf256", "aes128", "aes192", "aes256"),
				},
			},
			"hash": schema.StringAttribute{
				MarkdownDescription: "The hash of the server.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("sha1"),
				Validators: []validator.String{
					stringvalidator.OneOf("none", "md5", "sha1", "sha256", "sha512"),
				},
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "The port of the server.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.Int64{
					int64validator.Between(1, 65535),
				},
			},
			"network": schema.StringAttribute{
				MarkdownDescription: "The network of the server.",
				Optional:            true,
				Computed:            true,
			},
			"wg": schema.BoolAttribute{
				MarkdownDescription: "Enable WireGuard.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"port_wg": schema.Int64Attribute{
				MarkdownDescription: "The WireGuard port of the server.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.Int64{
					int64validator.Between(1, 65535),
				},
			},
			"network_wg": schema.StringAttribute{
				MarkdownDescription: "The WireGuard network of the server.",
				Optional:            true,
				Computed:            true,
			},
			"network_mode": schema.StringAttribute{
				MarkdownDescription: "The network mode of the server.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("tunnel"),
				Validators: []validator.String{
					stringvalidator.OneOf("tunnel", "bridge"),
				},
			},
			"network_start": schema.StringAttribute{
				MarkdownDescription: "The network start of the server.",
				Optional:            true,
				Computed:            true,
			},
			"network_end": schema.StringAttribute{
				MarkdownDescription: "The network end of the server.",
				Optional:            true,
				Computed:            true,
			},
			"restrict_routes": schema.BoolAttribute{
				MarkdownDescription: "Restrict routes.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"ipv6": schema.BoolAttribute{
				MarkdownDescription: "Enable IPv6.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"ipv6_firewall": schema.BoolAttribute{
				MarkdownDescription: "Enable IPv6 firewall.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"bind_address": schema.StringAttribute{
				MarkdownDescription: "The bind address of the server.",
				Optional:            true,
				Computed:            true,
			},
			"dh_param_bits": schema.Int64Attribute{
				MarkdownDescription: "The DH param bits of the server.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(2048),
				Validators: []validator.Int64{
					int64validator.OneOf(1024, 1536, 2048, 3072, 4096),
				},
			},
			"groups": schema.ListAttribute{
				MarkdownDescription: "Enter list of groups to allow connections from. Names are case sensitive. If empty all groups will able to connect.",
				Optional:            true,
				ElementType:         types.StringType,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
				},
			},
			"multi_device": schema.BoolAttribute{
				MarkdownDescription: "Allow multiple devices.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"dns_servers": schema.ListAttribute{
				MarkdownDescription: "Dns servers to push to clients.",
				Optional:            true,
				ElementType:         types.StringType,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
				},
			},
			"search_domain": schema.StringAttribute{
				MarkdownDescription: "Search domain to push to clients.",
				Optional:            true,
				Computed:            true,
			},
			"otp_auth": schema.BoolAttribute{
				MarkdownDescription: "Enable OTP authentication.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"lzo_compression": schema.BoolAttribute{
				MarkdownDescription: "Enable LZO compression.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"inter_client": schema.BoolAttribute{
				MarkdownDescription: "Allow inter client communication.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"ping_interval": schema.Int64Attribute{
				MarkdownDescription: "The ping interval of the server.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(10),
				Validators: []validator.Int64{
					int64validator.Between(1, 86400),
				},
			},
			"ping_timeout": schema.Int64Attribute{
				MarkdownDescription: "The ping timeout of the server.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(60),
				Validators: []validator.Int64{
					int64validator.Between(1, 86400),
				},
			},
			"link_ping_interval": schema.Int64Attribute{
				MarkdownDescription: "The link ping interval of the server.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(1),
				Validators: []validator.Int64{
					int64validator.Between(1, 86400),
				},
			},
			"link_ping_timeout": schema.Int64Attribute{
				MarkdownDescription: "The link ping timeout of the server.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(5),
				Validators: []validator.Int64{
					int64validator.Between(1, 86400),
				},
			},
			"inactive_timeout": schema.Int64Attribute{
				MarkdownDescription: "The inactive timeout of the server.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(0),
				Validators: []validator.Int64{
					int64validator.Between(0, 86400),
				},
			},
			"session_timeout": schema.Int64Attribute{
				MarkdownDescription: "The session timeout of the server.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(0),
				Validators: []validator.Int64{
					int64validator.Between(0, 86400),
				},
			},
			"allowed_devices": schema.StringAttribute{
				MarkdownDescription: "The allowed devices of the server.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("mobile_desktop"),
				Validators: []validator.String{
					stringvalidator.OneOf("mobile", "desktop", "mobile_desktop"),
				},
			},
			"max_clients": schema.Int64Attribute{
				MarkdownDescription: "The max clients of the server.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(2048),
				Validators: []validator.Int64{
					int64validator.Between(1, 32768),
				},
			},
			"max_devices": schema.Int64Attribute{
				MarkdownDescription: "The max devices of the server.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(1),
				Validators: []validator.Int64{
					int64validator.Between(1, 255),
				},
			},
			"replica_count": schema.Int64Attribute{
				MarkdownDescription: "The replica count of the server.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(1),
				Validators: []validator.Int64{
					int64validator.Between(1, 10),
				},
			},
			"vxlan": schema.BoolAttribute{
				MarkdownDescription: "Enable VXLAN.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"dns_mapping": schema.BoolAttribute{
				MarkdownDescription: "Enable DNS mapping.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"debug": schema.BoolAttribute{
				MarkdownDescription: "Enable debug.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"route_dns": schema.BoolAttribute{
				MarkdownDescription: "Route DNS.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"routes": schema.ListNestedAttribute{
				MarkdownDescription: "Routes for the server.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"network": schema.StringAttribute{
							MarkdownDescription: "Network address with cidr subnet.",
							Required:            true,
						},
						"comment": schema.StringAttribute{
							MarkdownDescription: "Comment for the route.",
							Optional:            true,
						},
						"nat": schema.BoolAttribute{
							MarkdownDescription: "Enable NAT for the route.",
							Optional:            true,
						},
						"nat_interface": schema.StringAttribute{
							MarkdownDescription: "NAT interface for the route.",
							Optional:            true,
						},
						"nat_netmap": schema.StringAttribute{
							MarkdownDescription: "NAT netmap for the route.",
							Optional:            true,
						},
						"advertise": schema.BoolAttribute{
							MarkdownDescription: "Advertise the route.",
							Optional:            true,
						},
						"vpc_region": schema.StringAttribute{
							MarkdownDescription: "VPC region for the route.",
							Optional:            true,
						},
						"vpc_id": schema.StringAttribute{
							MarkdownDescription: "VPC ID for the route.",
							Optional:            true,
						},
						"net_gateway": schema.BoolAttribute{
							MarkdownDescription: "Enable net gateway for the route.",
							Optional:            true,
						},
					},
				},
			},
		},
	}
}

func (r *ServerResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(pritunl.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected pritunl.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *ServerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ServerResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	server := pritunl.Server{
		Name:             data.Name.ValueString(),
		Protocol:         data.Protocol.ValueString(),
		Cipher:           data.Cipher.ValueString(),
		Hash:             data.Hash.ValueString(),
		Network:          data.Network.ValueString(),
		WG:               data.WG.ValueBool(),
		NetworkWG:        data.NetworkWG.ValueString(),
		NetworkMode:      data.NetworkMode.ValueString(),
		NetworkStart:     data.NetworkStart.ValueString(),
		NetworkEnd:       data.NetworkEnd.ValueString(),
		RestrictRoutes:   data.RestrictRoutes.ValueBool(),
		IPv6:             data.IPv6.ValueBool(),
		IPv6Firewall:     data.IPv6Firewall.ValueBool(),
		BindAddress:      data.BindAddress.ValueString(),
		MultiDevice:      data.MultiDevice.ValueBool(),
		SearchDomain:     data.SearchDomain.ValueString(),
		OTPAuth:          data.OTPAuth.ValueBool(),
		LZOCompression:   data.LZOCompression.ValueBool(),
		InterClient:      data.InterClient.ValueBool(),
		AllowedDevices:   data.AllowedDevices.ValueString(),
		VXLAN:            data.VXLAN.ValueBool(),
		DNSMapping:       data.DNSMapping.ValueBool(),
		Debug:            data.Debug.ValueBool(),
		RouteDNS:         data.RouteDNS.ValueBool(),
	}

	if !data.Port.IsNull() {
		server.Port = int(data.Port.ValueInt64())
	}
	if !data.PortWG.IsNull() {
		server.PortWG = int(data.PortWG.ValueInt64())
	}
	if !data.DHParamBits.IsNull() {
		server.DHParamBits = int(data.DHParamBits.ValueInt64())
	}
	if !data.PingInterval.IsNull() {
		server.PingInterval = int(data.PingInterval.ValueInt64())
	}
	if !data.PingTimeout.IsNull() {
		server.PingTimeout = int(data.PingTimeout.ValueInt64())
	}
	if !data.LinkPingInterval.IsNull() {
		server.LinkPingInterval = int(data.LinkPingInterval.ValueInt64())
	}
	if !data.LinkPingTimeout.IsNull() {
		server.LinkPingTimeout = int(data.LinkPingTimeout.ValueInt64())
	}
	if !data.InactiveTimeout.IsNull() {
		server.InactiveTimeout = int(data.InactiveTimeout.ValueInt64())
	}
	if !data.SessionTimeout.IsNull() {
		server.SessionTimeout = int(data.SessionTimeout.ValueInt64())
	}
	if !data.MaxClients.IsNull() {
		server.MaxClients = int(data.MaxClients.ValueInt64())
	}
	if !data.MaxDevices.IsNull() {
		server.MaxDevices = int(data.MaxDevices.ValueInt64())
	}
	if !data.ReplicaCount.IsNull() {
		server.ReplicaCount = int(data.ReplicaCount.ValueInt64())
	}

	if !data.Groups.IsNull() && !data.Groups.IsUnknown() {
		var groups []string
		resp.Diagnostics.Append(data.Groups.ElementsAs(ctx, &groups, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		server.Groups = groups
	}

	if !data.DNSServers.IsNull() && !data.DNSServers.IsUnknown() {
		var dnsServers []string
		resp.Diagnostics.Append(data.DNSServers.ElementsAs(ctx, &dnsServers, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		server.DNSServers = dnsServers
	}

	serverResponse, err := r.client.CreateServer(server)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create server, got error: %s", err))
		return
	}

	data.ID = types.StringValue(serverResponse.ID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ServerResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	server, err := r.client.GetServer(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read server, got error: %s", err))
		return
	}

	data.Name = types.StringValue(server.Name)
	data.Protocol = types.StringValue(server.Protocol)
	data.Cipher = types.StringValue(server.Cipher)
	data.Hash = types.StringValue(server.Hash)
	data.Port = types.Int64Value(int64(server.Port))
	data.Network = types.StringValue(server.Network)
	data.WG = types.BoolValue(server.WG)
	data.PortWG = types.Int64Value(int64(server.PortWG))
	data.NetworkWG = types.StringValue(server.NetworkWG)
	data.NetworkMode = types.StringValue(server.NetworkMode)
	data.NetworkStart = types.StringValue(server.NetworkStart)
	data.NetworkEnd = types.StringValue(server.NetworkEnd)
	data.RestrictRoutes = types.BoolValue(server.RestrictRoutes)
	data.IPv6 = types.BoolValue(server.IPv6)
	data.IPv6Firewall = types.BoolValue(server.IPv6Firewall)
	data.BindAddress = types.StringValue(server.BindAddress)
	data.DHParamBits = types.Int64Value(int64(server.DHParamBits))
	data.MultiDevice = types.BoolValue(server.MultiDevice)
	data.SearchDomain = types.StringValue(server.SearchDomain)
	data.OTPAuth = types.BoolValue(server.OTPAuth)
	data.LZOCompression = types.BoolValue(server.LZOCompression)
	data.InterClient = types.BoolValue(server.InterClient)
	data.PingInterval = types.Int64Value(int64(server.PingInterval))
	data.PingTimeout = types.Int64Value(int64(server.PingTimeout))
	data.LinkPingInterval = types.Int64Value(int64(server.LinkPingInterval))
	data.LinkPingTimeout = types.Int64Value(int64(server.LinkPingTimeout))
	data.InactiveTimeout = types.Int64Value(int64(server.InactiveTimeout))
	data.SessionTimeout = types.Int64Value(int64(server.SessionTimeout))
	data.AllowedDevices = types.StringValue(server.AllowedDevices)
	data.MaxClients = types.Int64Value(int64(server.MaxClients))
	data.MaxDevices = types.Int64Value(int64(server.MaxDevices))
	data.ReplicaCount = types.Int64Value(int64(server.ReplicaCount))
	data.VXLAN = types.BoolValue(server.VXLAN)
	data.DNSMapping = types.BoolValue(server.DNSMapping)
	data.Debug = types.BoolValue(server.Debug)
	data.RouteDNS = types.BoolValue(server.RouteDNS)

	if len(server.Groups) > 0 {
		groupsAttr := make([]attr.Value, len(server.Groups))
		for i, group := range server.Groups {
			groupsAttr[i] = types.StringValue(group)
		}
		data.Groups, _ = types.ListValue(types.StringType, groupsAttr)
	} else {
		data.Groups = types.ListNull(types.StringType)
	}

	if len(server.DNSServers) > 0 {
		dnsServersAttr := make([]attr.Value, len(server.DNSServers))
		for i, dnsServer := range server.DNSServers {
			dnsServersAttr[i] = types.StringValue(dnsServer)
		}
		data.DNSServers, _ = types.ListValue(types.StringType, dnsServersAttr)
	} else {
		data.DNSServers = types.ListNull(types.StringType)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ServerResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	server := pritunl.Server{
		Name:             data.Name.ValueString(),
		Protocol:         data.Protocol.ValueString(),
		Cipher:           data.Cipher.ValueString(),
		Hash:             data.Hash.ValueString(),
		Network:          data.Network.ValueString(),
		WG:               data.WG.ValueBool(),
		NetworkWG:        data.NetworkWG.ValueString(),
		NetworkMode:      data.NetworkMode.ValueString(),
		NetworkStart:     data.NetworkStart.ValueString(),
		NetworkEnd:       data.NetworkEnd.ValueString(),
		RestrictRoutes:   data.RestrictRoutes.ValueBool(),
		IPv6:             data.IPv6.ValueBool(),
		IPv6Firewall:     data.IPv6Firewall.ValueBool(),
		BindAddress:      data.BindAddress.ValueString(),
		MultiDevice:      data.MultiDevice.ValueBool(),
		SearchDomain:     data.SearchDomain.ValueString(),
		OTPAuth:          data.OTPAuth.ValueBool(),
		LZOCompression:   data.LZOCompression.ValueBool(),
		InterClient:      data.InterClient.ValueBool(),
		AllowedDevices:   data.AllowedDevices.ValueString(),
		VXLAN:            data.VXLAN.ValueBool(),
		DNSMapping:       data.DNSMapping.ValueBool(),
		Debug:            data.Debug.ValueBool(),
		RouteDNS:         data.RouteDNS.ValueBool(),
	}

	if !data.Port.IsNull() {
		server.Port = int(data.Port.ValueInt64())
	}
	if !data.PortWG.IsNull() {
		server.PortWG = int(data.PortWG.ValueInt64())
	}
	if !data.DHParamBits.IsNull() {
		server.DHParamBits = int(data.DHParamBits.ValueInt64())
	}
	if !data.PingInterval.IsNull() {
		server.PingInterval = int(data.PingInterval.ValueInt64())
	}
	if !data.PingTimeout.IsNull() {
		server.PingTimeout = int(data.PingTimeout.ValueInt64())
	}
	if !data.LinkPingInterval.IsNull() {
		server.LinkPingInterval = int(data.LinkPingInterval.ValueInt64())
	}
	if !data.LinkPingTimeout.IsNull() {
		server.LinkPingTimeout = int(data.LinkPingTimeout.ValueInt64())
	}
	if !data.InactiveTimeout.IsNull() {
		server.InactiveTimeout = int(data.InactiveTimeout.ValueInt64())
	}
	if !data.SessionTimeout.IsNull() {
		server.SessionTimeout = int(data.SessionTimeout.ValueInt64())
	}
	if !data.MaxClients.IsNull() {
		server.MaxClients = int(data.MaxClients.ValueInt64())
	}
	if !data.MaxDevices.IsNull() {
		server.MaxDevices = int(data.MaxDevices.ValueInt64())
	}
	if !data.ReplicaCount.IsNull() {
		server.ReplicaCount = int(data.ReplicaCount.ValueInt64())
	}

	if !data.Groups.IsNull() && !data.Groups.IsUnknown() {
		var groups []string
		resp.Diagnostics.Append(data.Groups.ElementsAs(ctx, &groups, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		server.Groups = groups
	}

	if !data.DNSServers.IsNull() && !data.DNSServers.IsUnknown() {
		var dnsServers []string
		resp.Diagnostics.Append(data.DNSServers.ElementsAs(ctx, &dnsServers, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		server.DNSServers = dnsServers
	}

	err := r.client.UpdateServer(data.ID.ValueString(), server)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update server, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ServerResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteServer(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete server, got error: %s", err))
		return
	}
}

func (r *ServerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
