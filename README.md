# About

This is a continuous delivery tool for terraform that supports a gitops methodology.

The tool merge several sources of terraform files together (git repo as well as filesystem), before applying the final result, allowing any files containing secrets (ex: provider, backend, etc) to be separated from the version-controlled files.

The tool runs a single iteration and then exits, relying on an external scheduler like systemd, kubernetes or cron to schedule recurrence.

# Usage

The tool expects a file named **config.yml** to be present in its running directory.

The file has the following top-level fields:
- **terraform_path**: Path to the terraform binary
- **sources**: Array of terraform file sources to be merged together and applied on

Each source can take one of the following two forms:
```
- dir: "<local directory with terraform scripts>"
  - repo:
      url: "<git repo ssh url>"
      ref: "<git repo branch>"
      path: "<path in git repo where scripts are>"
      auth:
        ssh_key_path: "<ssh key that has read access to the repo>"
        known_hosts_path: "<known host file containing the expect fingerprint of git server>"
```

Example of config file:

```
terraform_path: /home/myuser/bin/terraform
sources:
  - repo:
      url: "git@github.com:mygituser/terracd-test.git"
      ref: "main"
      path: "dir1"
      auth:
        ssh_key_path: "/home/myuser/terracd/id_rsa"
        known_hosts_path: "/home/myuser/terracd/known_hosts"
  - dir: "/home/myuser/terracd-test/dir2"
```

# Missing Functionality

The following functionality is planed, but not yet implemented:
- Support for signed commits
- Support for "prevent destroy" in the tool's configuration to circumvent this current limitation of Terraform: https://github.com/hashicorp/terraform/issues/18367
- Support for fluentd logger
- Support for running terraform plan on pull requests (at least for Github, probably for Gitlab and Gitea as well)