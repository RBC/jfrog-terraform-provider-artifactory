package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/remote"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DataSourceArtifactoryRemoteNugetRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(remote.Rclass, resource_repository.NugetPackageType)
		if err != nil {
			return nil, err
		}

		return &remote.NugetRemoteRepo{
			RepositoryRemoteBaseParams: remote.RepositoryRemoteBaseParams{
				Rclass:        remote.Rclass,
				PackageType:   resource_repository.NugetPackageType,
				RepoLayoutRef: repoLayout,
			},
		}, nil
	}

	nugetSchema := getSchema(remote.NugetSchemas)

	return &schema.Resource{
		Schema:      nugetSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(nugetSchema), constructor),
		Description: "Provides a data source for a remote NuGet repository",
	}
}
