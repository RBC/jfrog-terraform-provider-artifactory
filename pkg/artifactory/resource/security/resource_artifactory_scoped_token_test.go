package security_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

func TestAccScopedToken_UpgradeFromSDKv2(t *testing.T) {
	providerHost := os.Getenv("TF_ACC_PROVIDER_HOST")
	if providerHost == "registry.opentofu.org" {
		t.Skipf("provider host is registry.opentofu.org. Previous version of Artifactory provider is unknown to OpenTofu.")
	}

	// Version 7.11.1 is the last version before we migrated the resource from SDKv2 to Plugin Framework
	version := "7.11.1"
	title := fmt.Sprintf("from_v%s", version)
	t.Run(title, func(t *testing.T) {
		resource.Test(scopedTokenUpgradeTestCase(version, t))
	})
}

func TestAccScopedToken_UpgradeGH_758(t *testing.T) {
	providerHost := os.Getenv("TF_ACC_PROVIDER_HOST")
	if providerHost == "registry.opentofu.org" {
		t.Skipf("provider host is registry.opentofu.org. Previous version of Artifactory provider is unknown to OpenTofu.")
	}

	// Version 7.2.0 doesn't have `include_reference_token` attribute
	// This test verifies that there is no state drift on update
	version := "7.2.0"
	title := fmt.Sprintf("from_v%s", version)
	t.Run(title, func(t *testing.T) {
		resource.Test(scopedTokenUpgradeTestCase(version, t))
	})
}

func TestAccScopedToken_UpgradeGH_792(t *testing.T) {
	providerHost := os.Getenv("TF_ACC_PROVIDER_HOST")
	if providerHost == "registry.opentofu.org" {
		t.Skipf("provider host is registry.opentofu.org. Previous version of Artifactory provider is unknown to OpenTofu.")
	}

	_, fqrn, name := testutil.MkNames("test-access-token", "artifactory_scoped_token")
	config := util.ExecuteTemplate(
		"TestAccScopedToken",
		`resource "artifactory_user" "test-user" {
			name              = "testuser"
		    email             = "testuser@tempurl.org"
			admin             = true
			disable_ui_access = false
			groups            = ["readers"]
			password          = "Passw0rd!"
		}

		resource "artifactory_scoped_token" "{{ .name }}" {
			username    = artifactory_user.test-user.name
		    expires_in  = 31536000
		}`,
		map[string]interface{}{
			"name": name,
		},
	)

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"artifactory": {
						VersionConstraint: "7.11.2",
						Source:            "jfrog/artifactory",
					},
				},
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "username", "testuser"),
					resource.TestCheckNoResourceAttr(fqrn, "description"),
					resource.TestCheckResourceAttr(fqrn, "scopes.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "expires_in", "31536000"),
					resource.TestCheckNoResourceAttr(fqrn, "audiences"),
					resource.TestCheckResourceAttrSet(fqrn, "access_token"),
					resource.TestCheckNoResourceAttr(fqrn, "refresh_token"),
					resource.TestCheckNoResourceAttr(fqrn, "reference_token"),
					resource.TestCheckResourceAttr(fqrn, "token_type", "Bearer"),
					resource.TestCheckResourceAttrSet(fqrn, "subject"),
					resource.TestCheckResourceAttrSet(fqrn, "expiry"),
					resource.TestCheckResourceAttrSet(fqrn, "issued_at"),
					resource.TestCheckResourceAttrSet(fqrn, "issuer"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
				Config:                   config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccScopedToken_UpgradeGH_818(t *testing.T) {
	providerHost := os.Getenv("TF_ACC_PROVIDER_HOST")
	if providerHost == "registry.opentofu.org" {
		t.Skipf("provider host is registry.opentofu.org. Previous version of Artifactory provider is unknown to OpenTofu.")
	}

	_, fqrn, name := testutil.MkNames("test-scope-token", "artifactory_scoped_token")
	config := util.ExecuteTemplate(
		"TestAccScopedToken",
		`resource "artifactory_user" "test-user" {
			name              = "testuser"
		    email             = "testuser@tempurl.org"
			admin             = true
			disable_ui_access = false
			groups            = ["readers"]
			password          = "Passw0rd!"
		}

		resource "artifactory_scoped_token" "{{ .name }}" {
			scopes   = ["applied-permissions/user"]
			username = artifactory_user.test-user.name
		}`,
		map[string]interface{}{
			"name": name,
		},
	)

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"artifactory": {
						VersionConstraint: "7.2.0",
						Source:            "jfrog/artifactory",
					},
				},
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "username", "testuser"),
					resource.TestCheckResourceAttr(fqrn, "scopes.#", "1"),
					resource.TestCheckResourceAttrSet(fqrn, "expires_in"),
					resource.TestCheckNoResourceAttr(fqrn, "audiences"),
					resource.TestCheckResourceAttrSet(fqrn, "access_token"),
					resource.TestCheckNoResourceAttr(fqrn, "refresh_token"),
					resource.TestCheckNoResourceAttr(fqrn, "reference_token"),
					resource.TestCheckResourceAttr(fqrn, "token_type", "Bearer"),
					resource.TestCheckResourceAttrSet(fqrn, "subject"),
					resource.TestCheckResourceAttrSet(fqrn, "expiry"),
					resource.TestCheckResourceAttrSet(fqrn, "issued_at"),
					resource.TestCheckResourceAttrSet(fqrn, "issuer"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
				Config:                   config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccScopedToken_UpgradeToV1Schema(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test-scope-token", "artifactory_scoped_token")

	_, _, username := testutil.MkNames("test-user", "artifactory_user")

	config := util.ExecuteTemplate(
		"TestAccScopedToken",
		`resource "artifactory_user" "{{ .username }}" {
			name              = "{{ .username }}"
		    email             = "{{ .username }}@tempurl.org"
			admin             = true
			disable_ui_access = false
			groups            = ["readers"]
			password          = "Passw0rd!"
		}

		resource "artifactory_scoped_token" "{{ .name }}" {
			scopes   = ["applied-permissions/user"]
			username = artifactory_user.{{ .username }}.name
		}`,
		map[string]interface{}{
			"name":     name,
			"username": username,
		},
	)

	updatedConfig := util.ExecuteTemplate(
		"TestAccScopedToken",
		`resource "artifactory_user" "{{ .username }}" {
			name              = "{{ .username }}"
		    email             = "{{ .username }}@tempurl.org"
			admin             = true
			disable_ui_access = false
			groups            = ["readers"]
			password          = "Passw0rd!"
		}

		resource "artifactory_scoped_token" "{{ .name }}" {
			scopes   = ["applied-permissions/user"]
			username = artifactory_user.{{ .username }}.name
			ignore_missing_token_warning = true
		}`,
		map[string]interface{}{
			"name":     name,
			"username": username,
		},
	)

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"artifactory": {
						Source:            "jfrog/artifactory",
						VersionConstraint: "11.1.0",
					},
				},
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(fqrn, "ignore_missing_token_warning"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
				Config:                   updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "ignore_missing_token_warning", "true"),
				),
			},
		},
	})
}

