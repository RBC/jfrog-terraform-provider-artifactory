package security_test

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/security"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

func TestAccKeyPair_UpgradeFromSDKv2(t *testing.T) {
	providerHost := os.Getenv("TF_ACC_PROVIDER_HOST")
	if providerHost == "registry.opentofu.org" {
		t.Skipf("provider host is registry.opentofu.org. Previous version of Artifactory provider is unknown to OpenTofu.")
	}

	id, fqrn, name := testutil.MkNames("test", "artifactory_keypair")
	template := `
	resource "artifactory_keypair" "{{ .name }}" {
		pair_name  = "{{ .name }}"
		pair_type = "RSA"
		alias = "test-alias-{{ .id }}"
		passphrase = "{{ .passphrase }}"
		private_key = <<EOF
{{ .private_key }}
EOF
		public_key = <<EOF
{{ .public_key }}
EOF
	}`

	keyPairConfig := util.ExecuteTemplate(
		fqrn,
		template,
		map[string]string{
			"id":          fmt.Sprint(id),
			"name":        name,
			"passphrase":  "password",
			"private_key": os.Getenv("JFROG_TEST_RSA_PRIVATE_KEY"),
			"public_key":  os.Getenv("JFROG_TEST_RSA_PUBLIC_KEY"),
		},
	)

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"artifactory": {
						VersionConstraint: "9.6.0",
						Source:            "jfrog/artifactory",
					},
				},
				Config: keyPairConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "pair_name", name),
					resource.TestCheckResourceAttr(fqrn, "public_key", fmt.Sprintf("%s\n", os.Getenv("JFROG_TEST_RSA_PUBLIC_KEY"))),
					resource.TestCheckResourceAttr(fqrn, "private_key", fmt.Sprintf("%s\n", os.Getenv("JFROG_TEST_RSA_PRIVATE_KEY"))),
					resource.TestCheckResourceAttr(fqrn, "alias", fmt.Sprintf("test-alias-%d", id)),
					resource.TestCheckResourceAttr(fqrn, "pair_type", "RSA"),
					resource.TestCheckResourceAttr(fqrn, "passphrase", "password"),
					resource.TestCheckResourceAttr(fqrn, "unavailable", "false"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
				Config:                   keyPairConfig,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccKeyPair_FailPrivateCertCheck(t *testing.T) {
	id, fqrn, name := testutil.MkNames("test", "artifactory_keypair")
	keyBasic := fmt.Sprintf(`
		resource "artifactory_keypair" "%s" {
			pair_name  = "%s"
			pair_type = "RSA"
			alias = "test-alias-%d"
			private_key = "not a private key"
			public_key = <<EOF
%s
EOF
		}
	`, name, name, id, os.Getenv("JFROG_TEST_RSA_PUBLIC_KEY"))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckKeyPairDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config:      keyBasic,
				ExpectError: regexp.MustCompile(".*unable to decode private key pem format.*"),
			},
		},
	})
}

func TestAccKeyPair_FailPubCertCheck(t *testing.T) {
	id, fqrn, name := testutil.MkNames("test", "artifactory_keypair")
	keyBasic := fmt.Sprintf(`
		resource "artifactory_keypair" "%s" {
			pair_name  = "%s"
			pair_type = "RSA"
			alias = "test-alias-%d"
			private_key = <<EOF
%s
EOF
			public_key = "not a key"
		}
	`, name, name, id, os.Getenv("JFROG_TEST_RSA_PRIVATE_KEY"))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckKeyPairDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config:      keyBasic,
				ExpectError: regexp.MustCompile(".*RSA public key not in pem format.*"),
			},
		},
	})
}

