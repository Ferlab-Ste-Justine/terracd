cd /opt/code

(cd e2e_test/etcd-dependencies; terraform init; terraform apply -auto-approve)

(cd e2e_test/git-dependencies; terraform init; terraform apply -auto-approve)

go test