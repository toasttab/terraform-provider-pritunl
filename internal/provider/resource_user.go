package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/disc/terraform-provider-pritunl/internal/pritunl"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &UserResource{}
var _ resource.ResourceWithImportState = &UserResource{}

func NewUserResource() resource.Resource {
	return &UserResource{}
}

type UserResource struct {
	client pritunl.Client
}

type UserResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	OrganizationID   types.String `tfsdk:"organization_id"`
	Groups           types.List   `tfsdk:"groups"`
	Email            types.String `tfsdk:"email"`
	Disabled         types.Bool   `tfsdk:"disabled"`
	PortForwarding   types.List   `tfsdk:"port_forwarding"`
	NetworkLinks     types.List   `tfsdk:"network_links"`
	ClientToClient   types.Bool   `tfsdk:"client_to_client"`
	AuthType         types.String `tfsdk:"auth_type"`
	MacAddresses     types.List   `tfsdk:"mac_addresses"`
	DNSServers       types.List   `tfsdk:"dns_servers"`
	DNSSuffix        types.String `tfsdk:"dns_suffix"`
	BypassSecondary  types.Bool   `tfsdk:"bypass_secondary"`
	Pin              types.String `tfsdk:"pin"`
}

func (r *UserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *UserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The user resource allows managing information about a particular Pritunl user.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "User identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the user.",
				Required:            true,
			},
			"organization_id": schema.StringAttribute{
				MarkdownDescription: "The organizations that user belongs to.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
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
			"email": schema.StringAttribute{
				MarkdownDescription: "User email address.",
				Optional:            true,
			},
			"disabled": schema.BoolAttribute{
				MarkdownDescription: "Shows if user is disabled",
				Optional:            true,
			},
			"port_forwarding": schema.ListAttribute{
				MarkdownDescription: "Comma seperated list of ports to forward using format source_port:dest_port/protocol or start_port-end_port/protocol. Such as 80, 80/tcp, 80:8000/tcp, 1000-2000/udp.",
				Optional:            true,
				ElementType:         types.MapType{ElemType: types.StringType},
			},
			"network_links": schema.ListAttribute{
				MarkdownDescription: "Network address with cidr subnet. This will provision access to a clients local network to the attached vpn servers and other clients. Multiple networks may be separated by a comma. Router must have a static route to VPN virtual network through client.",
				Optional:            true,
				ElementType:         types.StringType,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
				},
			},
			"client_to_client": schema.BoolAttribute{
				MarkdownDescription: "Only allow this client to communicate with other clients. Access to routed networks will be blocked.",
				Optional:            true,
			},
			"auth_type": schema.StringAttribute{
				MarkdownDescription: "User authentication type. This will determine how the user authenticates. This should be set automatically when the user authenticates with single sign-on.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("local", "duo", "yubico", "azure", "azure_duo", "azure_yubico", "google", "google_duo", "google_yubico", "slack", "slack_duo", "slack_yubico", "saml", "saml_duo", "saml_yubico", "saml_okta", "saml_okta_duo", "saml_okta_yubico", "saml_onelogin", "saml_onelogin_duo", "saml_onelogin_yubico", "radius", "radius_duo", "plugin"),
				},
			},
			"mac_addresses": schema.ListAttribute{
				MarkdownDescription: "Comma separated list of MAC addresses client is allowed to connect from. The validity of the MAC address provided by the VPN client cannot be verified.",
				Optional:            true,
				ElementType:         types.StringType,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
				},
			},
			"dns_servers": schema.ListAttribute{
				MarkdownDescription: "Dns server with port to forward sub-domain dns requests coming from this users domain. Multiple dns servers may be separated by a comma.",
				Optional:            true,
				ElementType:         types.StringType,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
				},
			},
			"dns_suffix": schema.StringAttribute{
				MarkdownDescription: "The suffix to use when forwarding dns requests. The full dns request will be the combination of the sub-domain of the users dns name suffixed by the dns suffix.",
				Optional:            true,
			},
			"bypass_secondary": schema.BoolAttribute{
				MarkdownDescription: "Bypass secondary authentication such as the PIN and two-factor authentication. Use for server users that can't provide a two-factor code.",
				Optional:            true,
			},
			"pin": schema.StringAttribute{
				MarkdownDescription: "The PIN for user authentication.",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

func (r *UserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *UserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data UserResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	user := pritunl.User{
		Name:            data.Name.ValueString(),
		Organization:    data.OrganizationID.ValueString(),
		Email:           data.Email.ValueString(),
		Disabled:        data.Disabled.ValueBool(),
		ClientToClient:  data.ClientToClient.ValueBool(),
		AuthType:        data.AuthType.ValueString(),
		DnsSuffix:       data.DNSSuffix.ValueString(),
		BypassSecondary: data.BypassSecondary.ValueBool(),
	}

	if !data.Groups.IsNull() && !data.Groups.IsUnknown() {
		var groups []string
		resp.Diagnostics.Append(data.Groups.ElementsAs(ctx, &groups, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		user.Groups = groups
	}

	if !data.NetworkLinks.IsNull() && !data.NetworkLinks.IsUnknown() {
		var networkLinks []string
		resp.Diagnostics.Append(data.NetworkLinks.ElementsAs(ctx, &networkLinks, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		user.NetworkLinks = networkLinks
	}

	if !data.MacAddresses.IsNull() && !data.MacAddresses.IsUnknown() {
		var macAddresses []string
		resp.Diagnostics.Append(data.MacAddresses.ElementsAs(ctx, &macAddresses, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		user.MacAddresses = macAddresses
	}

	if !data.DNSServers.IsNull() && !data.DNSServers.IsUnknown() {
		var dnsServers []string
		resp.Diagnostics.Append(data.DNSServers.ElementsAs(ctx, &dnsServers, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		user.DnsServers = dnsServers
	}

	userResponse, err := r.client.CreateUser(user)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create user, got error: %s", err))
		return
	}

	data.ID = types.StringValue(userResponse.ID)
	data.Name = types.StringValue(userResponse.Name)
	data.AuthType = types.StringValue(userResponse.AuthType)
	
	if userResponse.Email != "" {
		data.Email = types.StringValue(userResponse.Email)
	}
	if userResponse.DnsSuffix != "" {
		data.DNSSuffix = types.StringValue(userResponse.DnsSuffix)
	}
	
	if userResponse.Disabled {
		data.Disabled = types.BoolValue(true)
	}
	if userResponse.ClientToClient {
		data.ClientToClient = types.BoolValue(true)
	}
	if userResponse.BypassSecondary {
		data.BypassSecondary = types.BoolValue(true)
	}

	if len(userResponse.Groups) > 0 {
		groupsAttr := make([]attr.Value, len(userResponse.Groups))
		for i, group := range userResponse.Groups {
			groupsAttr[i] = types.StringValue(group)
		}
		data.Groups, _ = types.ListValue(types.StringType, groupsAttr)
	} else {
		data.Groups = types.ListNull(types.StringType)
	}

	if len(userResponse.NetworkLinks) > 0 {
		networkLinksAttr := make([]attr.Value, len(userResponse.NetworkLinks))
		for i, networkLink := range userResponse.NetworkLinks {
			networkLinksAttr[i] = types.StringValue(networkLink)
		}
		data.NetworkLinks, _ = types.ListValue(types.StringType, networkLinksAttr)
	} else {
		data.NetworkLinks = types.ListNull(types.StringType)
	}

	if len(userResponse.MacAddresses) > 0 {
		macAddressesAttr := make([]attr.Value, len(userResponse.MacAddresses))
		for i, macAddress := range userResponse.MacAddresses {
			macAddressesAttr[i] = types.StringValue(macAddress)
		}
		data.MacAddresses, _ = types.ListValue(types.StringType, macAddressesAttr)
	} else {
		data.MacAddresses = types.ListNull(types.StringType)
	}

	if len(userResponse.DnsServers) > 0 {
		dnsServersAttr := make([]attr.Value, len(userResponse.DnsServers))
		for i, dnsServer := range userResponse.DnsServers {
			dnsServersAttr[i] = types.StringValue(dnsServer)
		}
		data.DNSServers, _ = types.ListValue(types.StringType, dnsServersAttr)
	} else {
		data.DNSServers = types.ListNull(types.StringType)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data UserResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	user, err := r.client.GetUser(data.ID.ValueString(), data.OrganizationID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read user, got error: %s", err))
		return
	}

	data.Name = types.StringValue(user.Name)
	data.Email = types.StringValue(user.Email)
	data.Disabled = types.BoolValue(user.Disabled)
	data.ClientToClient = types.BoolValue(user.ClientToClient)
	data.AuthType = types.StringValue(user.AuthType)
	data.DNSSuffix = types.StringValue(user.DnsSuffix)
	data.BypassSecondary = types.BoolValue(user.BypassSecondary)

	if len(user.Groups) > 0 {
		groupsAttr := make([]attr.Value, len(user.Groups))
		for i, group := range user.Groups {
			groupsAttr[i] = types.StringValue(group)
		}
		data.Groups, _ = types.ListValue(types.StringType, groupsAttr)
	} else {
		data.Groups = types.ListNull(types.StringType)
	}

	if len(user.NetworkLinks) > 0 {
		networkLinksAttr := make([]attr.Value, len(user.NetworkLinks))
		for i, networkLink := range user.NetworkLinks {
			networkLinksAttr[i] = types.StringValue(networkLink)
		}
		data.NetworkLinks, _ = types.ListValue(types.StringType, networkLinksAttr)
	} else {
		data.NetworkLinks = types.ListNull(types.StringType)
	}

	if len(user.MacAddresses) > 0 {
		macAddressesAttr := make([]attr.Value, len(user.MacAddresses))
		for i, macAddress := range user.MacAddresses {
			macAddressesAttr[i] = types.StringValue(macAddress)
		}
		data.MacAddresses, _ = types.ListValue(types.StringType, macAddressesAttr)
	} else {
		data.MacAddresses = types.ListNull(types.StringType)
	}

	if len(user.DnsServers) > 0 {
		dnsServersAttr := make([]attr.Value, len(user.DnsServers))
		for i, dnsServer := range user.DnsServers {
			dnsServersAttr[i] = types.StringValue(dnsServer)
		}
		data.DNSServers, _ = types.ListValue(types.StringType, dnsServersAttr)
	} else {
		data.DNSServers = types.ListNull(types.StringType)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state UserResourceModel

	// Get both plan and current state for comparison
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Start with basic user data that's always updated
	user := pritunl.User{
		Name:            plan.Name.ValueString(),
		Organization:    plan.OrganizationID.ValueString(),
		Email:           plan.Email.ValueString(),
		Disabled:        plan.Disabled.ValueBool(),
		ClientToClient:  plan.ClientToClient.ValueBool(),
		AuthType:        plan.AuthType.ValueString(),
		DnsSuffix:       plan.DNSSuffix.ValueString(),
		BypassSecondary: plan.BypassSecondary.ValueBool(),
	}

	// Only update groups if they have changed
	if !plan.Groups.Equal(state.Groups) {
		if !plan.Groups.IsNull() && !plan.Groups.IsUnknown() {
			var groups []string
			resp.Diagnostics.Append(plan.Groups.ElementsAs(ctx, &groups, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
			user.Groups = groups
		}
	}

	// Only update network links if they have changed
	if !plan.NetworkLinks.Equal(state.NetworkLinks) {
		if !plan.NetworkLinks.IsNull() && !plan.NetworkLinks.IsUnknown() {
			var networkLinks []string
			resp.Diagnostics.Append(plan.NetworkLinks.ElementsAs(ctx, &networkLinks, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
			user.NetworkLinks = networkLinks
		}
	}

	// Only update MAC addresses if they have changed
	if !plan.MacAddresses.Equal(state.MacAddresses) {
		if !plan.MacAddresses.IsNull() && !plan.MacAddresses.IsUnknown() {
			var macAddresses []string
			resp.Diagnostics.Append(plan.MacAddresses.ElementsAs(ctx, &macAddresses, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
			user.MacAddresses = macAddresses
		}
	}

	// Only update DNS servers if they have changed
	if !plan.DNSServers.Equal(state.DNSServers) {
		if !plan.DNSServers.IsNull() && !plan.DNSServers.IsUnknown() {
			var dnsServers []string
			resp.Diagnostics.Append(plan.DNSServers.ElementsAs(ctx, &dnsServers, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
			user.DnsServers = dnsServers
		}
	}

	err := r.client.UpdateUser(plan.ID.ValueString(), &user)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update user, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *UserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data UserResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteUser(data.ID.ValueString(), data.OrganizationID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete user, got error: %s", err))
		return
	}
}

func (r *UserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, "-")
	if len(idParts) != 2 {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: organization_id-user_id. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[1])...)
}
