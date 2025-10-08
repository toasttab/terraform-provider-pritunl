package provider

import (
	"context"
	"fmt"
	"regexp"

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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
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
	SSOAuth           types.Bool     `tfsdk:"sso_auth"`
	DeviceAuth        types.Bool     `tfsdk:"device_auth"`
	DynamicFirewall   types.Bool     `tfsdk:"dynamic_firewall"`
	Route             types.List     `tfsdk:"route"`
	OrganizationIDs   types.List     `tfsdk:"organization_ids"`
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
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.Between(1, 65535),
				},
			},
			"network": schema.StringAttribute{
				MarkdownDescription: "The network of the server.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.Between(1, 65535),
				},
			},
			"network_wg": schema.StringAttribute{
				MarkdownDescription: "The WireGuard network of the server.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"network_end": schema.StringAttribute{
				MarkdownDescription: "The network end of the server.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
				Default:             booldefault.StaticBool(false),
			},
			"bind_address": schema.StringAttribute{
				MarkdownDescription: "The bind address of the server.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
					listvalidator.ValueStringsAre(
						stringvalidator.LengthAtLeast(1),
						stringvalidator.RegexMatches(
							regexp.MustCompile(`^[^\s]+$`),
							"group names cannot contain spaces",
						),
					),
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
				Default:             int64default.StaticInt64(2000),
				Validators: []validator.Int64{
					int64validator.Between(1, 32768),
				},
			},
			"max_devices": schema.Int64Attribute{
				MarkdownDescription: "The max devices of the server.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(0),
				Validators: []validator.Int64{
					int64validator.Between(0, 255),
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
			"sso_auth": schema.BoolAttribute{
				MarkdownDescription: "Enable SSO authentication.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"device_auth": schema.BoolAttribute{
				MarkdownDescription: "Enable device authentication.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"dynamic_firewall": schema.BoolAttribute{
				MarkdownDescription: "Enable dynamic firewall.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"organization_ids": schema.ListAttribute{
				MarkdownDescription: "List of organization IDs attached to the server.",
				Optional:            true,
				ElementType:         types.StringType,
			},
		},
		Blocks: map[string]schema.Block{
			"route": schema.ListNestedBlock{
				MarkdownDescription: "Routes for the server.",
				NestedObject: schema.NestedBlockObject{
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
							Computed:            true,
							Default:             booldefault.StaticBool(false),
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
							Computed:            true,
							Default:             booldefault.StaticBool(false),
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
							Computed:            true,
							Default:             booldefault.StaticBool(false),
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
		OtpAuth:          data.OTPAuth.ValueBool(),
		LzoCompression:   data.LZOCompression.ValueBool(),
		InterClient:      data.InterClient.ValueBool(),
		AllowedDevices:   data.AllowedDevices.ValueString(),
		VxLan:            data.VXLAN.ValueBool(),
		DnsMapping:       data.DNSMapping.ValueBool(),
		Debug:            data.Debug.ValueBool(),
		SsoAuth:          data.SSOAuth.ValueBool(),
		DeviceAuth:       data.DeviceAuth.ValueBool(),
		DynamicFirewall:  data.DynamicFirewall.ValueBool(),
	}

	if !data.Port.IsNull() {
		server.Port = int(data.Port.ValueInt64())
	}
	if !data.PortWG.IsNull() {
		server.PortWG = int(data.PortWG.ValueInt64())
	}
	if !data.DHParamBits.IsNull() {
		server.DhParamBits = int(data.DHParamBits.ValueInt64())
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
		server.DnsServers = dnsServers
	}

	serverMap := map[string]interface{}{
		"name":              server.Name,
		"protocol":          server.Protocol,
		"cipher":            server.Cipher,
		"hash":              server.Hash,
		"port":              server.Port,
		"network":           server.Network,
		"wg":                server.WG,
		"port_wg":           server.PortWG,
		"network_wg":        server.NetworkWG,
		"network_mode":      server.NetworkMode,
		"network_start":     server.NetworkStart,
		"network_end":       server.NetworkEnd,
		"restrict_routes":   server.RestrictRoutes,
		"ipv6":              server.IPv6,
		"ipv6_firewall":     server.IPv6Firewall,
		"bind_address":      server.BindAddress,
		"dh_param_bits":     server.DhParamBits,
		"multi_device":      server.MultiDevice,
		"search_domain":     server.SearchDomain,
		"otp_auth":          server.OtpAuth,
		"lzo_compression":   server.LzoCompression,
		"inter_client":      server.InterClient,
		"allowed_devices":   server.AllowedDevices,
		"vxlan":             server.VxLan,
		"dns_mapping":       server.DnsMapping,
		"debug":             server.Debug,
		"sso_auth":          server.SsoAuth,
		"device_auth":       server.DeviceAuth,
		"dynamic_firewall":  server.DynamicFirewall,
		"groups":            server.Groups,
		"dns_servers":       server.DnsServers,
	}

	serverResponse, err := r.client.CreateServer(serverMap)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create server, got error: %s", err))
		return
	}

	data.ID = types.StringValue(serverResponse.ID)
	data.Name = types.StringValue(serverResponse.Name)
	data.Protocol = types.StringValue(serverResponse.Protocol)
	data.Cipher = types.StringValue(serverResponse.Cipher)
	data.Hash = types.StringValue(serverResponse.Hash)
	data.Port = types.Int64Value(int64(serverResponse.Port))
	data.Network = types.StringValue(serverResponse.Network)
	data.WG = types.BoolValue(serverResponse.WG)
	data.PortWG = types.Int64Value(int64(serverResponse.PortWG))
	data.NetworkWG = types.StringValue(serverResponse.NetworkWG)
	data.NetworkMode = types.StringValue(serverResponse.NetworkMode)
	data.NetworkStart = types.StringValue(serverResponse.NetworkStart)
	data.NetworkEnd = types.StringValue(serverResponse.NetworkEnd)
	data.RestrictRoutes = types.BoolValue(serverResponse.RestrictRoutes)
	data.IPv6 = types.BoolValue(serverResponse.IPv6)
	data.IPv6Firewall = types.BoolValue(serverResponse.IPv6Firewall)
	data.BindAddress = types.StringValue(serverResponse.BindAddress)
	data.DHParamBits = types.Int64Value(int64(serverResponse.DhParamBits))
	data.MultiDevice = types.BoolValue(serverResponse.MultiDevice)
	data.SearchDomain = types.StringValue(serverResponse.SearchDomain)
	data.OTPAuth = types.BoolValue(serverResponse.OtpAuth)
	data.LZOCompression = types.BoolValue(serverResponse.LzoCompression)
	data.InterClient = types.BoolValue(serverResponse.InterClient)
	data.PingInterval = types.Int64Value(int64(serverResponse.PingInterval))
	data.PingTimeout = types.Int64Value(int64(serverResponse.PingTimeout))
	data.LinkPingInterval = types.Int64Value(int64(serverResponse.LinkPingInterval))
	data.LinkPingTimeout = types.Int64Value(int64(serverResponse.LinkPingTimeout))
	data.InactiveTimeout = types.Int64Value(int64(serverResponse.InactiveTimeout))
	data.SessionTimeout = types.Int64Value(int64(serverResponse.SessionTimeout))
	data.AllowedDevices = types.StringValue(serverResponse.AllowedDevices)
	data.MaxClients = types.Int64Value(int64(serverResponse.MaxClients))
	data.MaxDevices = types.Int64Value(int64(serverResponse.MaxDevices))
	data.ReplicaCount = types.Int64Value(int64(serverResponse.ReplicaCount))
	data.VXLAN = types.BoolValue(serverResponse.VxLan)
	data.DNSMapping = types.BoolValue(serverResponse.DnsMapping)
	data.Debug = types.BoolValue(serverResponse.Debug)
	data.SSOAuth = types.BoolValue(serverResponse.SsoAuth)
	data.DeviceAuth = types.BoolValue(serverResponse.DeviceAuth)
	data.DynamicFirewall = types.BoolValue(serverResponse.DynamicFirewall)

	if len(serverResponse.Groups) > 0 {
		groupsAttr := make([]attr.Value, len(serverResponse.Groups))
		for i, group := range serverResponse.Groups {
			groupsAttr[i] = types.StringValue(group)
		}
		data.Groups, _ = types.ListValue(types.StringType, groupsAttr)
	} else {
		data.Groups = types.ListNull(types.StringType)
	}

	if len(serverResponse.DnsServers) > 0 {
		dnsServersAttr := make([]attr.Value, len(serverResponse.DnsServers))
		for i, dnsServer := range serverResponse.DnsServers {
			dnsServersAttr[i] = types.StringValue(dnsServer)
		}
		data.DNSServers, _ = types.ListValue(types.StringType, dnsServersAttr)
	} else {
		data.DNSServers = types.ListNull(types.StringType)
	}

	if !data.OrganizationIDs.IsNull() && !data.OrganizationIDs.IsUnknown() {
		var orgIDs []string
		resp.Diagnostics.Append(data.OrganizationIDs.ElementsAs(ctx, &orgIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		
		for _, orgID := range orgIDs {
			err := r.client.AttachOrganizationToServer(orgID, serverResponse.ID)
			if err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to attach organization %s to server, got error: %s", orgID, err))
				return
			}
		}
	}

	if !data.Route.IsNull() && !data.Route.IsUnknown() && len(data.Route.Elements()) > 0 {
		routeElements := data.Route.Elements()
		for _, routeElement := range routeElements {
			routeObj := routeElement.(types.Object)
			routeAttrs := routeObj.Attributes()
			
			route := pritunl.Route{
				Network: routeAttrs["network"].(types.String).ValueString(),
				Comment: routeAttrs["comment"].(types.String).ValueString(),
				Nat: routeAttrs["nat"].(types.Bool).ValueBool(),
				NatInterface: routeAttrs["nat_interface"].(types.String).ValueString(),
				NatNetmap: routeAttrs["nat_netmap"].(types.String).ValueString(),
				Advertise: routeAttrs["advertise"].(types.Bool).ValueBool(),
				VpcRegion: routeAttrs["vpc_region"].(types.String).ValueString(),
				VpcID: routeAttrs["vpc_id"].(types.String).ValueString(),
				NetGateway: routeAttrs["net_gateway"].(types.Bool).ValueBool(),
			}
			
			err := r.client.AddRouteToServer(serverResponse.ID, route)
			if err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to add route to server, got error: %s", err))
				return
			}
		}
	}

	if !data.OrganizationIDs.IsNull() && !data.OrganizationIDs.IsUnknown() {
	} else {
		data.OrganizationIDs = types.ListNull(types.StringType)
	}


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
	data.DHParamBits = types.Int64Value(int64(server.DhParamBits))
	data.MultiDevice = types.BoolValue(server.MultiDevice)
	data.SearchDomain = types.StringValue(server.SearchDomain)
	data.OTPAuth = types.BoolValue(server.OtpAuth)
	data.LZOCompression = types.BoolValue(server.LzoCompression)
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
	data.VXLAN = types.BoolValue(server.VxLan)
	data.DNSMapping = types.BoolValue(server.DnsMapping)
	data.Debug = types.BoolValue(server.Debug)
	data.SSOAuth = types.BoolValue(server.SsoAuth)
	data.DeviceAuth = types.BoolValue(server.DeviceAuth)
	data.DynamicFirewall = types.BoolValue(server.DynamicFirewall)
	
	if data.RouteDNS.IsNull() || data.RouteDNS.IsUnknown() {
		data.RouteDNS = types.BoolValue(true)
	}

	if len(server.Groups) > 0 {
		groupsAttr := make([]attr.Value, len(server.Groups))
		for i, group := range server.Groups {
			groupsAttr[i] = types.StringValue(group)
		}
		data.Groups, _ = types.ListValue(types.StringType, groupsAttr)
	} else {
		data.Groups = types.ListNull(types.StringType)
	}

	if len(server.DnsServers) > 0 {
		dnsServersAttr := make([]attr.Value, len(server.DnsServers))
		for i, dnsServer := range server.DnsServers {
			dnsServersAttr[i] = types.StringValue(dnsServer)
		}
		data.DNSServers, _ = types.ListValue(types.StringType, dnsServersAttr)
	} else {
		data.DNSServers = types.ListNull(types.StringType)
	}

	attachedOrgs, err := r.client.GetOrganizationsByServer(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get organizations for server, got error: %s", err))
		return
	}
	
	if !data.OrganizationIDs.IsNull() && !data.OrganizationIDs.IsUnknown() {
		var currentOrgIDs []string
		resp.Diagnostics.Append(data.OrganizationIDs.ElementsAs(ctx, &currentOrgIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		
		attachedOrgMap := make(map[string]bool)
		for _, org := range attachedOrgs {
			attachedOrgMap[org.ID] = true
		}
		
		validOrgIDs := make([]string, 0)
		for _, orgID := range currentOrgIDs {
			if attachedOrgMap[orgID] {
				validOrgIDs = append(validOrgIDs, orgID)
			}
		}
		
		if len(validOrgIDs) > 0 {
			orgIDsAttr := make([]attr.Value, len(validOrgIDs))
			for i, orgID := range validOrgIDs {
				orgIDsAttr[i] = types.StringValue(orgID)
			}
			data.OrganizationIDs, _ = types.ListValue(types.StringType, orgIDsAttr)
		} else {
			data.OrganizationIDs = types.ListNull(types.StringType)
		}
	} else {
		if len(attachedOrgs) > 0 {
			orgIDsAttr := make([]attr.Value, len(attachedOrgs))
			for i, org := range attachedOrgs {
				orgIDsAttr[i] = types.StringValue(org.ID)
			}
			data.OrganizationIDs, _ = types.ListValue(types.StringType, orgIDsAttr)
		} else {
			data.OrganizationIDs = types.ListNull(types.StringType)
		}
	}

	attachedRoutes, err := r.client.GetRoutesByServer(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get routes for server, got error: %s", err))
		return
	}
	
	if !data.Route.IsNull() && !data.Route.IsUnknown() {
		var currentRoutes []types.Object
		resp.Diagnostics.Append(data.Route.ElementsAs(ctx, &currentRoutes, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		
		attachedRouteMap := make(map[string]pritunl.Route)
		for _, route := range attachedRoutes {
			attachedRouteMap[route.Network] = route
		}
		
		validRoutes := make([]attr.Value, 0)
		for _, routeObj := range currentRoutes {
			routeAttrs := routeObj.Attributes()
			network := routeAttrs["network"].(types.String).ValueString()
			
			if attachedRoute, exists := attachedRouteMap[network]; exists {
				routeObjValue, _ := types.ObjectValue(map[string]attr.Type{
					"network":       types.StringType,
					"comment":       types.StringType,
					"nat":           types.BoolType,
					"nat_interface": types.StringType,
					"nat_netmap":    types.StringType,
					"advertise":     types.BoolType,
					"vpc_region":    types.StringType,
					"vpc_id":        types.StringType,
					"net_gateway":   types.BoolType,
				}, map[string]attr.Value{
					"network":       types.StringValue(attachedRoute.Network),
					"comment":       types.StringValue(attachedRoute.Comment),
					"nat":           types.BoolValue(attachedRoute.Nat),
					"nat_interface": types.StringValue(attachedRoute.NatInterface),
					"nat_netmap":    types.StringValue(attachedRoute.NatNetmap),
					"advertise":     types.BoolValue(attachedRoute.Advertise),
					"vpc_region":    types.StringValue(attachedRoute.VpcRegion),
					"vpc_id":        types.StringValue(attachedRoute.VpcID),
					"net_gateway":   types.BoolValue(attachedRoute.NetGateway),
				})
				validRoutes = append(validRoutes, routeObjValue)
			}
		}
		
		if len(validRoutes) > 0 {
			data.Route, _ = types.ListValue(types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"network":       types.StringType,
					"comment":       types.StringType,
					"nat":           types.BoolType,
					"nat_interface": types.StringType,
					"nat_netmap":    types.StringType,
					"advertise":     types.BoolType,
					"vpc_region":    types.StringType,
					"vpc_id":        types.StringType,
					"net_gateway":   types.BoolType,
				},
			}, validRoutes)
		} else {
			data.Route = types.ListNull(types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"network":       types.StringType,
					"comment":       types.StringType,
					"nat":           types.BoolType,
					"nat_interface": types.StringType,
					"nat_netmap":    types.StringType,
					"advertise":     types.BoolType,
					"vpc_region":    types.StringType,
					"vpc_id":        types.StringType,
					"net_gateway":   types.BoolType,
				},
			})
		}
	} else {
		if len(attachedRoutes) > 0 {
			routeAttrs := make([]attr.Value, len(attachedRoutes))
			for i, route := range attachedRoutes {
				routeObjValue, _ := types.ObjectValue(map[string]attr.Type{
					"network":       types.StringType,
					"comment":       types.StringType,
					"nat":           types.BoolType,
					"nat_interface": types.StringType,
					"nat_netmap":    types.StringType,
					"advertise":     types.BoolType,
					"vpc_region":    types.StringType,
					"vpc_id":        types.StringType,
					"net_gateway":   types.BoolType,
				}, map[string]attr.Value{
					"network":       types.StringValue(route.Network),
					"comment":       types.StringValue(route.Comment),
					"nat":           types.BoolValue(route.Nat),
					"nat_interface": types.StringValue(route.NatInterface),
					"nat_netmap":    types.StringValue(route.NatNetmap),
					"advertise":     types.BoolValue(route.Advertise),
					"vpc_region":    types.StringValue(route.VpcRegion),
					"vpc_id":        types.StringValue(route.VpcID),
					"net_gateway":   types.BoolValue(route.NetGateway),
				})
				routeAttrs[i] = routeObjValue
			}
			data.Route, _ = types.ListValue(types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"network":       types.StringType,
					"comment":       types.StringType,
					"nat":           types.BoolType,
					"nat_interface": types.StringType,
					"nat_netmap":    types.StringType,
					"advertise":     types.BoolType,
					"vpc_region":    types.StringType,
					"vpc_id":        types.StringType,
					"net_gateway":   types.BoolType,
				},
			}, routeAttrs)
		} else {
			data.Route = types.ListNull(types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"network":       types.StringType,
					"comment":       types.StringType,
					"nat":           types.BoolType,
					"nat_interface": types.StringType,
					"nat_netmap":    types.StringType,
					"advertise":     types.BoolType,
					"vpc_region":    types.StringType,
					"vpc_id":        types.StringType,
					"net_gateway":   types.BoolType,
				},
			})
		}
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
		OtpAuth:          data.OTPAuth.ValueBool(),
		LzoCompression:   data.LZOCompression.ValueBool(),
		InterClient:      data.InterClient.ValueBool(),
		AllowedDevices:   data.AllowedDevices.ValueString(),
		VxLan:            data.VXLAN.ValueBool(),
		DnsMapping:       data.DNSMapping.ValueBool(),
		Debug:            data.Debug.ValueBool(),
		SsoAuth:          data.SSOAuth.ValueBool(),
		DeviceAuth:       data.DeviceAuth.ValueBool(),
		DynamicFirewall:  data.DynamicFirewall.ValueBool(),
	}

	if !data.Port.IsNull() {
		server.Port = int(data.Port.ValueInt64())
	}
	if !data.PortWG.IsNull() {
		server.PortWG = int(data.PortWG.ValueInt64())
	}
	if !data.DHParamBits.IsNull() {
		server.DhParamBits = int(data.DHParamBits.ValueInt64())
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
		server.DnsServers = dnsServers
	}

	err := r.client.UpdateServer(data.ID.ValueString(), &server)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update server, got error: %s", err))
		return
	}

	var planOrgIDs, stateOrgIDs []string
	if !data.OrganizationIDs.IsNull() && !data.OrganizationIDs.IsUnknown() {
		resp.Diagnostics.Append(data.OrganizationIDs.ElementsAs(ctx, &planOrgIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	var currentState ServerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &currentState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !currentState.OrganizationIDs.IsNull() && !currentState.OrganizationIDs.IsUnknown() {
		resp.Diagnostics.Append(currentState.OrganizationIDs.ElementsAs(ctx, &stateOrgIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	for _, stateOrgID := range stateOrgIDs {
		found := false
		for _, planOrgID := range planOrgIDs {
			if stateOrgID == planOrgID {
				found = true
				break
			}
		}
		if !found {
			err := r.client.DetachOrganizationFromServer(stateOrgID, data.ID.ValueString())
			if err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to detach organization %s from server, got error: %s", stateOrgID, err))
				return
			}
		}
	}

	for _, planOrgID := range planOrgIDs {
		found := false
		for _, stateOrgID := range stateOrgIDs {
			if planOrgID == stateOrgID {
				found = true
				break
			}
		}
		if !found {
			err := r.client.AttachOrganizationToServer(planOrgID, data.ID.ValueString())
			if err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to attach organization %s to server, got error: %s", planOrgID, err))
				return
			}
		}
	}

	if !data.OrganizationIDs.IsNull() && !data.OrganizationIDs.IsUnknown() {
	} else {
		data.OrganizationIDs = types.ListNull(types.StringType)
	}

	var planRoutes, stateRoutes []types.Object
	if !data.Route.IsNull() && !data.Route.IsUnknown() {
		resp.Diagnostics.Append(data.Route.ElementsAs(ctx, &planRoutes, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if !currentState.Route.IsNull() && !currentState.Route.IsUnknown() {
		resp.Diagnostics.Append(currentState.Route.ElementsAs(ctx, &stateRoutes, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	for _, stateRouteObj := range stateRoutes {
		stateRouteAttrs := stateRouteObj.Attributes()
		stateNetwork := stateRouteAttrs["network"].(types.String).ValueString()
		
		found := false
		for _, planRouteObj := range planRoutes {
			planRouteAttrs := planRouteObj.Attributes()
			planNetwork := planRouteAttrs["network"].(types.String).ValueString()
			if stateNetwork == planNetwork {
				found = true
				break
			}
		}
		
		if !found {
			stateRoute := pritunl.Route{
				Network: stateNetwork,
			}
			err := r.client.DeleteRouteFromServer(data.ID.ValueString(), stateRoute)
			if err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete route %s from server, got error: %s", stateNetwork, err))
				return
			}
		}
	}

	for _, planRouteObj := range planRoutes {
		planRouteAttrs := planRouteObj.Attributes()
		planRoute := pritunl.Route{
			Network:      planRouteAttrs["network"].(types.String).ValueString(),
			Comment:      planRouteAttrs["comment"].(types.String).ValueString(),
			Nat:          planRouteAttrs["nat"].(types.Bool).ValueBool(),
			NatInterface: planRouteAttrs["nat_interface"].(types.String).ValueString(),
			NatNetmap:    planRouteAttrs["nat_netmap"].(types.String).ValueString(),
			Advertise:    planRouteAttrs["advertise"].(types.Bool).ValueBool(),
			VpcRegion:    planRouteAttrs["vpc_region"].(types.String).ValueString(),
			VpcID:        planRouteAttrs["vpc_id"].(types.String).ValueString(),
			NetGateway:   planRouteAttrs["net_gateway"].(types.Bool).ValueBool(),
		}

		found := false
		for _, stateRouteObj := range stateRoutes {
			stateRouteAttrs := stateRouteObj.Attributes()
			stateNetwork := stateRouteAttrs["network"].(types.String).ValueString()
			if planRoute.Network == stateNetwork {
				found = true
				break
			}
		}

		if !found {
			err := r.client.AddRouteToServer(data.ID.ValueString(), planRoute)
			if err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to add route %s to server, got error: %s", planRoute.Network, err))
				return
			}
		}
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
