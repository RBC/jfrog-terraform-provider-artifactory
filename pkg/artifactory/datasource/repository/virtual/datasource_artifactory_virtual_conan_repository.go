package virtual

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/virtual"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DatasourceArtifactoryVirtualConanRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(virtual.Rclass, resource_repository.ConanPackageType)
		if err != nil {
			return nil, err
		}

		return &virtual.ConanRepoParams{
			RepositoryBaseParamsWithRetrievalCachePeriodSecs: virtual.RepositoryBaseParamsWithRetrievalCachePeriodSecs{
				RepositoryBaseParams: virtual.RepositoryBaseParams{
					Rclass:        virtual.Rclass,
					PackageType:   resource_repository.ConanPackageType,
					RepoLayoutRef: repoLayout,
				},
			},
			ConanBaseParams: resource_repository.ConanBaseParams{
				EnableConanSupport: true,
			},
		}, nil
	}

	conanSchema := virtual.ConanSchemas[virtual.CurrentSchemaVersion]

	return &schema.Resource{
		Schema:      conanSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(conanSchema), constructor),
		Description: fmt.Sprintf("Provides a data source for a virtual %s repository", resource_repository.ConanPackageType),
	}
}