func scopedTokenUpgradeTestCase(version string, t *testing.T) (*testing.T, resource.TestCase) {
	_, fqrn, name := testutil.MkNames("test-access-token", "artifactory_scoped_token")

	id, _, userResourceName := testutil.MkNames("test-user-", "artifactory_user")
	username := fmt.Sprintf("dummy_user%d", id)
	email := username + "@test.com"

	config := util.ExecuteTemplate(
		"TestAccScopedToken",
		`resource "artifactory_user" "{{ .user_resource_name }}" {
			name              = "{{ .username }}"
			email             = "{{ .email }}"
			admin             = true
			disable_ui_access = false
			groups            = ["readers"]
			password          = "Passw0rd!"
		}

		resource "artifactory_scoped_token" "{{ .name }}" {
			username    = artifactory_user.{{ .user_resource_name }}.name
		    expires_in  = 31536000
		}`,
		map[string]interface{}{
			"name":               name,
			"user_resource_name": userResourceName,
			"username":           username,
			"email":              email,
		},
	)

	return t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"artifactory": {
						VersionConstraint: version,
						Source:            "jfrog/artifactory",
					},
				},
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "username", username),
					resource.TestCheckResourceAttr(fqrn, "description", ""),
					resource.TestCheckResourceAttr(fqrn, "scopes.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "expires_in", "31536000"),
					resource.TestCheckNoResourceAttr(fqrn, "audiences"),
					resource.TestCheckResourceAttrSet(fqrn, "access_token"),
					resource.TestCheckNoResourceAttr(fqrn, "refresh_token"),
					resource.TestCheckNoResourceAttr(fqrn, "reference_token"),
					resource.TestCheckResourceAttr(fqrn, "token_type", "Bearer"),
					resource.TestCheckResourceAttrSet(fqrn, "subject"),
					resource.TestCheckResourceAttrSet(fqrn, "expiry"),
					resource.TestCheckResourceAttrSet(fqrn, "issued_at"),
					resource.TestCheckResourceAttrSet(fqrn, "issuer"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
				Config:                   config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	}
}

