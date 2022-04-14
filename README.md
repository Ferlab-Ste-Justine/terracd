# About

This is a continuous delivery tool for terraform that supports a gitops methodology.

terracd merge several sources of terraform files together (git repo as well as filesystem), before applying the final result, allowing any files containing secrets (ex: provider, backend, etc) to be separated from the version-controlled files.

terracd runs a single iteration and then exits, relying on an external scheduler like systemd, kubernetes or cron to schedule recurrence.

# Usage

## Basics

By default, terracd expects a file named **config.yml** to be present in its running directory. You can change the expected directory or name of the file by setting the **TERRACD_CONFIG_FILE** environment variable.

The file has the following top-level fields:
- **terraform_path**: Path to the terraform binary
- **timeouts**: Execution timeouts for the various stages of the terraform lifecycle
- **sources**: Array of terraform file sources to be merged together and applied on

The **timeouts** entry has the following fields (each taking the duration string format, see: https://pkg.go.dev/time#ParseDuration):
  - **terraform_init**: Execution timeout for the **terraform init** operation.
  - **terraform_plan**: Execution timeout for the **terraform_plan** operation.
  - **terraform_apply**: Execution timeout for the **terraform_apply** operation.

Note that the default behavior is not to apply any timeouts for fields that are omitted.

Each **sources** entry can take one of the following two forms:
```
- dir: "<local directory with terraform scripts>"
- repo:
    url: "<git repo ssh url>"
    ref: "<git repo branch>"
    path: "<path in git repo where scripts are>"
    auth:
      ssh_key_path: "<ssh key that has read access to the repo>"
      known_hosts_path: "<known host file containing the expect fingerprint of git server>"
    gpg_public_keys_paths: <Optional list of armored keyrings to validate signature of latest commit>
```

Example of config file:

```
terraform_path: /home/myuser/bin/terraform
timeouts:
  terraform_init: "15m"
  terraform_plan: "15m"
  terraform_apply: "1h"
sources:
  - repo:
      url: "git@github.com:mygituser/terracd-test.git"
      ref: "main"
      path: "dir1"
      auth:
        ssh_key_path: "/home/myuser/terracd/id_rsa"
        known_hosts_path: "/home/myuser/terracd/known_hosts"
      gpg_public_keys_paths:
        - /home/myuser/myarmoredkeyring.asc
  - dir: "/home/myuser/terracd-test/dir2"
```

## Resource Protection

terracd supports resource protection to circumvent a current limitation in terraform when managing prevent_destroy flags in modules: https://github.com/hashicorp/terraform/issues/18367

You can put yaml files in your terraform code that has the following naming convention:

```
<some custom prefix>.terracd-fo.yml
```

The file has the following format:

```
forbidden_operations:
  - resource_address: <Address of the resource>
    operations: [<operations to forbid on the resource: create, delete or update>]
    provider: <optionally specify for the provider this applies to>
  - <repeat for more resources>
```

For example, assume I have the following, totally useless purely illustrative, module in my terraform code:

```
file_module/
  main.tf
  variables.tf
```

**variables.tf**:

```
variable "content" {
  description = "Content of the file"
  type = string
}

variable "name" {
  description = "Name of the file"
  type = string
}
```

**main.tf**:

```
resource "local_file" "file" {
  content         = var.content
  file_permission = "0660"
  filename        = pathexpand("~/Projects/terracd-test/${var.name}")
}
```

Assume that I have the following code in my top-level terraform code:

```
module "filemon" {
    source = "./file_module"
    name = "filemon"
    content = "filemon"
}
```

And finally, assume that I have the following file named **file.terracd-fo.yml**:

```
forbidden_operations:
  - resource_address: "module.filemon.local_file.file"
    operations: ["delete"]
```

If I change my module invocation to this:

```
module "filemon" {
    source = "./file_module"
    name = "filemon"
    content = "filemon2"
}
```

I'll get the following runtime error with terracd:

```
panic: Aborting as forbidden operation is about to be performed on protected resource "module.filemon.local_file.file"
```

# Missing Functionality

The following functionality is planned, but not yet implemented:
  - Support for fluentd logger
  - Support for running a terraform plan on branches whose name follow a given pattern