package configuration_test

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

func TestAccArchivePolicy_invalid_key(t *testing.T) {
	testCases := []struct {
		key        string
		errorRegex string
	}{
		{key: "1", errorRegex: ".*string length must be at least 3"},
		{key: "ab#", errorRegex: ".*only letters, numbers, underscore and hyphen are allowed"},
		{key: "ab1#", errorRegex: ".*only letters, numbers, underscore and hyphen are allowed"},
	}
	for _, testCase := range testCases {
		t.Run(testCase.key, func(t *testing.T) {
			client := acctest.GetTestResty(t)
			version, err := util.GetArtifactoryVersion(client)
			if err != nil {
				t.Fatal(err)
			}
			valid, err := util.CheckVersion(version, "7.90.1")
			if err != nil {
				t.Fatal(err)
			}
			if !valid {
				t.Skipf("Artifactory version %s is earlier than 7.90.1", version)
			}

			_, _, policyName := testutil.MkNames("test-archive-policy", "artifactory_archive_policy")

			temp := `
			resource "artifactory_archive_policy" "{{ .policyName }}" {
				key = "{{ .policyKey }}"
				description = "Test policy"
				cron_expression = "0 0 2 ? * MON-SAT *"
				duration_in_minutes = 60
				enabled = true
				skip_trashcan = false
				
				search_criteria = {
					repos = ["**"]
					package_types = ["docker"]
					include_all_projects = true
					included_packages = ["**"]
					excluded_packages = ["com/jfrog/latest"]
					created_before_in_months = 1
					last_downloaded_before_in_months = 6
				}
			}`

			config := util.ExecuteTemplate(
				policyName,
				temp,
				map[string]string{
					"policyName": policyName,
					"policyKey":  testCase.key,
				},
			)

			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { acctest.PreCheck(t) },
				ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
				Steps: []resource.TestStep{
					{
						Config:      config,
						ExpectError: regexp.MustCompile(testCase.errorRegex),
					},
				},
			})
		})
	}
}

func TestAccArchivePolicy_invalid_conditions(t *testing.T) {
	client := acctest.GetTestResty(t)
	version, err := util.GetArtifactoryVersion(client)
	if err != nil {
		t.Fatal(err)
	}
	valid, err := util.CheckVersion(version, "7.101.0")
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Skipf("Artifactory version %s is earlier than 7.101.0", version)
	}

	_, _, policyName := testutil.MkNames("test-package-cleanup-policy", "artifactory_archive_policy")

	temp := `
	resource "artifactory_archive_policy" "{{ .policyName }}" {
		key = "{{ .policyName }}"
		description = "Test policy"
		cron_expression = "0 0 2 ? * MON-SAT *"
		duration_in_minutes = 60
		enabled = true
		skip_trashcan = false
		
		search_criteria = {
			repos = ["**"]
			package_types = ["docker"]
			include_all_projects = true
			included_packages = ["**"]
			excluded_packages = ["com/jfrog/latest"]
			created_before_in_months = 0
			last_downloaded_before_in_months = 0
		}
	}`

	config := util.ExecuteTemplate(
		policyName,
		temp,
		map[string]string{
			"policyName": policyName,
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(".*Both created_before_in_months and last_downloaded_before_in_months cannot be\n.*zero at the same time.*"),
			},
		},
	})
}