func TestAccScopedToken_WithDefaults(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test-access-token", "artifactory_scoped_token")

	id, _, userResourceName := testutil.MkNames("test-user-", "artifactory_user")
	username := fmt.Sprintf("dummy_user%d", id)
	email := username + "@test.com"

	template := `resource "artifactory_user" "{{ .user_resource_name }}" {
		name              = "{{ .username }}"
		email             = "{{ .email }}"
		admin             = true
		disable_ui_access = false
		groups            = ["readers"]
		password          = "Passw0rd!"
	}

	resource "artifactory_scoped_token" "{{ .name }}" {
		username    = artifactory_user.{{ .user_resource_name }}.name
		description = "{{ .description }}"
	}`

	accessTokenConfig := util.ExecuteTemplate(
		"TestAccScopedToken",
		template,
		map[string]interface{}{
			"name":               name,
			"description":        "",
			"user_resource_name": userResourceName,
			"username":           username,
			"email":              email,
		},
	)

	accessTokenUpdatedConfig := util.ExecuteTemplate(
		"TestAccScopedToken",
		template,
		map[string]interface{}{
			"name":               name,
			"description":        "test updated description",
			"user_resource_name": userResourceName,
			"username":           username,
			"email":              email,
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "", checkAccessToken),
		Steps: []resource.TestStep{
			{
				Config: accessTokenConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "username", username),
					resource.TestCheckResourceAttr(fqrn, "scopes.#", "1"),
					resource.TestCheckTypeSetElemAttr(fqrn, "scopes.*", "applied-permissions/user"),
					resource.TestCheckResourceAttr(fqrn, "refreshable", "false"),
					resource.TestCheckResourceAttr(fqrn, "description", ""),
					resource.TestCheckNoResourceAttr(fqrn, "audiences"),
					resource.TestCheckResourceAttrSet(fqrn, "access_token"),
					resource.TestCheckNoResourceAttr(fqrn, "refresh_token"),
					resource.TestCheckResourceAttr(fqrn, "token_type", "Bearer"),
					resource.TestCheckResourceAttrSet(fqrn, "subject"),
					resource.TestCheckResourceAttrSet(fqrn, "expiry"),
					resource.TestCheckResourceAttrSet(fqrn, "issued_at"),
					resource.TestCheckResourceAttrSet(fqrn, "issuer"),
					resource.TestCheckNoResourceAttr(fqrn, "reference_token"),
				),
			},
			{
				Config: accessTokenUpdatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "description", "test updated description"),
				),
			},
			{
				ResourceName: fqrn,
				ImportState:  true,
				ExpectError:  regexp.MustCompile("resource artifactory_scoped_token doesn't support import"),
			},
		},
	})
}

func TestAccScopedToken_WithAttributes(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test-access-token", "artifactory_scoped_token")
	_, _, projectKey := testutil.MkNames("test-project", "project")

	id, _, userResourceName := testutil.MkNames("test-user-", "artifactory_user")
	username := fmt.Sprintf("dummy_user%d", id)
	email := username + "@test.com"

	accessTokenConfig := util.ExecuteTemplate(
		"TestAccScopedToken",
		`resource "artifactory_user" "{{ .user_resource_name }}" {
			name              = "{{ .username }}"
			email             = "{{ .email }}"
			admin             = true
			disable_ui_access = false
			groups            = ["readers"]
			password          = "Passw0rd!"
		}

		resource "project" "{{ .projectKey }}" {
			key = "{{ .projectKey }}"
			display_name = "{{ .projectKey }}"
			admin_privileges {
				manage_members = true
				manage_resources = true
				index_resources = true
			}
		}

		resource "artifactory_scoped_token" "{{ .name }}" {
			username    = artifactory_user.{{ .user_resource_name }}.name
			project_key = project.{{ .projectKey }}.key
			scopes      = ["applied-permissions/admin", "system:metrics:r", "system:identities:r"]
			description = "test description"
			refreshable = true
			expires_in  = 0
			audiences   = ["jfrt@1", "jfxr@*"]
		}`,
		map[string]interface{}{
			"name":               name,
			"projectKey":         projectKey,
			"user_resource_name": userResourceName,
			"username":           username,
			"email":              email,
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"project": {
				Source: "jfrog/project",
			},
		},
		CheckDestroy: acctest.VerifyDeleted(t, fqrn, "", checkAccessToken),
		Steps: []resource.TestStep{
			{
				Config: accessTokenConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "username", username),
					resource.TestCheckResourceAttr(fqrn, "project_key", projectKey),
					resource.TestCheckResourceAttr(fqrn, "scopes.#", "3"),
					resource.TestCheckTypeSetElemAttr(fqrn, "scopes.*", "applied-permissions/admin"),
					resource.TestCheckTypeSetElemAttr(fqrn, "scopes.*", "system:metrics:r"),
					resource.TestCheckTypeSetElemAttr(fqrn, "scopes.*", "system:identities:r"),
					resource.TestCheckResourceAttr(fqrn, "refreshable", "true"),
					resource.TestCheckResourceAttr(fqrn, "expires_in", "0"),
					resource.TestCheckResourceAttr(fqrn, "description", "test description"),
					resource.TestCheckResourceAttr(fqrn, "audiences.#", "2"),
					resource.TestCheckTypeSetElemAttr(fqrn, "audiences.*", "jfrt@1"),
					resource.TestCheckTypeSetElemAttr(fqrn, "audiences.*", "jfxr@*"),
					resource.TestCheckResourceAttrSet(fqrn, "access_token"),
					resource.TestCheckResourceAttrSet(fqrn, "refresh_token"),
					resource.TestCheckNoResourceAttr(fqrn, "reference_token"),
					resource.TestCheckResourceAttr(fqrn, "token_type", "Bearer"),
					resource.TestCheckResourceAttrSet(fqrn, "subject"),
					resource.TestCheckResourceAttrSet(fqrn, "expiry"),
					resource.TestCheckResourceAttrSet(fqrn, "issued_at"),
					resource.TestCheckResourceAttrSet(fqrn, "issuer"),
					resource.TestCheckResourceAttr(fqrn, "include_reference_token", "false"),
				),
			},
			{
				ResourceName: fqrn,
				ImportState:  true,
				ExpectError:  regexp.MustCompile("resource artifactory_scoped_token doesn't support import"),
			},
		},
	})
}

