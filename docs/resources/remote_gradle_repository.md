---
subcategory: "Remote Repositories"
---
# Artifactory Remote Gradle Repository Resource

Creates a remote Gradle repository resource.
Official documentation can be found [here](https://www.jfrog.com/confluence/display/JFROG/Remote+Repositories#RemoteRepositories-Maven,Gradle,IvyandSBTRepositories).

## Example Usage

```hcl
resource "artifactory_remote_gradle_repository" "gradle-remote" {
  key                             = "gradle-remote-foo"
  url                             = "https://repo1.maven.org/maven2/"
  fetch_jars_eagerly              = true
  fetch_sources_eagerly           = false
  suppress_pom_consistency_checks = true
  reject_invalid_jars             = true
  max_unique_snapshots            = 10
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON).
The following arguments are supported, along with the [common list of arguments for the remote repositories](remote.md):

* `key` - (Required) A mandatory identifier for the repository that must be unique. It cannot begin with a number or
  contain spaces or special characters.
* `description` - (Optional)
* `notes` - (Optional)
* `url` - (Required) The remote repo URL.
* `fetch_jars_eagerly` - (Optional, Default: `false`) When set, if a POM is requested, Artifactory attempts to fetch the corresponding jar in the background. This will accelerate first access time to the jar when it is subsequently requested. 
* `fetch_sources_eagerly` - (Optional, Default: `false`) When set, if a binaries jar is requested, Artifactory attempts to fetch the corresponding source jar in the background. This will accelerate first access time to the source jar when it is subsequently requested.
* `handle_releases` - (Optional, Default: `true`) If set, Artifactory allows you to deploy release artifacts into this repository.
* `handle_snapshots` - (Optional, Default: `true`) If set, Artifactory allows you to deploy snapshot artifacts into this repository.
* `suppress_pom_consistency_checks` - (Optional, Default: `true`) - By default, the system keeps your repositories healthy by refusing POMs with incorrect coordinates (path). If the groupId:artifactId:version information inside the POM does not match the deployed path, Artifactory rejects the deployment with a "409 Conflict" error. You can disable this behavior by setting this attribute to `true`.
* `reject_invalid_jars` - (Optional, Default: `false`) Reject the caching of jar files that are found to be invalid. For example, pseudo jars retrieved behind a "captive portal".
* `remote_repo_checksum_policy_type` - (Optional, Default: `generate-if-absent`) Checking the Checksum effectively verifies the integrity of a deployed resource. The Checksum Policy determines how the system behaves when a client checksum for a remote resource is missing or conflicts with the locally calculated checksum. Available policies are `generate-if-absent`, `fail`, `ignore-and-generate`, and `pass-thru`.
* `max_unique_snapshots` - (Optional) The maximum number of unique snapshots of a single artifact to store. Once the number of snapshots exceeds this setting, older versions are removed. A value of 0 (default) indicates there is no limit, and unique snapshots are not cleaned up.
* `curated` - (Optional, Default: `false`) Enable repository to be protected by the Curation service.

## Import

Remote repositories can be imported using their name, e.g.
```
$ terraform import artifactory_remote_gradle_repository.gradle-remote gradle-remote
```
