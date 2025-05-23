package security

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	regex2 "github.com/dlclark/regexp2"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory"
	"github.com/jfrog/terraform-provider-shared/util"
	utilfw "github.com/jfrog/terraform-provider-shared/util/fw"
	"github.com/samber/lo"
	"golang.org/x/exp/slices"
)

func NewScopedTokenResource() resource.Resource {
	return &ScopedTokenResource{
		TypeName: "artifactory_scoped_token",
	}
}

type ScopedTokenResource struct {
	ProviderData util.ProviderMetadata
	TypeName     string
}

// ScopedTokenResourceModel describes the Terraform resource data model to match the
// resource schema.
type ScopedTokenResourceModelV0 struct {
	Id                    types.String `tfsdk:"id"`
	GrantType             types.String `tfsdk:"grant_type"`
	Username              types.String `tfsdk:"username"`
	ProjectKey            types.String `tfsdk:"project_key"`
	Scopes                types.Set    `tfsdk:"scopes"`
	ExpiresIn             types.Int64  `tfsdk:"expires_in"`
	Refreshable           types.Bool   `tfsdk:"refreshable"`
	IncludeReferenceToken types.Bool   `tfsdk:"include_reference_token"`
	Description           types.String `tfsdk:"description"`
	Audiences             types.Set    `tfsdk:"audiences"`
	AccessToken           types.String `tfsdk:"access_token"`
	RefreshToken          types.String `tfsdk:"refresh_token"`
	ReferenceToken        types.String `tfsdk:"reference_token"`
	TokenType             types.String `tfsdk:"token_type"`
	Subject               types.String `tfsdk:"subject"`
	Expiry                types.Int64  `tfsdk:"expiry"`
	IssuedAt              types.Int64  `tfsdk:"issued_at"`
	Issuer                types.String `tfsdk:"issuer"`
}

type ScopedTokenResourceModel struct {
	Id                        types.String `tfsdk:"id"`
	GrantType                 types.String `tfsdk:"grant_type"`
	Username                  types.String `tfsdk:"username"`
	ProjectKey                types.String `tfsdk:"project_key"`
	Scopes                    types.Set    `tfsdk:"scopes"`
	ExpiresIn                 types.Int64  `tfsdk:"expires_in"`
	Refreshable               types.Bool   `tfsdk:"refreshable"`
	IncludeReferenceToken     types.Bool   `tfsdk:"include_reference_token"`
	Description               types.String `tfsdk:"description"`
	Audiences                 types.Set    `tfsdk:"audiences"`
	AccessToken               types.String `tfsdk:"access_token"`
	RefreshToken              types.String `tfsdk:"refresh_token"`
	ReferenceToken            types.String `tfsdk:"reference_token"`
	TokenType                 types.String `tfsdk:"token_type"`
	Subject                   types.String `tfsdk:"subject"`
	Expiry                    types.Int64  `tfsdk:"expiry"`
	IssuedAt                  types.Int64  `tfsdk:"issued_at"`
	Issuer                    types.String `tfsdk:"issuer"`
	IgnoreMissingTokenWarning types.Bool   `tfsdk:"ignore_missing_token_warning"`
}

type AccessTokenPostResponseAPIModel struct {
	TokenId        string `json:"token_id"`
	AccessToken    string `json:"access_token"`
	RefreshToken   string `json:"refresh_token"`
	ExpiresIn      int64  `json:"expires_in"`
	Scope          string `json:"scope"`
	TokenType      string `json:"token_type"`
	ReferenceToken string `json:"reference_token"`
}

type AccessTokenErrorResponseAPIModel struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail"`
}

type AccessTokenPostRequestAPIModel struct {
	GrantType             string `json:"grant_type"`
	Username              string `json:"username,omitempty"`
	ProjectKey            string `json:"project_key"`
	Scope                 string `json:"scope,omitempty"`
	ExpiresIn             int64  `json:"expires_in"`
	Refreshable           bool   `json:"refreshable"`
	Description           string `json:"description,omitempty"`
	Audience              string `json:"audience,omitempty"`
	IncludeReferenceToken bool   `json:"include_reference_token"`
}