func TestAccScopedToken_WithSingleGroupScope(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test-access-token", "artifactory_scoped_token")

	accessTokenConfig := util.ExecuteTemplate(
		"TestAccScopedToken",
		`resource "artifactory_group" "test-group-1" {
			name = "{{ .groupName1 }}"
		}

		resource "artifactory_scoped_token" "{{ .name }}" {
			username    = artifactory_group.test-group-1.name
			scopes      = [
				"applied-permissions/groups:{{ .groupName1 }}",
			]
		}`,
		map[string]interface{}{
			"name":       name,
			"groupName1": "test-group-1",
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accessTokenConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "username", "test-group-1"),
					resource.TestCheckResourceAttr(fqrn, "scopes.#", "1"),
					resource.TestCheckTypeSetElemAttr(fqrn, "scopes.*", "applied-permissions/groups:test-group-1"),
				),
			},
			{
				ResourceName: fqrn,
				ImportState:  true,
				ExpectError:  regexp.MustCompile("resource artifactory_scoped_token doesn't support import"),
			},
		},
	})
}

func TestAccScopedToken_WithMultipleGroupScopes(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test-access-token", "artifactory_scoped_token")

	accessTokenConfig := util.ExecuteTemplate(
		"TestAccScopedToken",
		`resource "artifactory_group" "test-group-1" {
			name = "{{ .groupName1 }}"
		}

		resource "artifactory_group" "test-group-2" {
			name = "{{ .groupName2 }}"
		}

		resource "artifactory_group" "test-group-3" {
			name = "{{ .groupName3 }}"
		}

		resource "artifactory_scoped_token" "{{ .name }}" {
			username    = artifactory_group.test-group-1.name
			scopes      = [
				"applied-permissions/groups:\"{{ .groupName1 }}\"",
				"applied-permissions/groups:${artifactory_group.test-group-2.name}",
				"applied-permissions/groups:${artifactory_group.test-group-3.name}",
			]
		}`,
		map[string]interface{}{
			"name":       name,
			"groupName1": "test group 1",
			"groupName2": "test-group-2",
			"groupName3": "test-group-3",
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accessTokenConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "username", "test group 1"),
					resource.TestCheckResourceAttr(fqrn, "scopes.#", "3"),
					resource.TestCheckTypeSetElemAttr(fqrn, "scopes.*", "applied-permissions/groups:\"test group 1\""),
					resource.TestCheckTypeSetElemAttr(fqrn, "scopes.*", "applied-permissions/groups:test-group-2"),
					resource.TestCheckTypeSetElemAttr(fqrn, "scopes.*", "applied-permissions/groups:test-group-3"),
				),
			},
			{
				ResourceName: fqrn,
				ImportState:  true,
				ExpectError:  regexp.MustCompile("resource artifactory_scoped_token doesn't support import"),
			},
		},
	})
}

