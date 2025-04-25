# About

This is a continuous delivery tool for terraform that supports a gitops methodology.

terracd merge several sources of terraform files together (git repo as well as filesystem and generated http backend files), before running an operation on the final result, allowing any files containing secrets (ex: provider, backend, etc) to be separated from the version-controlled files.

terracd runs a single iteration and then exits, relying on an external scheduler like systemd, kubernetes or cron to schedule recurrence.

# Usage

## Basics

By default, terracd expects a file named **config.yml** to be present in its running directory. You can change the expected directory or name of the file by setting the **TERRACD_CONFIG_FILE** environment variable.

The file has the following top-level fields:
- **terraform_path**: Path to the terraform binary
- **working_directory**: Directory where terracd will assemble its workspace from the various sources. Defaults to the working directory of the process if omitted.
- **timeouts**: Execution timeouts for the various stages of the terraform lifecycle
- **random_jitter**: Golang duration format indicating a random start delay up to that duration. Useful to spread the load a little when you use a scheduler that triggers at the same time for all your jobs.
- **state_store**: Storage strategy to store a persistent terracd state between executions. Needed to support provider caching and recurrence control.
- **recurrence**: Allows more fine-grained control on when terracd re-executes beyond what schedulers normallly support. Note that it is dependant on a state store.
- **cache**: Allows the caching of terraform providers (currently only supported on the filesystem of stable runtime environments) between executions of terracd. Note that it is dependent on a state store.
- **metrics**: Specify configuration to push timestamp metric on a prometheus pushgateway. Note that since only a stateless timestamp metric is currently exported, a state store is **not** necessary to use this feature.
- **sources**: Array of terraform file sources to be merged together and applied on
- **command**: Command to execute. Can be **apply** to run **terraform apply**, **plan** to run **terraform plan**, **destroy** to run **terraform destroy**, **migrate_backend** to migrate the terraform state to another backend file or **wait** to simply assemble all the sources together and wait a given duration before exiting (useful for importing resources). Defaults to **apply** if omitted.
- **backend_migration**: Parameters specifying the backend files to rotate when migrating your backend.
- **termination_hooks**: Logic to call when the terraform command is done