type AccessTokenGetAPIModel struct {
	TokenId     string `json:"token_id"`
	Subject     string `json:"subject"`
	Expiry      int64  `json:"expiry"`
	IssuedAt    int64  `json:"issued_at"`
	Issuer      string `json:"issuer"`
	Description string `json:"description"`
	Refreshable bool   `json:"refreshable"`
}

func (r *ScopedTokenResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = r.TypeName
}

var schemaAttributesV0 = map[string]schema.Attribute{
	"id": schema.StringAttribute{
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"grant_type": schema.StringAttribute{
		MarkdownDescription: "The grant type used to authenticate the request. In this case, the only value supported is `client_credentials` which is also the default value if this parameter is not specified.",
		Optional:            true,
		Computed:            true,
		Default:             stringdefault.StaticString("client_credentials"),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplaceIfConfigured(),
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"username": schema.StringAttribute{
		MarkdownDescription: "The user name for which this token is created. The username is based " +
			"on the authenticated user - either from the user of the authenticated token or based " +
			"on the username (if basic auth was used). The username is then used to set the subject " +
			"of the token: <service-id>/users/<username>. Limited to 255 characters.",
		Optional:      true,
		PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplaceIfConfigured()},
		Validators:    []validator.String{stringvalidator.LengthBetween(1, 255)},
	},
	"project_key": schema.StringAttribute{
		MarkdownDescription: "The project for which this token is created. Enter the project name on which you want to apply this token.",
		Optional:            true,
		Validators: []validator.String{
			stringvalidator.RegexMatches(
				regexp.MustCompile(`^^[a-z][a-z0-9\-]{1,31}$`),
				"must be 2 - 32 lowercase alphanumeric and hyphen characters",
			),
		},
	},
	"scopes": schema.SetAttribute{
		MarkdownDescription: "The scope of access that the token provides. Access to the REST API is always " +
			"provided by default. Administrators can set any scope, while non-admin users can only set " +
			"the scope to a subset of the groups to which they belong. The supported scopes include:\n" +
			"  - `applied-permissions/user` - provides user access. If left at the default setting, the " +
			"token will be created with the user-identity scope, which allows users to identify themselves " +
			"in the Platform but does not grant any specific access permissions.\n" +
			"  - `applied-permissions/admin` - the scope assigned to admin users.\n" +
			"  - `applied-permissions/groups` - this scope assigns permissions to groups using the following format: `applied-permissions/groups:<group-name>[,<group-name>...]`\n" +
			"  - Resource Permissions: From Artifactory 7.38.x, resource permissions scoped tokens are also supported in the REST API. " +
			"A permission can be represented as a scope token string in the following format: `<resource-type>:<target>[/<sub-resource>]:<actions>`\n" +
			"    - Where:\n" +
			"      - `<resource-type>` - one of the permission resource types, from a predefined closed list. " +
			"Currently, the only resource type that is supported is the artifact resource type.\n" +
			"      - `<target>` - the target resource, can be exact name or a pattern\n" +
			"      - `<sub-resource>` - optional, the target sub-resource, can be exact name or a pattern\n" +
			"      - `<actions>` - comma-separated list of action acronyms. " +
			"The actions allowed are `r`, `w`, `d`, `a`, `m`, `x`, `s`, or any combination of these actions. To allow all actions - use `*`\n" +
			"    - Examples:\n" +
			"      - `[\"applied-permissions/user\", \"artifact:generic-local:r\"]`\n" +
			"      - `[\"applied-permissions/group\", \"artifact:generic-local/path:*\"]`\n" +
			"      - `[\"applied-permissions/admin\", \"system:metrics:r\", \"artifact:generic-local:*\"]`\n" +
			"  - `applied-permissions/roles:project-key` - provides access to elements associated with the project based on the project role. For example, `applied-permissions/roles:project-type:developer,qa`.\n\n" +
			"  - System Permissions: Used to grant access to system resources. " +
			"A permission can be represented as a scope token string in the following format: `system:(metrics|livelogs|identities|permissions):<actions>`\n" +
			"    - Where:\n" +
			"      - `metrics|livelogs|identities|permissions` - one of these options can be chosen" +
			"      - `<actions>` - comma-separated list of action acronyms. " +
			"The actions allowed are `r`, `w`, `d`, `a`, `m`, `x`, `s`, or any combination of these actions. To allow all actions - use `*`\n" +
			"    - Examples:\n" +
			"      - `[\"system:livelogs:r\", \"system:metrics:r,w,d\"]`\n" +
			"->The scope to assign to the token should be provided as a list of scope tokens, limited to 500 characters in total.\n" +
			"From Artifactory 7.84.3, [project admins](https://jfrog.com/help/r/jfrog-platform-administration-documentation/access-token-creation-by-project-admins) can create access tokens that are tied to the projects in which they hold administrative privileges.",
		Optional:    true,
		Computed:    true,
		ElementType: types.StringType,
		PlanModifiers: []planmodifier.Set{
			setplanmodifier.RequiresReplaceIfConfigured(),
			setplanmodifier.UseStateForUnknown(),
		},
		Validators: []validator.Set{
			setvalidator.ValueStringsAre(
				stringvalidator.Any(
					stringvalidator.OneOf(
						"applied-permissions/user",
						"applied-permissions/admin",
					),
					stringvalidator.RegexMatches(regexp.MustCompile(`^applied-permissions\/groups:.+$`), "must be 'applied-permissions/groups:<group-name>[,<group-name>...]'"),
					stringvalidator.RegexMatches(regexp.MustCompile(`^applied-permissions\/roles:.+:.+$`), "must be 'applied-permissions/roles:<project-key>:<role-name>[,<role-name>...]'"),
					stringvalidator.RegexMatches(regexp.MustCompile(`^artifact:(?:.+):(?:(?:[rwdamxs*]+)|(?:[rwdamxs]+)(?:,[rwdamxs]+)+)$`), "must be '<resource-type>:<target>[/<sub-resource>]:<actions>'"),
					stringvalidator.RegexMatches(regexp.MustCompile(`^system:(?:metrics|livelogs|identities|permissions):(?:(?:[rwdamxs*]+)|(?:[rwdamxs]+)(?:,[rwdamxs]+)+)$`), "must be 'system:(metrics|livelogs|identities|permissions):<actions>'"),
				),
			),
		},
	},
	"expires_in": schema.Int64Attribute{
		MarkdownDescription: "The amount of time, in seconds, it would take for the token to expire. An admin shall be able to set whether expiry is mandatory, what is the default expiry, and what is the maximum expiry allowed. Must be non-negative. Default value is based on configuration in 'access.config.yaml'. See [API documentation](https://jfrog.com/help/r/jfrog-rest-apis/revoke-token-by-id) for details. Access Token would not be saved by Artifactory if this is less than the persistence threshold value (default to 10800 seconds) set in Access configuration. See [official documentation](https://jfrog.com/help/r/jfrog-platform-administration-documentation/persistency-threshold) for details.",
		Optional:            true,
		Computed:            true,
		PlanModifiers: []planmodifier.Int64{
			int64planmodifier.RequiresReplaceIfConfigured(),
			int64planmodifier.UseStateForUnknown(),
		},
		Validators: []validator.Int64{int64validator.AtLeast(0)},
	},
	"refreshable": schema.BoolAttribute{
		MarkdownDescription: "Is this token refreshable? Default is `false`.",
		Optional:            true,
		Computed:            true,
		Default:             booldefault.StaticBool(false),
		PlanModifiers: []planmodifier.Bool{
			boolplanmodifier.RequiresReplaceIfConfigured(),
			boolplanmodifier.UseStateForUnknown(),
		},
	},
	"include_reference_token": schema.BoolAttribute{
		MarkdownDescription: "Also create a reference token which can be used like an API key. Default is `false`.",
		Optional:            true,
		Computed:            true,
		PlanModifiers: []planmodifier.Bool{
			boolplanmodifier.RequiresReplaceIfConfigured(),
			boolplanmodifier.UseStateForUnknown(),
		},
	},
	"description": schema.StringAttribute{
		MarkdownDescription: "Free text token description. Useful for filtering and managing tokens. Limited to 1024 characters.",
		Optional:            true,
		Computed:            true,
		Default:             stringdefault.StaticString(""),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplaceIfConfigured(),
			stringplanmodifier.UseStateForUnknown(),
		},
		Validators: []validator.String{stringvalidator.LengthBetween(0, 1024)},
	},
	"audiences": schema.SetAttribute{
		MarkdownDescription: "A list of the other instances or services that should accept this " +
			"token identified by their Service-IDs. Limited to total 255 characters. " +
			"Default to '*@*' if not set. Service ID must begin with valid JFrog service type. " +
			"Options: jfrt, jfxr, jfpip, jfds, jfmc, jfac, jfevt, jfmd, jfcon, or *. For instructions to retrieve the Artifactory Service ID see this [documentation](https://jfrog.com/help/r/jfrog-rest-apis/get-service-id)",
		Optional:    true,
		ElementType: types.StringType,
		PlanModifiers: []planmodifier.Set{
			setplanmodifier.RequiresReplaceIfConfigured(),
			setplanmodifier.UseStateForUnknown(),
		},
		Validators: []validator.Set{
			setvalidator.ValueStringsAre(
				stringvalidator.All(
					stringvalidator.LengthAtLeast(1),
					stringvalidator.RegexMatches(regexp.MustCompile(fmt.Sprintf(`^(%s|\*)@.+`, strings.Join(serviceTypesScopedToken, "|"))),
						fmt.Sprintf(
							"must either begin with %s, or *",
							strings.Join(serviceTypesScopedToken, ", "),
						),
					),
				),
			),
		},
	},
	"access_token": schema.StringAttribute{
		MarkdownDescription: "Returns the access token to authenticate to Artifactory.",
		Sensitive:           true,
		Computed:            true,
		PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
	},
	"refresh_token": schema.StringAttribute{
		MarkdownDescription: "Refresh token.",
		Sensitive:           true,
		Computed:            true,
		PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
	},
	"reference_token": schema.StringAttribute{
		MarkdownDescription: "Reference Token (alias to Access Token).",
		Sensitive:           true,
		Computed:            true,
		PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
	},
	"token_type": schema.StringAttribute{
		MarkdownDescription: "Returns the token type.",
		Computed:            true,
		PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
	},
	"subject": schema.StringAttribute{
		MarkdownDescription: "Returns the token type.",
		Computed:            true,
		PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
	},
	"expiry": schema.Int64Attribute{
		MarkdownDescription: "Returns the token expiry.",
		Computed:            true,
		PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
	},
	"issued_at": schema.Int64Attribute{
		MarkdownDescription: "Returns the token issued at date/time.",
		Computed:            true,
		PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
	},
	"issuer": schema.StringAttribute{
		MarkdownDescription: "Returns the token issuer.",
		Computed:            true,
		PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
	},
}