func TestAccScopedToken_WithResourceScopes(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test-access-token", "artifactory_scoped_token")

	id, _, userResourceName := testutil.MkNames("test-user-", "artifactory_user")
	username := fmt.Sprintf("dummy_user%d", id)
	email := username + "@test.com"

	accessTokenConfig := util.ExecuteTemplate(
		"TestAccScopedToken",
		`resource "artifactory_user" "{{ .user_resource_name }}" {
			name              = "{{ .username }}"
			email             = "{{ .email }}"
			admin             = true
			disable_ui_access = false
			groups            = ["readers"]
			password          = "Passw0rd!"
		}

		resource "artifactory_scoped_token" "{{ .name }}" {
			username = artifactory_user.{{ .user_resource_name }}.name
			scopes   = [
				"artifact:generic-local:r",
				"artifact:generic-local:w",
				"artifact:generic-local:d",
				"artifact:generic-local:a",
				"artifact:generic-local:m",
				"artifact:generic-local:x",
				"artifact:generic-local:s",
			]
		}`,
		map[string]interface{}{
			"name":               name,
			"user_resource_name": userResourceName,
			"username":           username,
			"email":              email,
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accessTokenConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "scopes.#", "7"),
					resource.TestCheckTypeSetElemAttr(fqrn, "scopes.*", "artifact:generic-local:r"),
					resource.TestCheckTypeSetElemAttr(fqrn, "scopes.*", "artifact:generic-local:w"),
					resource.TestCheckTypeSetElemAttr(fqrn, "scopes.*", "artifact:generic-local:d"),
					resource.TestCheckTypeSetElemAttr(fqrn, "scopes.*", "artifact:generic-local:a"),
					resource.TestCheckTypeSetElemAttr(fqrn, "scopes.*", "artifact:generic-local:m"),
					resource.TestCheckTypeSetElemAttr(fqrn, "scopes.*", "artifact:generic-local:x"),
					resource.TestCheckTypeSetElemAttr(fqrn, "scopes.*", "artifact:generic-local:s"),
				),
			},
		},
	})
}

func TestAccScopedToken_WithInvalidResourceScopes(t *testing.T) {
	_, _, name := testutil.MkNames("test-access-token", "artifactory_scoped_token")

	id, _, userResourceName := testutil.MkNames("test-user-", "artifactory_user")
	username := fmt.Sprintf("dummy_user%d", id)
	email := username + "@test.com"

	accessTokenConfig := util.ExecuteTemplate(
		"TestAccScopedToken",
		`resource "artifactory_user" "{{ .user_resource_name }}" {
			name              = "{{ .username }}"
			email             = "{{ .email }}"
			admin             = true
			disable_ui_access = false
			groups            = ["readers"]
			password          = "Passw0rd!"
		}

		resource "artifactory_scoped_token" "{{ .name }}" {
			username = artifactory_user.{{ .user_resource_name }}.name
			scopes   = [
				"artifact:generic-local:q",
			]
		}`,
		map[string]interface{}{
			"name":               name,
			"user_resource_name": userResourceName,
			"username":           username,
			"email":              email,
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      accessTokenConfig,
				ExpectError: regexp.MustCompile(`.*'<resource-type>:<target>\[\/<sub-resource>]:<actions>'.*`),
			},
		},
	})
}

func TestAccScopedToken_WithInvalidSystemScopes(t *testing.T) {
	_, _, name := testutil.MkNames("test-access-token", "artifactory_scoped_token")

	id, _, userResourceName := testutil.MkNames("test-user-", "artifactory_user")
	username := fmt.Sprintf("dummy_user%d", id)
	email := username + "@test.com"

	accessTokenConfig := util.ExecuteTemplate(
		"TestAccScopedToken",
		`resource "artifactory_user" "{{ .user_resource_name }}" {
			name              = "{{ .username }}"
			email             = "{{ .email }}"
			admin             = true
			disable_ui_access = false
			groups            = ["readers"]
			password          = "Passw0rd!"
		}

		resource "artifactory_scoped_token" "{{ .name }}" {
			username = artifactory_user.{{ .user_resource_name }}.name
			scopes   = [
				"system:invalid:r",
			]
		}`,
		map[string]interface{}{
			"name":               name,
			"user_resource_name": userResourceName,
			"username":           username,
			"email":              email,
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      accessTokenConfig,
				ExpectError: regexp.MustCompile(`.*'system:\(metrics|livelogs|identities|permissions\):<actions>'.*`),
			},
		},
	})
}

