resource "local_file" "file_other" {
  content         = "O"
  file_permission = "0660"
  filename        = "${path.module}/../output/file-other"
}