var schemaAttributesV1 = lo.Assign(
	schemaAttributesV0,
	map[string]schema.Attribute{
		"ignore_missing_token_warning": schema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Default:     booldefault.StaticBool(false),
			Description: "Toggle to ignore warning message when token was missing or not created and stored by Artifactory. Default is `false`.",
		},
	},
)

func (r *ScopedTokenResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Create scoped tokens for any of the services in your JFrog Platform and to " +
			"manage user access to these services. If left at the default setting, the token will " +
			"be created with the user-identity scope, which allows users to identify themselves in " +
			"the Platform but does not grant any specific access permissions.",
		Attributes: schemaAttributesV1,
		Version:    1,
	}
}

func (r *ScopedTokenResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		// State upgrade implementation from 0 (prior state version) to 1 (Schema.Version)
		0: {
			PriorSchema: &schema.Schema{
				Attributes: schemaAttributesV0,
			},
			// Optionally, the PriorSchema field can be defined.
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				var priorStateData ScopedTokenResourceModelV0

				resp.Diagnostics.Append(req.State.Get(ctx, &priorStateData)...)
				if resp.Diagnostics.HasError() {
					return
				}

				upgradedStateData := ScopedTokenResourceModel{
					Id:                        priorStateData.Id,
					GrantType:                 priorStateData.GrantType,
					Username:                  priorStateData.Username,
					ProjectKey:                priorStateData.ProjectKey,
					Scopes:                    priorStateData.Scopes,
					ExpiresIn:                 priorStateData.ExpiresIn,
					Refreshable:               priorStateData.Refreshable,
					IncludeReferenceToken:     priorStateData.IncludeReferenceToken,
					Description:               priorStateData.Description,
					Audiences:                 priorStateData.Audiences,
					AccessToken:               priorStateData.AccessToken,
					RefreshToken:              priorStateData.RefreshToken,
					ReferenceToken:            priorStateData.ReferenceToken,
					TokenType:                 priorStateData.TokenType,
					Subject:                   priorStateData.Subject,
					Expiry:                    priorStateData.Expiry,
					IssuedAt:                  priorStateData.IssuedAt,
					Issuer:                    priorStateData.Issuer,
					IgnoreMissingTokenWarning: types.BoolValue(false),
				}

				resp.Diagnostics.Append(resp.State.Set(ctx, upgradedStateData)...)
			},
		},
	}
}

