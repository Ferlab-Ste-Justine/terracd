package main

import (
	_ "embed"
	"fmt"
	"os"
	"path"
	"text/template"
)

var (
	//go:embed backend_http.tf
	BackendHttpTemplate string
)

func generateBackendAddress(address Address) string {
	url := address.Base
	
	if len(address.QueryString) > 0 {
		url = fmt.Sprintf("%s?%s", url, address.QueryString.Encode())
	}

	return url
}

func GenerateBackendFile(filePath string, backend BackendHttp) error {
	funcMap := template.FuncMap{
		"genAddr": generateBackendAddress,
	}

	tmpl, tmplErr := template.New("backendHttp").Funcs(funcMap).Parse(BackendHttpTemplate)
	if tmplErr != nil {
		return tmplErr
	}

	f, openErr := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0700)
	if openErr != nil {
		return openErr
	}
	defer f.Close()

	execErr := tmpl.Execute(f, backend)
	return execErr
}

func GenerateBackendFiles(backendDir string, config Config) error {
	for _, source := range config.Sources {
		if source.BackendHttp.Filename != "" {
			genErr := GenerateBackendFile(path.Join(backendDir, source.BackendHttp.Filename), source.BackendHttp)
			if genErr != nil {
				return genErr
			}
		}
	}
	return nil
}