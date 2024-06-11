---
subcategory: "Local Repositories"
---

# Artifactory Local Docker V2 Repository Data Source

Retrieves a local Docker (V2) repository resource

## Example Usage

```hcl
data "artifactory_local_docker_v2_repository" "artifactory_local_test_docker_v2_repository" {
  key = "artifactory_local_test_docker_v2_repository"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [common list of attributes for the local repositories](local.md):

* `block_pushing_schema1` - When set, Artifactory will block the pushing of Docker images with manifest v2
  schema 1 to this repository.
* `tag_retention` - If greater than 1, overwritten tags will be saved by their digest, up to the set up
  number. This only applies to manifest V2.
* `max_unique_tags` - The maximum number of unique tags of a single Docker image to store in this repository.
  Once the number tags for an image exceeds this setting, older tags are removed. A value of 0 (default) indicates there
  is no limit. This only applies to manifest v2.
* `api_version` "The Docker API version to use. This cannot be set"