func TestAccArchivePolicy_full(t *testing.T) {
	archivePolicyEnabled := os.Getenv("JFROG_ARCHIVE_POLICY_ENABLED")
	if strings.ToLower(archivePolicyEnabled) != "true" {
		t.Skipf("JFROG_ARCHIVE_POLICY_ENABLED env var is not set to 'true'")
	}

	client := acctest.GetTestResty(t)
	version, err := util.GetArtifactoryVersion(client)
	if err != nil {
		t.Fatal(err)
	}
	valid, err := util.CheckVersion(version, "7.101.0")
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Skipf("Artifactory version %s is earlier than 7.101.0", version)
	}

	_, fqrn, policyName := testutil.MkNames("test-archive-policy", "artifactory_archive_policy")
	_, _, repoName := testutil.MkNames("test-docker-local", "artifactory_local_docker_v2_repository")
	_, _, projectKey := testutil.MkNames("testproj", "project")

	temp := `
	resource "artifactory_local_docker_v2_repository" "{{ .repoName }}" {
		key             = "{{ .repoName }}"
		tag_retention   = 3
		max_unique_tags = 5
	}

	resource "project" "{{ .projectKey }}" {
		key = "{{ .projectKey }}"
		display_name = "Test Project"
		description  = "Test Project"
		admin_privileges {
			manage_members   = true
			manage_resources = true
			index_resources  = true
		}
		max_storage_in_gibibytes   = 10
		block_deployments_on_limit = false
		email_notification         = true
	}

	resource "artifactory_archive_policy" "{{ .policyName }}" {
		key = "{{ .policyName }}"
		description = "Test policy"
		cron_expression = "0 0 2 ? * MON-SAT *"
		duration_in_minutes = 60
		enabled = true
		skip_trashcan = false
		
		search_criteria = {
			package_types = ["docker"]
			repos = [artifactory_local_docker_v2_repository.{{ .repoName }}.key]
			included_projects = [project.{{ .projectKey }}.key]
			included_packages = ["**"]
			excluded_packages = ["com/jfrog/latest"]
			created_before_in_months = 0
			last_downloaded_before_in_months = 6
		}
	}`

	updatedTemp := `
	resource "artifactory_local_docker_v2_repository" "{{ .repoName }}" {
		key             = "{{ .repoName }}"
		tag_retention   = 3
		max_unique_tags = 5
	}

	resource "artifactory_archive_policy" "{{ .policyName }}" {
		key = "{{ .policyName }}"
		description = "Test policy"
		cron_expression = "0 0 2 ? * MON-SAT *"
		duration_in_minutes = 120
		enabled = false
		skip_trashcan = false
		
		search_criteria = {
			package_types = ["cargo", "cocoapods", "conan", "debian", "docker", "gems", "generic", "go", "gradle", "helm", "helmoci", "huggingfaceml", "maven", "npm", "nuget", "oci", "pypi", "terraform", "yum"]
			repos = ["**"]
			included_packages = ["**"]
			excluded_packages = ["com/jfrog/latest"]
			include_all_projects = true
			created_before_in_months = 12
			last_downloaded_before_in_months = 0
		}
	}`

	config := util.ExecuteTemplate(
		policyName,
		temp,
		map[string]string{
			"policyName": policyName,
			"repoName":   repoName,
			"projectKey": projectKey,
		},
	)

	updatedConfig := util.ExecuteTemplate(
		policyName,
		updatedTemp,
		map[string]string{
			"policyName": policyName,
			"repoName":   repoName,
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"project": {
				Source: "jfrog/project",
			},
		},
		CheckDestroy: testAccArchivePolicyDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", policyName),
					resource.TestCheckResourceAttr(fqrn, "description", "Test policy"),
					resource.TestCheckResourceAttr(fqrn, "cron_expression", "0 0 2 ? * MON-SAT *"),
					resource.TestCheckResourceAttr(fqrn, "duration_in_minutes", "60"),
					resource.TestCheckResourceAttr(fqrn, "enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "skip_trashcan", "false"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.package_types.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.package_types.0", "docker"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.repos.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.repos.0", repoName),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.included_packages.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.included_packages.0", "**"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.excluded_packages.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.excluded_packages.0", "com/jfrog/latest"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.created_before_in_months", "0"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.last_downloaded_before_in_months", "6"),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", policyName),
					resource.TestCheckResourceAttr(fqrn, "description", "Test policy"),
					resource.TestCheckResourceAttr(fqrn, "cron_expression", "0 0 2 ? * MON-SAT *"),
					resource.TestCheckResourceAttr(fqrn, "duration_in_minutes", "120"),
					resource.TestCheckResourceAttr(fqrn, "enabled", "false"),
					resource.TestCheckResourceAttr(fqrn, "skip_trashcan", "false"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.package_types.#", "19"),
					resource.TestCheckTypeSetElemAttr(fqrn, "search_criteria.package_types.*", "cargo"),
					resource.TestCheckTypeSetElemAttr(fqrn, "search_criteria.package_types.*", "cocoapods"),
					resource.TestCheckTypeSetElemAttr(fqrn, "search_criteria.package_types.*", "conan"),
					resource.TestCheckTypeSetElemAttr(fqrn, "search_criteria.package_types.*", "debian"),
					resource.TestCheckTypeSetElemAttr(fqrn, "search_criteria.package_types.*", "docker"),
					resource.TestCheckTypeSetElemAttr(fqrn, "search_criteria.package_types.*", "gems"),
					resource.TestCheckTypeSetElemAttr(fqrn, "search_criteria.package_types.*", "generic"),
					resource.TestCheckTypeSetElemAttr(fqrn, "search_criteria.package_types.*", "go"),
					resource.TestCheckTypeSetElemAttr(fqrn, "search_criteria.package_types.*", "gradle"),
					resource.TestCheckTypeSetElemAttr(fqrn, "search_criteria.package_types.*", "helm"),
					resource.TestCheckTypeSetElemAttr(fqrn, "search_criteria.package_types.*", "helmoci"),
					resource.TestCheckTypeSetElemAttr(fqrn, "search_criteria.package_types.*", "huggingfaceml"),
					resource.TestCheckTypeSetElemAttr(fqrn, "search_criteria.package_types.*", "maven"),
					resource.TestCheckTypeSetElemAttr(fqrn, "search_criteria.package_types.*", "npm"),
					resource.TestCheckTypeSetElemAttr(fqrn, "search_criteria.package_types.*", "nuget"),
					resource.TestCheckTypeSetElemAttr(fqrn, "search_criteria.package_types.*", "oci"),
					resource.TestCheckTypeSetElemAttr(fqrn, "search_criteria.package_types.*", "pypi"),
					resource.TestCheckTypeSetElemAttr(fqrn, "search_criteria.package_types.*", "terraform"),
					resource.TestCheckTypeSetElemAttr(fqrn, "search_criteria.package_types.*", "yum"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.repos.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.repos.0", "**"),
					resource.TestCheckNoResourceAttr(fqrn, "search_criteria.include_projects"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.include_all_projects", "true"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.included_packages.#", "1"),
					resource.TestCheckTypeSetElemAttr(fqrn, "search_criteria.included_packages.*", "**"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.excluded_packages.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.excluded_packages.0", "com/jfrog/latest"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.created_before_in_months", "12"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.last_downloaded_before_in_months", "0"),
				),
			},
			{
				ResourceName:                         fqrn,
				ImportState:                          true,
				ImportStateId:                        policyName,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "key",
			},
		},
	})
}