var serviceTypesScopedToken = []string{"jfrt", "jfxr", "jfpip", "jfds", "jfmc", "jfac", "jfevt", "jfmd", "jfcon"}

func (r *ScopedTokenResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	r.ProviderData = req.ProviderData.(util.ProviderMetadata)
}

func (r *ScopedTokenResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	go util.SendUsageResourceCreate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan *ScopedTokenResourceModel
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	scopes := []string{}
	if !plan.Scopes.IsNull() {
		scopes = utilfw.StringSetToStrings(plan.Scopes)
	}
	scopesString := strings.Join(scopes, " ") // Join slice into space-separated string
	if len(scopesString) > 500 {
		resp.Diagnostics.AddError(
			"Scopes length exceeds 500 characters",
			"total combined length of scopes field exceeds 500 characters:"+scopesString,
		)
		return
	}

	audiences := []string{}
	if !plan.Audiences.IsNull() {
		audiences = utilfw.StringSetToStrings(plan.Audiences)
	}
	audiencesString := strings.Join(audiences, " ") // Join slice into space-separated string
	if len(audiencesString) > 255 {
		resp.Diagnostics.AddError(
			"Audiences length exceeds 255 characters",
			"total combined length of audiences field exceeds 255 characters:"+audiencesString,
		)
		return
	}

	// Convert from Terraform data model into API data model
	accessTokenPostBody := AccessTokenPostRequestAPIModel{
		GrantType:             plan.GrantType.ValueString(),
		Username:              plan.Username.ValueString(),
		ProjectKey:            plan.ProjectKey.ValueString(),
		Scope:                 scopesString,
		ExpiresIn:             plan.ExpiresIn.ValueInt64(),
		Refreshable:           plan.Refreshable.ValueBool(),
		Description:           plan.Description.ValueString(),
		Audience:              audiencesString,
		IncludeReferenceToken: plan.IncludeReferenceToken.ValueBool(),
	}

	postResult := AccessTokenPostResponseAPIModel{}

	var artifactoryError artifactory.ArtifactoryErrorsResponse
	response, err := r.ProviderData.Client.R().
		SetBody(accessTokenPostBody).
		SetResult(&postResult).
		SetError(&artifactoryError).
		Post("access/api/v1/tokens")

	if err != nil {
		utilfw.UnableToCreateResourceError(resp, err.Error())
		return
	}

	// Return error if the HTTP status code is not 200 OK
	if response.StatusCode() != http.StatusOK {
		utilfw.UnableToCreateResourceError(resp, response.String())
		return
	}

	if response.IsError() {
		utilfw.UnableToCreateResourceError(resp, artifactoryError.String())
		return
	}

	getResult := AccessTokenGetAPIModel{}
	id := types.StringValue(postResult.TokenId)

	response, err = r.ProviderData.Client.R().
		SetPathParam("id", id.ValueString()).
		SetResult(&getResult).
		SetError(&artifactoryError).
		Get("access/api/v1/tokens/{id}")

	if err != nil {
		utilfw.UnableToCreateResourceError(resp, err.Error())
		return
	}

	if response.IsError() {
		if response.StatusCode() == http.StatusNotFound {
			resp.Diagnostics.AddWarning(
				fmt.Sprintf("Scoped token with ID %s is not found", id.ValueString()),
				"Token would not be saved by Artifactory if 'expires_in' is less than the persistence threshold value (default to 10800 seconds) set in Access configuration. "+
					"See https://jfrog.com/help/r/jfrog-platform-administration-documentation/persistency-threshold for details.",
			)
		} else {
			utilfw.UnableToCreateResourceError(resp, artifactoryError.String())
			return
		}
	}

	// Assign the attribute values for the resource in the state
	resp.Diagnostics.Append(plan.PostResponseToState(ctx, &postResult, &accessTokenPostBody, &getResult)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...) // All attributes are assigned in data
}

