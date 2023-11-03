resource "local_file" "file" {
  content         = "A"
  file_permission = "0660"
  filename        = "${path.module}/../output/file"
}