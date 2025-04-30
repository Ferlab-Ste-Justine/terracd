#!/bin/bash -e

apt-get update && apt-get install unzip

#Setup Terraform
curl -L https://releases.hashicorp.com/terraform/1.4.6/terraform_1.4.6_linux_amd64.zip -o /tmp/terraform.zip
unzip /tmp/terraform.zip
mv terraform /usr/local/bin/terraform
rm /tmp/terraform.zip

#Setup Etcd
curl -L https://github.com/etcd-io/etcd/releases/download/v3.5.8/etcd-v3.5.8-linux-amd64.tar.gz -o /tmp/etcd.tar.gz
mkdir /tmp/etcd
tar xzvf /tmp/etcd.tar.gz -C /tmp/etcd
mv /tmp/etcd/etcd-v3.5.8-linux-amd64/etcd /usr/local/bin/etcd
rm /tmp/etcd.tar.gz

#Setup Gitea
curl -L https://github.com/go-gitea/gitea/releases/download/v1.23.6/gitea-1.23.6-linux-amd64 -o /tmp/gitea
chmod +x /tmp/gitea
cp /tmp/gitea /usr/local/bin/gitea

if [ -n "$EXEC_ID" ]; then
    groupadd -g $EXEC_ID testuser
    useradd testuser -u $EXEC_ID -g $EXEC_ID -m -s /bin/bash
    su testuser /opt/entrypoint_run_tests.sh
else
    /opt/entrypoint_run_tests.sh
fi