func TestAccScopedToken_WithRoleScope(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test-access-token", "artifactory_scoped_token")
	_, _, projectName := testutil.MkNames("test-project", "project")
	_, _, projectUserName := testutil.MkNames("test-projecuser", "project_user")
	_, _, username := testutil.MkNames("test-user", "artifactory_managed_user")

	email := username + "@tempurl.org"

	accessTokenConfig := util.ExecuteTemplate(
		"TestAccScopedToken",
		`resource "artifactory_managed_user" "{{ .username }}" {
			name              = "{{ .username }}"
		    email             = "{{ .email }}"
			admin             = true
			disable_ui_access = false
			groups            = ["readers"]
			password          = "Passw0rd!"
		}

		resource "project" "{{ .projectName }}" {
			key = "{{ .projectName }}"
			display_name = "{{ .projectName }}"
			admin_privileges {
				manage_members = true
				manage_resources = true
				index_resources = true
			}
		}

		resource "project_user" "{{ .projectUserName }}" {
			name = artifactory_managed_user.{{ .username }}.name
			project_key = project.{{ .projectName }}.key
			roles = ["Developer"]
		}

		resource "artifactory_scoped_token" "{{ .name }}" {
			username = artifactory_managed_user.{{ .username }}.name
			scopes   = [
				"applied-permissions/roles:${project.{{ .projectName }}.key}:Developer",
			]
		}`,
		map[string]interface{}{
			"name":            name,
			"username":        username,
			"email":           email,
			"projectName":     projectName,
			"projectUserName": projectUserName,
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		ExternalProviders: map[string]resource.ExternalProvider{
			"project": {
				Source: "jfrog/project",
			},
		},
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accessTokenConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "username", username),
					resource.TestCheckResourceAttr(fqrn, "scopes.#", "1"),
					resource.TestCheckTypeSetElemAttr(fqrn, "scopes.*", fmt.Sprintf("applied-permissions/roles:%s:Developer", projectName)),
				),
			},
		},
	})
}

func TestAccScopedToken_WithActionsScope(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test-access-token", "artifactory_scoped_token")
	_, _, projectName := testutil.MkNames("test-project", "project")
	_, _, projectUserName := testutil.MkNames("test-projecuser", "project_user")
	_, _, username := testutil.MkNames("test-user", "artifactory_managed_user")

	email := username + "@tempurl.org"

	accessTokenConfig := util.ExecuteTemplate(
		"TestAccScopedToken",
		`resource "artifactory_managed_user" "{{ .username }}" {
			name              = "{{ .username }}"
		    email             = "{{ .email }}"
			admin             = true
			disable_ui_access = false
			groups            = ["readers"]
			password          = "Passw0rd!"
		}

		resource "project" "{{ .projectName }}" {
			key = "{{ .projectName }}"
			display_name = "{{ .projectName }}"
			admin_privileges {
				manage_members = true
				manage_resources = true
				index_resources = true
			}
		}

		resource "project_user" "{{ .projectUserName }}" {
			name = artifactory_managed_user.{{ .username }}.name
			project_key = project.{{ .projectName }}.key
			roles = ["Developer"]
		}

		resource "artifactory_scoped_token" "{{ .name }}" {
			username = artifactory_managed_user.{{ .username }}.name
			scopes   = [
				"artifact:generic-local-1:r",
				"artifact:generic-local-2:r,w,d,a,m",
			]
		}`,
		map[string]interface{}{
			"name":            name,
			"username":        username,
			"email":           email,
			"projectName":     projectName,
			"projectUserName": projectUserName,
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		ExternalProviders: map[string]resource.ExternalProvider{
			"project": {
				Source: "jfrog/project",
			},
		},
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accessTokenConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "username", username),
					resource.TestCheckResourceAttr(fqrn, "scopes.#", "2"),
					resource.TestCheckTypeSetElemAttr(fqrn, "scopes.*", "artifact:generic-local-1:r"),
					resource.TestCheckTypeSetElemAttr(fqrn, "scopes.*", "artifact:generic-local-2:r,w,d,a,m"),
				),
			},
		},
	})
}

