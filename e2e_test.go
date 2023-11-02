package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"text/template"
	"testing"
)

type TestConfTemplateSrc struct {
	Dir string
}

type TestConfTemplate struct {
	TerraformPath string
	Command       string
	MinInterval   string
	Jitter        string
	Sources       []TestConfTemplateSrc
}

func (tpl *TestConfTemplate) SetTfPath() error {
	tfPath, tfPathErr := exec.LookPath("terraform")
	if tfPathErr != nil {
		return tfPathErr
	}

	tpl.TerraformPath = tfPath

	return nil
}

func (tpl *TestConfTemplate) GenerateConfig() error {
	fContent, err := ioutil.ReadFile(path.Join("tests_e2e", "config.yml.tpl"))
	if err != nil {
		return err
	}

	tmpl, tErr := template.New("template").Parse(string(fContent))
	if tErr != nil {
		return tErr
	}

	var b bytes.Buffer
	exErr := tmpl.Execute(&b, tpl)
	if exErr != nil {
		return exErr
	}

	return os.WriteFile("config.yml", b.Bytes(), 0640)
}

func (tpl *TestConfTemplate) CleanupConfig() error {
	return os.Remove("config.yml")
}

func TestPlan(t *testing.T) {
	tpl := TestConfTemplate{
		Command: "plan",
		MinInterval: "10s",
		Jitter: "3s",
		Sources: []TestConfTemplateSrc{
			TestConfTemplateSrc{Dir: path.Join("tests_e2e", "tf", "fileValA")},
			TestConfTemplateSrc{Dir: path.Join("tests_e2e", "tf", "version")},
		},
	}

	err := tpl.SetTfPath()
	if err != nil {
		t.Errorf(err.Error())
	}

	err = tpl.GenerateConfig()
	if err != nil {
		t.Errorf(err.Error())
	}
	
	main()
	
	err = tpl.CleanupConfig()
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestApply(t *testing.T) {
}

func TestDestroy(t *testing.T) {

}