func (r *ScopedTokenResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	go util.SendUsageResourceRead(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state *ScopedTokenResourceModel
	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert from Terraform data model into API data model
	var accessToken AccessTokenGetAPIModel

	var artifactoryError artifactory.ArtifactoryErrorsResponse
	response, err := r.ProviderData.Client.R().
		SetPathParam("id", state.Id.ValueString()).
		SetResult(&accessToken).
		SetError(&artifactoryError).
		Get("access/api/v1/tokens/{id}")

	// Treat HTTP 404 Not Found status as a signal to recreate resource
	// and return early
	if err != nil {
		utilfw.UnableToRefreshResourceError(resp, err.Error())
		return
	}

	if response.StatusCode() == http.StatusNotFound {
		if !state.IgnoreMissingTokenWarning.ValueBool() {
			resp.Diagnostics.AddWarning(
				fmt.Sprintf("Scoped token %s not found or not created", state.Id.ValueString()),
				"Access Token would not be saved by Artifactory if 'expires_in' is less than the persistence threshold value (default to 10800 seconds) set in Access configuration. See https://www.jfrog.com/confluence/display/JFROG/Access+Tokens#AccessTokens-PersistencyThreshold for details."+response.String(),
			)
		}
		resp.State.RemoveResource(ctx)
		return
	}

	if response.IsError() {
		utilfw.UnableToRefreshResourceError(resp, artifactoryError.String())
		return
	}

	// Convert from the API data model to the Terraform data model
	// and refresh any attribute values.
	state.GetResponseToState(ctx, &accessToken)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ScopedTokenResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	go util.SendUsageResourceUpdate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan *ScopedTokenResourceModel
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	//
	// We only care about updating state for 'ignore_missing_token_warning' attribute
	// All other attributes should trigger a recreation instead
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ScopedTokenResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	go util.SendUsageResourceDelete(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state ScopedTokenResourceModel
	respError := AccessTokenErrorResponseAPIModel{}

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	id := state.Id.ValueString()

	response, err := r.ProviderData.Client.R().
		SetPathParam("id", id).
		SetError(&respError).
		Delete("access/api/v1/tokens/{id}")

	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Failed to revoke scoped token %s", id),
			"An unexpected error occurred while attempting to delete the resource. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"HTTP Error: "+err.Error(),
		)

		return
	}

	if response.IsError() {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Failed to revoke scoped token %s", id),
			"An unexpected error occurred while attempting to delete the resource. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"HTTP Error: "+respError.Message,
		)

		return
	}
	// If the logic reaches here, it implicitly succeeded and will remove
	// the resource from state if there are no other errors.
}

