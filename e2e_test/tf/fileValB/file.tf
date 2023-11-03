resource "local_file" "file" {
  content         = "B"
  file_permission = "0660"
  filename        = "output/file"
}