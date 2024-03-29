package source

import (
	_ "embed"
	"fmt"
	"net/url"
	"os"
	"path"
	"text/template"
)

type Address struct {
	Base        string
	QueryString url.Values `yaml:"query_string"`
}

type BackendHttp struct {
	Filename      string
	Address       Address
	UpdateMethod  string  `yaml:"update_method"`
	LockAddress   Address `yaml:"lock_address"`
	LockMethod    string  `yaml:"lock_method"`
	UnlockAddress Address `yaml:"unlock_address"`
	UnlockMethod  string  `yaml:"unlock_method"`
}

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

func (backend *BackendHttp) GenerateFile(filePath string) error {
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

func (srcs *Sources) GenerateBackendFiles(backendDir string) error {
	for _, source := range *srcs {
		if source.GetType() == TypeBackendHttp {
			genErr := source.BackendHttp.GenerateFile(path.Join(backendDir, source.BackendHttp.Filename))
			if genErr != nil {
				return genErr
			}
		}
	}

	return nil
}