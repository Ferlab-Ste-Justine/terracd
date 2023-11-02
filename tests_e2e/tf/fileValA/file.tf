resource "local_file" "file" {
  content         = "A"
  file_permission = "0660"
  filename        = "output/file"
}