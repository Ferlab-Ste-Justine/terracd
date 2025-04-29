terraform_path: {{ .TerraformPath }}
command: {{ .Command }}
working_directory: e2e_test/runtime
sources:
{{- range .Sources}}
  - dir: "{{.Dir}}"
{{- end}}
state_store:
{{- if eq .State.Type "Fs" }}
  fs:
    enabled: true
{{ else if eq .State.Type "Etcd" }}
  etcd:
    prefix: "/state/"
    endpoints:
      - "127.0.0.1:3379"
      - "127.0.0.2:3379"
      - "127.0.0.3:3379"
    connection_timeout: "10s"
    request_timeout: "10s"
    retry_interval: "10s"
    retries: 3
    auth:
      ca_cert: "e2e_test/etcd-dependencies/certs/ca.crt"
      client_cert: "e2e_test/etcd-dependencies/certs/root.pem"
      client_key: "e2e_test/etcd-dependencies/certs/root.key"
{{- end}}
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