package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/remote"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DataSourceArtifactoryRemoteConanRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(rclass, resource_repository.ConanPackageType)()
		if err != nil {
			return nil, err
		}

		return &remote.ConanRepo{
			RepositoryRemoteBaseParams: remote.RepositoryRemoteBaseParams{
				Rclass:        rclass,
				PackageType:   resource_repository.ConanPackageType,
				RepoLayoutRef: repoLayout.(string),
			},
		}, nil
	}

	conanSchema := remote.ConanSchema(false)

	return &schema.Resource{
		Schema:      conanSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(conanSchema), constructor),
		Description: "Provides a data source for a remote Conan repository",
	}
}