// ImportState imports the resource into the Terraform state.
func (r *ScopedTokenResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.AddError(
		"Import is not supported",
		"resource artifactory_scoped_token doesn't support import.",
	)
}

// splitScopes use positive lookahead regex to find the space character between scopes
// but ignore group name with space wraps in double quotes
func (r *ScopedTokenResourceModel) splitScopes(ctx context.Context, scopes string) []string {
	if scopes == "" {
		return []string{}
	}

	// regexp doesn't support lookahead so we have to import regex2 lib which does
	re := regex2.MustCompile(`\s(?=(?:[^"]*"[^"]*")*[^"]*$)`, 0)
	match, err := re.FindStringMatch(scopes)
	if err != nil {
		tflog.Warn(ctx, "fail to find scopes match", map[string]any{
			"err": err,
		})
	}

	// return the entire 'scopes' string if there's no space separator, i.e. only one item in the list
	if match == nil {
		return []string{scopes}
	}

	separatorIndices := []int{}
	// collect all the indices for the space delimited scope string
	for ok := true; ok; ok = (match != nil) { // mimic do...while loop
		for _, g := range match.Groups() {
			for _, c := range g.Captures {
				separatorIndices = append(separatorIndices, c.Index)
			}
		}

		match, err = re.FindNextMatch(match)
		if err != nil {
			tflog.Warn(ctx, "fail to find next scopes match", map[string]any{
				"err": err,
			})
		}
	}

	// insert a zero to the begining of the slice to represent the first index
	separatorIndices = append([]int{0}, separatorIndices...)
	// reverse the slice so the string splitting starts from the end
	slices.Reverse(separatorIndices)

	scopesCopy := scopes
	scopesList := []string{}
	for _, idx := range separatorIndices {
		// pad the start index by 1 to take care of the space prefix character
		startIdx := idx + 1
		if idx == 0 {
			startIdx = 0
		}
		scopesList = append(scopesList, scopesCopy[startIdx:])
		// trim the end of string off for next iteration
		scopesCopy = scopesCopy[:idx]
	}

	return scopesList
}

