#!/bin/bash -e

apt-get update && apt-get install unzip
curl -L https://releases.hashicorp.com/terraform/1.4.6/terraform_1.4.6_linux_amd64.zip -o /tmp/terraform.zip
unzip /tmp/terraform.zip
mv terraform /usr/local/bin/terraform
rm /tmp/terraform.zip

cd /opt/code

go test