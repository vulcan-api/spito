package test

import (
	"bytes"
	"github.com/avorty/spito/cmd/cmdApi"
	"github.com/avorty/spito/internal/checker"
	"github.com/avorty/spito/pkg/shared"
	"github.com/avorty/spito/pkg/vrct"
	"html/template"
	"os"
	"path/filepath"
	"testing"
)

const testTemplate = `
{{ .Decorator }}
function main ()
    local path = "{{ .Path }}"
    local defaultContent = "{{ .Content }}"

    local err = api.fs.CreateFile(path, defaultContent)
    if err ~= nil then
        api.info.Error(err)
        return false
    end

    local content, err = api.fs.ReadFile(path)
    if err ~= nil then
        api.info.Error(err)
        return false
    end

    if content ~= defaultContent then
        api.info.Error("Content of file: " .. path .. "is: " .. content .. "instead of: " .. defaultContent)
        return false
    end

    return true
end
`

type templateDataT struct {
	Path      string
	Content   string
	Decorator string
}

func TestRevertingPreviousEnv(t *testing.T) {
	firstEnvTemplate := templateDataT{
		Content:   "it should be reverted after applying other env",
		Decorator: "#![environment]",
	}
	firstEnvTemplate = applyRule(firstEnvTemplate, t)

	secondEnvTemplate := templateDataT{
		Content:   "it should be alive",
		Decorator: "#![environment]",
	}
	secondEnvTemplate = applyRule(secondEnvTemplate, t)

	firstFileExists, err := shared.PathExists(firstEnvTemplate.Path)
	if err != nil {
		t.Fatal(err.Error())
	}
	if firstFileExists {
		t.Fatalf("File %s should be reverted\n", firstEnvTemplate.Path)
	}

	secondFileExists, err := shared.PathExists(secondEnvTemplate.Path)
	if err != nil {
		t.Fatal(err.Error())
	}
	if secondFileExists {
		t.Fatalf("File %s should be reverted\n", secondEnvTemplate.Path)
	}
}

func applyRule(templateData templateDataT, t *testing.T) templateDataT {
	testFile, err := os.CreateTemp("/tmp", "test-file-")
	if err != nil {
		t.Fatal(err.Error())
	}

	defer func() {
		if err := testFile.Close(); err != nil {
			t.Fatal("Failed to close file\n " + err.Error())
		}
	}()

	templateData.Path = testFile.Name()

	ruleSourceCode := getSourceCode(t, templateData)
	importLoopData := getImportLoopData(t)

	scriptDir, err := os.MkdirTemp("/tmp", "spito-rules-")

	ruleFile, err := os.Create(filepath.Join(scriptDir, "rule.lua"))
	if err != nil {
		t.Fatal(err.Error())
	}

	if _, err := ruleFile.WriteString(ruleSourceCode); err != nil {
		t.Fatal(err.Error())
	}
	defer func() {
		if err := ruleFile.Close(); err != nil {
			t.Fatal("Failed to close file\n " + err.Error())
		}
	}()

	if err := checker.ApplyEnvironmentScript(importLoopData, ruleSourceCode, ruleFile.Name()); err != nil {
		t.Fatal(err.Error())
	}

	firstEnvContent, err := os.ReadFile(templateData.Path)
	if err != nil {
		t.Fatal(err.Error())
	}

	if string(firstEnvContent) != templateData.Content {
		t.Fatalf("after applying environment test file (%s) content is incorrect", templateData.Path)
	}

	return templateData
}

func getImportLoopData(t *testing.T) *shared.ImportLoopData {
	ruleVRCT, err := vrct.NewRuleVRCT()
	if err != nil {
		t.Fatal(err.Error())
	}

	return &shared.ImportLoopData{
		VRCT:         *ruleVRCT,
		InfoApi:      cmdApi.InfoApi{},
		RulesHistory: shared.RulesHistory{},
		ErrChan:      make(chan error),
	}
}

func getSourceCode(t *testing.T, templateData templateDataT) string {
	tmpl, err := template.New("").Parse(testTemplate)
	if err != nil {
		t.Fatal("Failed to render parse template, error: \n" + err.Error())
	}

	var sourceCode bytes.Buffer
	if err := tmpl.Execute(&sourceCode, templateData); err != nil {
		t.Fatal("Failed to execute template\n" + err.Error())
	}

	return sourceCode.String()
}
