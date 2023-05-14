terraform {
  backend "http" {
{{- if ne .UpdateMethod ""}}    
    update_method = "{{.UpdateMethod}}"
{{- end}}
{{- if ne .LockMethod ""}}
    lock_method = "{{.LockMethod}}"
{{- end}}
{{- if ne .UnlockMethod ""}}
    unlock_method = "{{.UnlockMethod}}"
{{- end}}
{{- if ne .LockAddress.Base ""}}
    lock_address = "{{genAddr .LockAddress}}"
{{- end}}
{{- if ne .UnlockAddress.Base ""}}
    unlock_address = "{{genAddr .UnlockAddress}}"
{{- end}}
    address = "{{genAddr .Address}}"
  }
}