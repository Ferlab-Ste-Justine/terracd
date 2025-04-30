resource "time_static" "time" {}

resource "tls_private_key" "gpg_key_1" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "tlsext_pgp_private_key" "gpg_key_1" {
    private_key = tls_private_key.gpg_key_1.private_key_pem_pkcs8
    timestamp = time_static.time.id
    name = "user1"
    email = "user1@email.com"
}

resource "local_file" "gpg_key_1_private" {
  content = tlsext_pgp_private_key.gpg_key_1.private_key_gpg_armor
  filename = "${path.module}/keys/gpg_key_1"
  file_permission = "0600"
}

resource "local_file" "gpg_key_1_public" {
  content = tlsext_pgp_private_key.gpg_key_1.public_key_gpg_armor
  filename = "${path.module}/keys/gpg_key_1.pub"
  file_permission = "0600"
}

resource "tls_private_key" "gpg_key_2" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "tlsext_pgp_private_key" "gpg_key_2" {
    private_key = tls_private_key.gpg_key_2.private_key_pem_pkcs8
    timestamp = time_static.time.id
    name = "user2"
    email = "user2@email.com"
}

resource "local_file" "gpg_key_2_private" {
  content = tlsext_pgp_private_key.gpg_key_2.private_key_gpg_armor
  filename = "${path.module}/keys/gpg_key_2"
  file_permission = "0600"
}

resource "local_file" "gpg_key_2_public" {
  content = tlsext_pgp_private_key.gpg_key_2.public_key_gpg_armor
  filename = "${path.module}/keys/gpg_key_2.pub"
  file_permission = "0600"
}

resource "tls_private_key" "gpg_key_3" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "tlsext_pgp_private_key" "gpg_key_3" {
    private_key = tls_private_key.gpg_key_3.private_key_pem_pkcs8
    timestamp = time_static.time.id
    name = "user3"
    email = "user3@email.com"
}

resource "local_file" "gpg_key_3_private" {
  content = tlsext_pgp_private_key.gpg_key_3.private_key_gpg_armor
  filename = "${path.module}/keys/gpg_key_3"
  file_permission = "0600"
}

resource "local_file" "gpg_key_3_public" {
  content = tlsext_pgp_private_key.gpg_key_3.public_key_gpg_armor
  filename = "${path.module}/keys/gpg_key_3.pub"
  file_permission = "0600"
}

resource "tls_private_key" "ssh_key" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "local_file" "ssh_key_private" {
  content = chomp(tls_private_key.ssh_key.private_key_pem)
  filename = "${path.module}/keys/ssh/id_rsa"
  file_permission = "0600"
}

resource "local_file" "ssh_key_pub" {
  content = chomp(tls_private_key.ssh_key.public_key_openssh)
  filename = "${path.module}/keys/ssh/id_rsa.pub"
  file_permission = "0600"
}