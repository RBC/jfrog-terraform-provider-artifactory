---
subcategory: "Webhook"
---
# Artifactory "User" Webhook Resource

Provides an Artifactory webhook resource. This can be used to register and manage Artifactory webhook subscription which enables you to be notified or notify other users when such events take place in Artifactory.

## Example Usage
.
```hcl
resource "artifactory_user_webhook" "user-webhook" {
  key         = "user-webhook"
  event_types = ["locked"]

  handler {
    url    = "http://tempurl.org/webhook"
    secret = "some-secret"
    proxy  = "proxy-key"

    custom_http_headers = {
      header-1 = "value-1"
      header-2 = "value-2"
    }
  }
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog Webhook API](https://www.jfrog.com/confluence/display/JFROG/Artifactory+REST+API). The following arguments are supported:

The following arguments are supported:

* `key` - (Required) The identity key of the webhook. Must be between 2 and 200 characters. Cannot contain spaces.
* `description` - (Optional) Webhook description. Max length 1000 characters.
* `enabled` - (Optional) Status of webhook. Default to `true`
* `event_types` - (Required) List of event triggers for the Webhook. Allow values: `locked`
* `handler` - (Required) At least one is required.
  * `url` - (Required) Specifies the URL that the Webhook invokes. This will be the URL that Artifactory will send an HTTP POST request to.
  * `secret` - (Optional) Secret authentication token that will be sent to the configured URL. The value will be sent as `x-jfrog-event-auth` header.
  * `use_secret_for_signing` - (Optional) When set to `true`, the secret will be used to sign the event payload, allowing the target to validate that the payload content has not been changed and will not be passed as part of the event. If left unset or set to `false`, the secret is passed through the `X-JFrog-Event-Auth` HTTP header.
  * `proxy` - (Optional) Proxy key from Artifactory UI (Administration -> Proxies -> Configuration).
  * `custom_http_headers` - (Optional) Custom HTTP headers you wish to use to invoke the Webhook, comprise of key/value pair.
