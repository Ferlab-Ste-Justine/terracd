resource "local_file" "other_file" {
  content         = "A"
  file_permission = "0660"
  filename        = "output/other_file"
}