func TestAccKeyPair_RSA(t *testing.T) {
	id, fqrn, name := testutil.MkNames("test", "artifactory_keypair")
	template := `
	resource "artifactory_keypair" "{{ .name }}" {
		pair_name  = "{{ .name }}"
		pair_type = "RSA"
		alias = "test-alias-{{ .id }}"
		passphrase = "{{ .passphrase }}"
		private_key = <<EOF
{{ .private_key }}
EOF
		public_key = <<EOF
{{ .public_key }}
EOF
	}`

	keyBasic := util.ExecuteTemplate(
		fqrn,
		template,
		map[string]string{
			"id":          fmt.Sprint(id),
			"name":        name,
			"passphrase":  "password",
			"private_key": os.Getenv("JFROG_TEST_RSA_PRIVATE_KEY"),
			"public_key":  os.Getenv("JFROG_TEST_RSA_PUBLIC_KEY"),
		},
	)

	keyUpdatedPassphrase := util.ExecuteTemplate(
		fqrn,
		template,
		map[string]string{
			"id":          fmt.Sprint(id),
			"name":        name,
			"passphrase":  "new-password",
			"private_key": os.Getenv("JFROG_TEST_RSA_PRIVATE_KEY"),
			"public_key":  os.Getenv("JFROG_TEST_RSA_PUBLIC_KEY"),
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckKeyPairDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: keyBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "pair_name", name),
					resource.TestCheckResourceAttr(fqrn, "public_key", fmt.Sprintf("%s\n", os.Getenv("JFROG_TEST_RSA_PUBLIC_KEY"))),
					resource.TestCheckResourceAttr(fqrn, "private_key", fmt.Sprintf("%s\n", os.Getenv("JFROG_TEST_RSA_PRIVATE_KEY"))),
					resource.TestCheckResourceAttr(fqrn, "alias", fmt.Sprintf("test-alias-%d", id)),
					resource.TestCheckResourceAttr(fqrn, "pair_type", "RSA"),
					resource.TestCheckResourceAttr(fqrn, "passphrase", "password"),
				),
			},
			{
				Config: keyUpdatedPassphrase,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "pair_name", name),
					resource.TestCheckResourceAttr(fqrn, "public_key", fmt.Sprintf("%s\n", os.Getenv("JFROG_TEST_RSA_PUBLIC_KEY"))),
					resource.TestCheckResourceAttr(fqrn, "private_key", fmt.Sprintf("%s\n", os.Getenv("JFROG_TEST_RSA_PRIVATE_KEY"))),
					resource.TestCheckResourceAttr(fqrn, "alias", fmt.Sprintf("test-alias-%d", id)),
					resource.TestCheckResourceAttr(fqrn, "pair_type", "RSA"),
					resource.TestCheckResourceAttr(fqrn, "passphrase", "new-password"),
				),
			},
			{
				ResourceName:                         fqrn,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateId:                        name,
				ImportStateVerifyIdentifierAttribute: "pair_name",
				ImportStateVerifyIgnore:              []string{"passphrase", "private_key"},
			},
		},
	})
}