func TestAccScopedToken_InvalidActionsScope(t *testing.T) {
	_, _, name := testutil.MkNames("test-access-token", "artifactory_scoped_token")
	_, _, projectName := testutil.MkNames("test-project", "project")
	_, _, projectUserName := testutil.MkNames("test-projecuser", "project_user")
	_, _, username := testutil.MkNames("test-user", "artifactory_managed_user")

	email := username + "@tempurl.org"

	accessTokenConfig := util.ExecuteTemplate(
		"TestAccScopedToken",
		`resource "artifactory_managed_user" "{{ .username }}" {
			name              = "{{ .username }}"
		    email             = "{{ .email }}"
			admin             = true
			disable_ui_access = false
			groups            = ["readers"]
			password          = "Passw0rd!"
		}

		resource "project" "{{ .projectName }}" {
			key = "{{ .projectName }}"
			display_name = "{{ .projectName }}"
			admin_privileges {
				manage_members = true
				manage_resources = true
				index_resources = true
			}
		}

		resource "project_user" "{{ .projectUserName }}" {
			name = artifactory_managed_user.{{ .username }}.name
			project_key = project.{{ .projectName }}.key
			roles = ["Developer"]
		}

		resource "artifactory_scoped_token" "{{ .name }}" {
			username = artifactory_managed_user.{{ .username }}.name
			scopes   = [
				"artifact:generic-local-1:t",
			]
		}`,
		map[string]interface{}{
			"name":            name,
			"username":        username,
			"email":           email,
			"projectName":     projectName,
			"projectUserName": projectUserName,
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		ExternalProviders: map[string]resource.ExternalProvider{
			"project": {
				Source: "jfrog/project",
			},
		},
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      accessTokenConfig,
				ExpectError: regexp.MustCompile(`.*'<resource-type>:<target>\[\/<sub-resource>\]:<actions>'.*`),
			},
		},
	})
}

func TestAccScopedToken_WithInvalidScopes(t *testing.T) {
	_, _, name := testutil.MkNames("test-scoped-token", "artifactory_scoped_token")

	scopedTokenConfig := util.ExecuteTemplate(
		"TestAccScopedToken",
		`resource "artifactory_scoped_token" "{{ .name }}" {
			scopes      = ["invalid-scope"]
		}`,
		map[string]interface{}{
			"name": name,
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      scopedTokenConfig,
				ExpectError: regexp.MustCompile(`.*'applied-permissions\/groups:<group-name>\[,<group-name>\.\.\.]'.*`),
			},
		},
	})
}

func TestAccScopedToken_WithTooLongScopes(t *testing.T) {
	_, _, name := testutil.MkNames("test-scoped-token", "artifactory_scoped_token")

	scopedTokenConfig := util.ExecuteTemplate(
		"TestAccScopedToken",
		`resource "artifactory_local_generic_repository" "generic-local-1" {
			key = "generic-local-1"
		}

		resource "artifactory_local_generic_repository" "generic-local-2" {
			key = "generic-local-2"
		}

		resource "artifactory_local_generic_repository" "generic-local-3" {
			key = "generic-local-3"
		}

		resource "artifactory_local_generic_repository" "generic-local-4" {
			key = "generic-local-4"
		}

		resource "artifactory_scoped_token" "{{ .name }}" {
			scopes      = [
				"applied-permissions/admin",
				"applied-permissions/user",
				"system:metrics:r",
				"system:livelogs:r",
				"artifact:generic-local-1:r",
				"artifact:generic-local-1:w",
				"artifact:generic-local-1:d",
				"artifact:generic-local-1:a",
				"artifact:generic-local-1:m",
				"artifact:generic-local-2:r",
				"artifact:generic-local-2:w",
				"artifact:generic-local-2:d",
				"artifact:generic-local-2:a",
				"artifact:generic-local-2:m",
				"artifact:generic-local-3:r",
				"artifact:generic-local-3:w",
				"artifact:generic-local-3:d",
				"artifact:generic-local-3:a",
				"artifact:generic-local-3:m",
				"artifact:generic-local-4:r",
				"artifact:generic-local-4:w",
				"artifact:generic-local-4:d",
				"artifact:generic-local-4:a",
				"artifact:generic-local-4:m",
			]

			depends_on = [
				artifactory_local_generic_repository.generic-local-1,
				artifactory_local_generic_repository.generic-local-2,
				artifactory_local_generic_repository.generic-local-3,
				artifactory_local_generic_repository.generic-local-4,
			]
		}`,
		map[string]interface{}{
			"name": name,
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      scopedTokenConfig,
				ExpectError: regexp.MustCompile(".*Scopes length exceeds 500 characters.*"),
			},
		},
	})
}

func TestAccScopedToken_WithAudience(t *testing.T) {

	for _, prefix := range []string{"jfrt", "jfxr", "jfpip", "jfds", "jfmc", "jfac", "jfevt", "jfmd", "jfcon", "*"} {
		t.Run(prefix, func(t *testing.T) {
			resource.Test(mkAudienceTestCase(prefix, t))
		})
	}
}