func (r *ScopedTokenResourceModel) PostResponseToState(ctx context.Context,
	accessTokenResp *AccessTokenPostResponseAPIModel, accessTokenPostBody *AccessTokenPostRequestAPIModel, getResult *AccessTokenGetAPIModel) diag.Diagnostics {

	r.Id = types.StringValue(accessTokenResp.TokenId)

	if len(accessTokenResp.Scope) > 0 {
		scopesList := r.splitScopes(ctx, accessTokenResp.Scope)
		scopes, diags := types.SetValueFrom(ctx, types.StringType, scopesList)
		if diags != nil {
			return diags
		}

		r.Scopes = scopes
	}

	r.ExpiresIn = types.Int64Value(accessTokenResp.ExpiresIn)

	r.AccessToken = types.StringValue(accessTokenResp.AccessToken)

	// only have refresh token if 'refreshable' is set to true in the request
	r.RefreshToken = types.StringNull()
	if accessTokenPostBody.Refreshable && len(accessTokenResp.RefreshToken) > 0 {
		r.RefreshToken = types.StringValue(accessTokenResp.RefreshToken)
	}

	// only have reference token if 'include_reference_token' is set to true in the request
	r.ReferenceToken = types.StringNull()
	if accessTokenPostBody.IncludeReferenceToken && len(accessTokenResp.ReferenceToken) > 0 {
		r.ReferenceToken = types.StringValue(accessTokenResp.ReferenceToken)
	}

	r.IncludeReferenceToken = types.BoolValue(accessTokenPostBody.IncludeReferenceToken)
	r.TokenType = types.StringValue(accessTokenResp.TokenType)
	r.Subject = types.StringValue(getResult.Subject)
	r.Expiry = types.Int64Value(getResult.Expiry) // could be absent in the get response!
	r.IssuedAt = types.Int64Value(getResult.IssuedAt)
	r.Issuer = types.StringValue(getResult.Issuer)

	return nil
}

func (r *ScopedTokenResourceModel) GetResponseToState(ctx context.Context, accessToken *AccessTokenGetAPIModel) {
	r.Id = types.StringValue(accessToken.TokenId)
	if r.GrantType.IsNull() {
		r.GrantType = types.StringValue("client_credentials")
	}
	r.Subject = types.StringValue(accessToken.Subject)
	r.Expiry = types.Int64Value(accessToken.Expiry)
	r.IssuedAt = types.Int64Value(accessToken.IssuedAt)
	r.Issuer = types.StringValue(accessToken.Issuer)

	if r.Description.IsNull() {
		r.Description = types.StringValue("")
	}
	if len(accessToken.Description) > 0 {
		r.Description = types.StringValue(accessToken.Description)
	}

	r.Refreshable = types.BoolValue(accessToken.Refreshable)

	// Need to set empty string for null state value to avoid state drift.
	// See https://discuss.hashicorp.com/t/diffsuppressfunc-alternative-in-terraform-framework/52578/2
	if r.RefreshToken.IsNull() {
		r.RefreshToken = types.StringValue("")
	}
	if r.ReferenceToken.IsNull() {
		r.ReferenceToken = types.StringValue("")
	}
}
