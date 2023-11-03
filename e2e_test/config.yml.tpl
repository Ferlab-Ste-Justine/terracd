terraform_path: {{ .TerraformPath }}
command: {{ .Command }}
working_directory: e2e_test/runtime
sources:
{{range .Sources -}}
  - dir: "{{.Dir}}"
{{end -}}
state_store:
  fs:
    enabled: true
cache:
  versions_file: "version.tf"
random_jitter: "{{ .Jitter }}"
recurrence:
  min_interval: "{{ .MinInterval }}"
  git_triggers: true
#metrics:
#  job_name: terracd_metrics
#  pushgateway:
#    url: https://127.0.0.1:9091
#    auth:
#      ca_cert: pushgateway/certs/local_ca.crt
#      password_auth: pushgateway/confs/password.yml
termination_hooks:
  success:
    command:
      command: e2e_test/hook.sh
      args:
        - success
  skip:
    command:
      command: e2e_test/hook.sh
      args:
        - skip
  failure:
    command:
      command: e2e_test/hook.sh
      args:
        - failure