The **timeouts** entry has the following fields (each taking the duration string format, see: https://pkg.go.dev/time#ParseDuration):
  - **terraform_init**: Execution timeout for the **terraform init** operation.
  - **terraform_plan**: Execution timeout for the **terraform plan** operation.
  - **terraform_apply**: Execution timeout for the **terraform apply** operation.
  - **terraform_destroy**: Execution timeout for the **terraform destroy** operation.
  - **wait**: Execution timeout for the **wait** command.

Note that the default behavior is not to apply any timeouts for fields that are omitted.

The **state_store** entry has the following fields:
- **fs**: Configuration if you want to store terracd's state under **working_directory** in a stable runtime environment (ex: an automation server, but not a kubernetes cron job). It has a single **enabled** field which should be set to **true** (boolean value) if you want to use the filesystem
- **etcd**: Configuration if you want to store terracd's state in a remote **etcd** store. It has the following fields:
  - **prefix**: Key prefix to store the state under in etcd
  - **Endpoints**: Endpoints of your etcd store
  - **connection_timeout**: Timeout for the etcd connection
  - **request_timeout**: Timeout for individual etcd requests
  - **retry_interval**: Retry interval to wait for after an etcd request failed
  - **retries**: Number of retries to perform when an etcd request fails before giving up
  - **auth**: mTLS or tls + password authentication parameters. It takes the following fields:
    - **ca_cert**: Path to a CA certificate validating the server certificates of the etcd cluster.
    - **client_cert**: Client certificate to use to authentify itself to the cluster when using mTLS.
    - **client_key**: Client private key to use to authentify itself to the cluster when using mTLS.
    - **password_auth**: Yaml file containing a **username** and **password** key to authentify against an etcd cluster using password authentication.

Each **sources** entry can take one of the following 3 forms:
```
- dir: "<local directory with terraform scripts>"
- repo:
    url: "<git repo ssh url>"
    ref: "<git repo branch>"
    path: "<path in git repo where scripts are>"
    auth:
      ssh_key_path: "<ssh key that has read access to the repo>"
      known_hosts_path: "<known host file containing the expect fingerprint of git server>"
      user: "<user to ssh as. Can often be omitted, but some git server implementations require it>"
    gpg_public_keys_paths: <Optional list of armored keyrings to validate signature of latest commit>
- backend_http:
    filename: "<File name to give the generated backend file>"
    address:
      base: "<Base state url that terraform will use to update and retrieve the state>"
      query_string: "<list of key/string array pairs representing the query string parameters. It will be url encoded by terracd>"
    update_method: "<Http method to use when updating the state via the state url>"
    lock_address:
      base: "<Base state url that terraform will use to obtain a lock on the terraform state>"
      query_string: "<list of key/string array pairs representing the query string parameters. It will be url encoded by terracd>"
    lock_method: "<Http method to use when acquiring a lock on the state using the lock url>"
    unlock_address:
      base: "<Base state url that terraform will use to release a lock on the terraform state>"
      query_string: "<list of key/string array pairs representing the query string parameters. It will be url encoded by terracd>"
    unlock_method: "<Http method to use when releasing a lock on the state using the unlock url>"
```

Note that secret parameters (username/password or client certificate) are absent from the **backend_http** source. They should be passed via environment variables when running terracd.

The **recurrence** entry takes the following fields:
- **min_interval**: Minimum interval of time between execution. If terracd finds that less than this interval of time has elapsed since the last time it ran, it will skip its execution.
- **git_triggers**: Boolean flag indicating whether a change in the git history of any of its git sources should also trigger a change, despite the minimum interval of time not having elapsed.

The **cache** entry takes the following field:
- **versions_file**: Path to a terraform provider versions file to hash in its assembled runtime directory. If the sha256 checksum value of this file changes, the cached providers will be discarded and redownloaded.

The **metrics** entry takes the following fields:
- **job_name**: Job tag to provide to the exported metric
- **pushgateway**: Parameters to communicate with the prometheus pushgateway. It takes the following fields:
  - **url**: Url of the prometheus pushgateway.
  - **auth**: mTLS or tls + password authentication parameters. It takes the following fields:
    - **ca_cert**: Path to a CA certificate validating the server certificates of the pushgateway.
    - **client_cert**: Client certificate to use to authentify itself to the pushgateway when using mTLS.
    - **client_key**: Client private key to use to authentify itself to the pushgateway when using mTLS.
    - **password_auth**: Yaml file containing a **username** and **password** key to authentify against a pushgateway using password authentication.

The **backend_migration** parameter takes the following fields:
  - **current_backend**: File name of the current backend to migrate from. It is assumed to be relative filename that will be part of the files assembled in the working directory.
  - **next_backend**: Absolute file name of the next backend to migrate to. It is assumed to be an absolute path not present in the working directory.

The **termination_hooks** parameter takes the following fields:
  - **always**: Always call a hook. If defined, it will always override the success/failure/skip hooks.
  - **success**: Hook to call when the terraform command succeeds.
  - **failure**: Hook to call when the terraform command fails.
  - **skip**: Hook to call when the terraform command is skipped due to the recurrence rule.

Each termination hook has the following fields:
  - **command**: Defines a command to run with its arguments
  - **http_call**: Defines a simple http call to make against a remote endpoint

The command hook has the following fields:
  - **command**: Main command to run
  - **args**: Array of arguments to pass to the command

The http call hook has the following fields:
  - **method**: Http method to use
  - **endpoint**: Fully defined endpoint to call

Example of a config file to run terraform apply:

```
terraform_path: /home/myuser/bin/terraform
command: apply
timeouts:
  terraform_init: "15m"
  terraform_plan: "15m"
  terraform_apply: "1h"
  terraform_destroy: "1h"
sources:
  - repo:
      url: "git@github.com:mygituser/terracd-test.git"
      ref: "main"
      path: "dir1"
      auth:
        ssh_key_path: "/home/myuser/terracd/id_rsa"
        known_hosts_path: "/home/myuser/terracd/known_hosts"
      gpg_public_keys_paths:
        - /home/myuser/dirhavingkeyrings
        - /home/myuser/armoredkeyringfile.asc
  - dir: "/home/myuser/terracd-test/dir2"
termination_hooks:
  success:
    http_call:
      method: POST
      endpoint: http://127.0.0.1/terraform_success
  failure:
    command:
      command: "./email_support.sh"
      args:
        - "--important"
```

Example of a config file to run a backend migration:

```
terraform_path: /home/myuser/bin/terraform
command: migrate_backend
timeouts:
  terraform_init: "15m"
  terraform_plan: "15m"
  terraform_apply: "1h"
  terraform_destroy: "1h"
backend_migration:
  current_backend: "backend.tf"
  next_backend: "/home/myuser/nextbackenddir/backend.tf"
sources:
  - repo:
      url: "git@github.com:mygituser/terracd-test.git"
      ref: "main"
      path: "dir1"
      auth:
        ssh_key_path: "/home/myuser/terracd/id_rsa"
        known_hosts_path: "/home/myuser/terracd/known_hosts"
      gpg_public_keys_paths:
        - /home/myuser/dirhavingkeyrings
        - /home/myuser/armoredkeyringfile.asc
  - dir: "/home/myuser/currentbackenddir"
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
Aborting as forbidden operation is about to be performed on protected resource "module.filemon.local_file.file"
```

# Running End to End Tests

You can run end to end tests locally by running `go test` at the root of the project.

If you want to emulate the end to end test pipeline instead, you can type: `./e2e_test/docker-runtime/test.sh`. Note that you will need to have docker installed locally to run the test as they run in the pipeline.