func TestAccKeyPair_GPG(t *testing.T) {
	id, fqrn, name := testutil.MkNames("test", "artifactory_keypair")
	keyBasic := fmt.Sprintf(`
		resource "artifactory_keypair" "%s" {
			pair_name  = "%s"
			pair_type = "GPG"
			alias = "test-alias-%d"
			passphrase = "password"
			private_key = <<EOF
-----BEGIN PGP PRIVATE KEY BLOCK-----
Version: Keybase OpenPGP v1.0.0
Comment: https://keybase.io/crypto

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
zoQkekoL4H4GU8QB6GY=
=JQxa
-----END PGP PRIVATE KEY BLOCK-----
EOF
			public_key = <<EOF
-----BEGIN PGP PUBLIC KEY BLOCK-----
Version: Keybase OpenPGP v1.0.0
Comment: https://keybase.io/crypto

xo0EYGrVNAEEAPD3YDt0qP8kSV8bnmqVP5XDPoN40gEpUGtDLjAn6d+cRMeNGaru
6H0bdgwQpND8Gz9Qx2pCNSxlWDZpY1fCvRQ174iGjvO/3527f148cgKNZtwLsKrZ
laW8z3tB2LuCM2e97ijX+lzRf7YJUXU3pOfoCFWpOPoRg1CHV0NyHl0VABEBAAHN
FmFsYW4gPGFsYW5uQGpmcm9nLmNvbT7CrQQTAQoAFwUCYGrVNAIbLwMLCQcDFQoI
Ah4BAheAAAoJENzR2QJlA6glZmsD/iqhnNFy1Elj3hGL0HaEzeb+KDpcSL/L5a/8
WIGCQFeLcEn9lC+68b/eERKGIoXJ7z8HfPDFNRTKvomKIdAqFiAeDAUUD0B82rsx
xDf8USnTwJlnd0bPe9nxgXYcrwioEYbPVYGl3jima/KQrbW8XlKyiypy4Nd66Wcn
TuM6PwRFzo0EYGrVNAEEANVNINyfCQ+y1haaaAJ0uCgx3dW52LwcZfvOP6i798WZ
dyGA+WSUCEcrklUwZ595E2dNkNKptksftwSeQ0+EH5S1ZlEaq2YUv8fCx32F1ckh
D3eHaCKRxTPx/zbb96q4ruEGKhOBXceid3o341HbtGVKi8VjBx3XNukskQ+EOvgt
ABEBAAHCwIMEGAEKAA8FAmBq1TQFCQ8JnAACGy4AqAkQ3NHZAmUDqCWdIAQZAQoA
BgUCYGrVNAAKCRBddQ63FhKl6NmDBACqxC4lAnsCQERjs02LYAEAwVDhDf0rXxD0
H+hKDyxQZc80M7WIpXaBHmbs8ekJRnY7JHcer7sizDMdfkR3xB62jNGhc0XiW6nc
mlwvtWt3+E6AkObmWnocRy5ztTQI0gye0B3cPs2txE2fCs+WD7yLRnM3HqIAh83W
Cccvh0+uG96dBADlPbZ8g8q6bkeeT72gOi3OCN0A+Y8lUPifhrpSiI9xMpP3aomM
beZJB6fWjEzNoblQ9jUr/E54bF9jMr6L3uE4OJH9SYJ/HvqcKJC+1TFeQ9lXR7g7
MTdfxEvhMDhcsd/pYIgrzvDry+B2+jANW10R1yejT/C8QdlWIndDsEsaKs6NBGBq
1TQBBADJD6gVGTMGb/WsfnSaL5yA3AEMczhPqxD4FDC0vzqGG3XzKgxtmW8cJXls
NCK80F61daxJ72/UAmfxHbNP1qmHSiosEe+h1YZ/Zo3pN/LODzg9JrOs9A2xwjqE
nU9mS0jDz5oEQtr9K4+YKOdJvmuaN85ueBizfQ1TYRNuDtmbnQARAQABwsCDBBgB
CgAPBQJgatU0BQkPCZwAAhsuAKgJENzR2QJlA6glnSAEGQEKAAYFAmBq1TQACgkQ
ehctgYvYtmh38gP6A9lnQaLuVnTElJLy2XSDTqwWOcy/5J842S/xdQEsWUMXh4I5
mlotkZwkrdvXp8E/F3P8X7GbxhNAVZX+Xcm95V3g/kmP+Pq7PeUmoZR5LD8ppBfO
7v6XgaUhraUPAZl6lx4L5pYNCX9JBNUtQAG9xIoap4slvksdz5SN/BwSgV6qqwQA
tr4YTDXvLyoWwMFB2FjWcw4zwV+7yHwGzogKfGCQy5qVlDoQyWdkwwF1awyk5RIe
ZxwPZ2SDaiznOmZ+4LjR2NPmjnT96d9RKRtgEjkfW+a19BofrvEalS9wh/jkboea
d8rDu8wMbLAl77dq1c6dpJDgzoQkekoL4H4GU8QB6GY=
=fot9
-----END PGP PUBLIC KEY BLOCK-----
EOF
		}
	`, name, name, id)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckKeyPairDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: keyBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "pair_name", name),
					resource.TestCheckResourceAttr(fqrn, "public_key", "-----BEGIN PGP PUBLIC KEY BLOCK-----\nVersion: Keybase OpenPGP v1.0.0\nComment: https://keybase.io/crypto\n\nxo0EYGrVNAEEAPD3YDt0qP8kSV8bnmqVP5XDPoN40gEpUGtDLjAn6d+cRMeNGaru\n6H0bdgwQpND8Gz9Qx2pCNSxlWDZpY1fCvRQ174iGjvO/3527f148cgKNZtwLsKrZ\nlaW8z3tB2LuCM2e97ijX+lzRf7YJUXU3pOfoCFWpOPoRg1CHV0NyHl0VABEBAAHN\nFmFsYW4gPGFsYW5uQGpmcm9nLmNvbT7CrQQTAQoAFwUCYGrVNAIbLwMLCQcDFQoI\nAh4BAheAAAoJENzR2QJlA6glZmsD/iqhnNFy1Elj3hGL0HaEzeb+KDpcSL/L5a/8\nWIGCQFeLcEn9lC+68b/eERKGIoXJ7z8HfPDFNRTKvomKIdAqFiAeDAUUD0B82rsx\nxDf8USnTwJlnd0bPe9nxgXYcrwioEYbPVYGl3jima/KQrbW8XlKyiypy4Nd66Wcn\nTuM6PwRFzo0EYGrVNAEEANVNINyfCQ+y1haaaAJ0uCgx3dW52LwcZfvOP6i798WZ\ndyGA+WSUCEcrklUwZ595E2dNkNKptksftwSeQ0+EH5S1ZlEaq2YUv8fCx32F1ckh\nD3eHaCKRxTPx/zbb96q4ruEGKhOBXceid3o341HbtGVKi8VjBx3XNukskQ+EOvgt\nABEBAAHCwIMEGAEKAA8FAmBq1TQFCQ8JnAACGy4AqAkQ3NHZAmUDqCWdIAQZAQoA\nBgUCYGrVNAAKCRBddQ63FhKl6NmDBACqxC4lAnsCQERjs02LYAEAwVDhDf0rXxD0\nH+hKDyxQZc80M7WIpXaBHmbs8ekJRnY7JHcer7sizDMdfkR3xB62jNGhc0XiW6nc\nmlwvtWt3+E6AkObmWnocRy5ztTQI0gye0B3cPs2txE2fCs+WD7yLRnM3HqIAh83W\nCccvh0+uG96dBADlPbZ8g8q6bkeeT72gOi3OCN0A+Y8lUPifhrpSiI9xMpP3aomM\nbeZJB6fWjEzNoblQ9jUr/E54bF9jMr6L3uE4OJH9SYJ/HvqcKJC+1TFeQ9lXR7g7\nMTdfxEvhMDhcsd/pYIgrzvDry+B2+jANW10R1yejT/C8QdlWIndDsEsaKs6NBGBq\n1TQBBADJD6gVGTMGb/WsfnSaL5yA3AEMczhPqxD4FDC0vzqGG3XzKgxtmW8cJXls\nNCK80F61daxJ72/UAmfxHbNP1qmHSiosEe+h1YZ/Zo3pN/LODzg9JrOs9A2xwjqE\nnU9mS0jDz5oEQtr9K4+YKOdJvmuaN85ueBizfQ1TYRNuDtmbnQARAQABwsCDBBgB\nCgAPBQJgatU0BQkPCZwAAhsuAKgJENzR2QJlA6glnSAEGQEKAAYFAmBq1TQACgkQ\nehctgYvYtmh38gP6A9lnQaLuVnTElJLy2XSDTqwWOcy/5J842S/xdQEsWUMXh4I5\nmlotkZwkrdvXp8E/F3P8X7GbxhNAVZX+Xcm95V3g/kmP+Pq7PeUmoZR5LD8ppBfO\n7v6XgaUhraUPAZl6lx4L5pYNCX9JBNUtQAG9xIoap4slvksdz5SN/BwSgV6qqwQA\ntr4YTDXvLyoWwMFB2FjWcw4zwV+7yHwGzogKfGCQy5qVlDoQyWdkwwF1awyk5RIe\nZxwPZ2SDaiznOmZ+4LjR2NPmjnT96d9RKRtgEjkfW+a19BofrvEalS9wh/jkboea\nd8rDu8wMbLAl77dq1c6dpJDgzoQkekoL4H4GU8QB6GY=\n=fot9\n-----END PGP PUBLIC KEY BLOCK-----\n"),
					resource.TestCheckResourceAttr(fqrn, "private_key", "-----BEGIN PGP PRIVATE KEY BLOCK-----\nVersion: Keybase OpenPGP v1.0.0\nComment: https://keybase.io/crypto\n\n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\nzoQkekoL4H4GU8QB6GY=\n=JQxa\n-----END PGP PRIVATE KEY BLOCK-----\n"),
					resource.TestCheckResourceAttr(fqrn, "alias", fmt.Sprintf("test-alias-%d", id)),
					resource.TestCheckResourceAttr(fqrn, "pair_type", "GPG"),
					resource.TestCheckResourceAttr(fqrn, "passphrase", "password"),
				),
			},
			{
				ResourceName:                         fqrn,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateId:                        name,
				ImportStateVerifyIdentifierAttribute: "pair_name",
				ImportStateVerifyIgnore:              []string{"passphrase", "private_key"},
			},
		},
	})
}

func testAccCheckKeyPairDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		client := acctest.Provider.Meta().(util.ProviderMetadata).Client
		rs, ok := s.RootModule().Resources[id]
		if !ok {
			return fmt.Errorf("err: Resource id[%s] not found", id)
		}

		pairName := rs.Primary.Attributes["pair_name"]
		var keyPair security.KeyPairAPIModel

		resp, err := client.R().
			SetResult(&keyPair).
			Get(security.KeypairEndPoint + pairName)
		if err != nil && resp.StatusCode() != http.StatusNotFound {
			return err
		}

		if keyPair.PairName != "" {
			return fmt.Errorf("error: key pair %s still exists", pairName)
		}

		return nil
	}
}