func TestAccArchivePolicy_with_project_key(t *testing.T) {
	archivePolicyEnabled := os.Getenv("JFROG_ARCHIVE_POLICY_ENABLED")
	if strings.ToLower(archivePolicyEnabled) != "true" {
		t.Skipf("JFROG_ARCHIVE_POLICY_ENABLED env var is not set to 'true'")
	}

	client := acctest.GetTestResty(t)
	version, err := util.GetArtifactoryVersion(client)
	if err != nil {
		t.Fatal(err)
	}
	valid, err := util.CheckVersion(version, "7.101.0")
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Skipf("Artifactory version %s is earlier than 7.101.0", version)
	}

	_, fqrn, policyName := testutil.MkNames("test-package-cleanup-policy", "artifactory_archive_policy")
	_, _, repoName := testutil.MkNames("test-docker-local", "artifactory_local_docker_v2_repository")
	_, _, projectKey := testutil.MkNames("testproj", "project")

	temp := `
	resource "artifactory_local_docker_v2_repository" "{{ .repoName }}" {
		key             = "{{ .repoName }}"
		tag_retention   = 3
		max_unique_tags = 5

		lifecycle {
			ignore_changes = ["project_key"]
		}
	}

	resource "project" "{{ .projectKey }}" {
		key = "{{ .projectKey }}"
		display_name = "Test Project"
		description  = "Test Project"
		admin_privileges {
			manage_members   = true
			manage_resources = true
			index_resources  = true
		}
		max_storage_in_gibibytes   = 10
		block_deployments_on_limit = false
		email_notification         = true
	}

	resource "project_repository" "{{ .projectKey }}-{{ .repoName }}" {
		project_key = project.{{ .projectKey }}.key
		key = artifactory_local_docker_v2_repository.{{ .repoName }}.key
	}

	resource "artifactory_archive_policy" "{{ .policyName }}" {
		key = "${project.{{ .projectKey }}.key}-{{ .policyName }}"
		description = "Test policy"
		cron_expression = "0 0 2 ? * MON-SAT *"
		duration_in_minutes = 60
		enabled = true
		skip_trashcan = false
		project_key = project_repository.{{ .projectKey }}-{{ .repoName }}.project_key
		
		search_criteria = {
			package_types = ["docker"]
			repos = [project_repository.{{ .projectKey }}-{{ .repoName }}.key]
			included_packages = ["**"]
			excluded_packages = ["com/jfrog/latest"]
			included_projects = []
			created_before_in_months = 1
			last_downloaded_before_in_months = 6
		}
	}`

	updatedTemp := `
	resource "artifactory_local_docker_v2_repository" "{{ .repoName }}" {
		key             = "{{ .repoName }}"
		tag_retention   = 3
		max_unique_tags = 5

		lifecycle {
			ignore_changes = ["project_key"]
		}
	}

	resource "project" "{{ .projectKey }}" {
		key = "{{ .projectKey }}"
		display_name = "Test Project"
		description  = "Test Project"
		admin_privileges {
			manage_members   = true
			manage_resources = true
			index_resources  = true
		}
		max_storage_in_gibibytes   = 10
		block_deployments_on_limit = false
		email_notification         = true
	}

	resource "artifactory_archive_policy" "{{ .policyName }}" {
		key = "${project.{{ .projectKey }}.key}-{{ .policyName }}"
		description = "Test policy"
		cron_expression = "0 0 2 ? * MON-SAT *"
		duration_in_minutes = 120
		enabled = false
		skip_trashcan = false
		project_key = project.{{ .projectKey }}.key

		search_criteria = {
			package_types = ["docker", "maven", "gradle"]
			repos = ["**"]
			included_packages = ["**"]
			excluded_packages = ["com/jfrog/latest"]
			included_projects = []
			created_before_in_months = 12
			last_downloaded_before_in_months = 24
		}
	}`

	config := util.ExecuteTemplate(
		policyName,
		temp,
		map[string]string{
			"policyName": policyName,
			"repoName":   repoName,
			"projectKey": projectKey,
		},
	)

	updatedConfig := util.ExecuteTemplate(
		policyName,
		updatedTemp,
		map[string]string{
			"policyName": policyName,
			"repoName":   repoName,
			"projectKey": projectKey,
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"project": {
				Source: "jfrog/project",
			},
		},
		CheckDestroy: testAccArchivePolicyDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", fmt.Sprintf("%s-%s", projectKey, policyName)),
					resource.TestCheckResourceAttr(fqrn, "description", "Test policy"),
					resource.TestCheckResourceAttr(fqrn, "cron_expression", "0 0 2 ? * MON-SAT *"),
					resource.TestCheckResourceAttr(fqrn, "duration_in_minutes", "60"),
					resource.TestCheckResourceAttr(fqrn, "enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "skip_trashcan", "false"),
					resource.TestCheckResourceAttr(fqrn, "project_key", projectKey),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.package_types.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.package_types.0", "docker"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.repos.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.repos.0", repoName),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.included_packages.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.included_packages.0", "**"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.excluded_packages.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.excluded_packages.0", "com/jfrog/latest"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.created_before_in_months", "1"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.last_downloaded_before_in_months", "6"),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", fmt.Sprintf("%s-%s", projectKey, policyName)),
					resource.TestCheckResourceAttr(fqrn, "description", "Test policy"),
					resource.TestCheckResourceAttr(fqrn, "cron_expression", "0 0 2 ? * MON-SAT *"),
					resource.TestCheckResourceAttr(fqrn, "duration_in_minutes", "120"),
					resource.TestCheckResourceAttr(fqrn, "enabled", "false"),
					resource.TestCheckResourceAttr(fqrn, "skip_trashcan", "false"),
					resource.TestCheckResourceAttr(fqrn, "project_key", projectKey),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.package_types.#", "3"),
					resource.TestCheckTypeSetElemAttr(fqrn, "search_criteria.package_types.*", "docker"),
					resource.TestCheckTypeSetElemAttr(fqrn, "search_criteria.package_types.*", "maven"),
					resource.TestCheckTypeSetElemAttr(fqrn, "search_criteria.package_types.*", "gradle"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.repos.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.repos.0", "**"),
					resource.TestCheckTypeSetElemAttr(fqrn, "search_criteria.included_packages.*", "**"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.excluded_packages.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.excluded_packages.0", "com/jfrog/latest"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.created_before_in_months", "12"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.last_downloaded_before_in_months", "24"),
				),
			},
			{
				ResourceName:                         fqrn,
				ImportState:                          true,
				ImportStateId:                        fmt.Sprintf("%s-%s:%s", projectKey, policyName, projectKey),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "key",
			},
		},
	})
}

func testAccArchivePolicyDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[id]
		if !ok {
			return fmt.Errorf("error: resource id [%s] not found", id)
		}

		client := acctest.Provider.Meta().(util.ProviderMetadata).Client
		resp, err := client.R().
			SetPathParam("policyKey", rs.Primary.Attributes["key"]).
			Get("artifactory/api/archive/v2/packages/policies/{policyKey}")
		if err != nil {
			return err
		}

		if resp != nil && resp.StatusCode() == http.StatusNotFound {
			return nil
		}

		return fmt.Errorf("error: Archive Policy %s still exists", rs.Primary.ID)
	}
}