func mkAudienceTestCase(prefix string, t *testing.T) (*testing.T, resource.TestCase) {
	_, fqrn, name := testutil.MkNames("test-access-token", "artifactory_scoped_token")

	accessTokenConfig := util.ExecuteTemplate(
		"TestAccScopedToken",
		`resource "artifactory_scoped_token" "{{ .name }}" {
			audiences = ["{{ .prefix }}@*"]
		}`,
		map[string]interface{}{
			"name":   name,
			"prefix": prefix,
		},
	)

	return t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accessTokenConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "audiences.#", "1"),
					resource.TestCheckTypeSetElemAttr(fqrn, "audiences.*", fmt.Sprintf("%s@*", prefix)),
				),
			},
			{
				ResourceName: fqrn,
				ImportState:  true,
				ExpectError:  regexp.MustCompile("resource artifactory_scoped_token doesn't support import"),
			},
		},
	}
}

func TestAccScopedToken_WithInvalidAudiences(t *testing.T) {
	_, _, name := testutil.MkNames("test-scoped-token", "artifactory_scoped_token")

	scopedTokenConfig := util.ExecuteTemplate(
		"TestAccScopedToken",
		`resource "artifactory_scoped_token" "{{ .name }}" {
			audiences = ["foo@*"]
		}`,
		map[string]interface{}{
			"name": name,
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      scopedTokenConfig,
				ExpectError: regexp.MustCompile(`.*must either begin with jfrt, jfxr, jfpip,.*`),
			},
		},
	})
}

func TestAccScopedToken_WithTooLongAudiences(t *testing.T) {
	_, _, name := testutil.MkNames("test-scoped-token", "artifactory_scoped_token")

	var audiences []string
	for i := 0; i < 100; i++ {
		audiences = append(audiences, fmt.Sprintf("jfrt@%d", i))
	}

	scopedTokenConfig := util.ExecuteTemplate(
		"TestAccScopedToken",
		`resource "artifactory_scoped_token" "{{ .name }}" {
			audiences    = [
				{{range .audiences}}"{{.}}",{{end}}
			]
		}`,
		map[string]interface{}{
			"name":      name,
			"audiences": audiences,
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      scopedTokenConfig,
				ExpectError: regexp.MustCompile(".*Audiences length exceeds 255 characters.*"),
			},
		},
	})
}

func TestAccScopedToken_WithExpiresInLessThanPersistencyThreshold(t *testing.T) {
	_, _, name := testutil.MkNames("test-access-token", "artifactory_scoped_token")

	id, _, userResourceName := testutil.MkNames("test-user-", "artifactory_user")
	username := fmt.Sprintf("dummy_user%d", id)
	email := username + "@test.com"

	accessTokenConfig := util.ExecuteTemplate(
		"TestAccScopedToken",
		`resource "artifactory_user" "{{ .user_resource_name }}" {
			name              = "{{ .username }}"
			email             = "{{ .email }}"
			admin             = true
			disable_ui_access = false
			groups            = ["readers"]
			password          = "Passw0rd!"
		}

		resource "artifactory_scoped_token" "{{ .name }}" {
			username    = artifactory_user.{{ .user_resource_name }}.name
			description = "test description"
			expires_in  = {{ .expires_in }}
		}`,
		map[string]interface{}{
			"name":               name,
			"expires_in":         600, // any value > 0 and less than default persistency threshold (10800) will result in token not being saved.
			"user_resource_name": userResourceName,
			"username":           username,
			"email":              email,
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             accessTokenConfig,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccScopedToken_WithExpiresInSetToZeroForNonExpiringToken(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test-access-token", "artifactory_scoped_token")

	id, _, userResourceName := testutil.MkNames("test-user-", "artifactory_user")
	username := fmt.Sprintf("dummy_user%d", id)
	email := username + "@test.com"

	accessTokenConfig := util.ExecuteTemplate(
		"TestAccScopedToken",
		`resource "artifactory_user" "{{ .user_resource_name }}" {
			name              = "{{ .username }}"
			email             = "{{ .email }}"
			admin             = true
			disable_ui_access = false
			groups            = ["readers"]
			password          = "Passw0rd!"
		}

		resource "artifactory_scoped_token" "{{ .name }}" {
			username    = artifactory_user.{{ .user_resource_name }}.name
			description = "test description"
			expires_in  = 0
		}`,
		map[string]interface{}{
			"name":               name,
			"user_resource_name": userResourceName,
			"username":           username,
			"email":              email,
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accessTokenConfig,
				Check:  resource.TestCheckResourceAttr(fqrn, "expires_in", "0"),
			},
		},
	})
}

func checkAccessToken(id string, request *resty.Request) (*resty.Response, error) {
	return request.SetPathParam("id", id).Get("access/api/v1/tokens/{id}")
}
