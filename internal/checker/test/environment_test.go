package test

import (
	"bytes"
	"github.com/avorty/spito/cmd/cmdApi"
	"github.com/avorty/spito/internal/checker"
	daemon_tracker "github.com/avorty/spito/pkg"
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

    local err = api.fs.createFile(path, defaultContent, false)
    if err ~= nil then
        api.info.error(err)
        return false
    end

    local content, err = api.fs.readFile(path)
    if err ~= nil then
        api.info.error(err)
        return false
    end

    if content ~= defaultContent then
        api.info.error("Content of file: " .. path .. "is: " .. content .. "instead of: " .. defaultContent)
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
		Content:   "it should be reverted after applying other environment",
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

	secondFileContent, err := os.ReadFile(secondEnvTemplate.Path)
	if err != nil {
		t.Fatal(err)
	}
	if string(secondFileContent) != secondEnvTemplate.Content {
		t.Fatalf("File %s has \"%s\" content instead of \"%s\"\n",
			secondEnvTemplate.Path,
			secondFileContent,
			secondEnvTemplate.Content,
		)
	}
}

func applyRule(templateData templateDataT, t *testing.T) templateDataT {
	templateData.Path = "/tmp/test-file-" + shared.RandomLetters(10)

	ruleSourceCode := getSourceCode(t, templateData)
	importLoopData := getImportLoopData(t)

	scriptDir, err := os.MkdirTemp("/tmp", "spito-rules-")
	if err != nil {
		t.Fatal(err.Error())
	}

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
		VRCT:          *ruleVRCT,
		InfoApi:       cmdApi.InfoApi{},
		RulesHistory:  shared.RulesHistory{},
		DaemonTracker: daemon_tracker.NewDaemonTracker(),
		ErrChan:       make(chan error),
	}
}

func getSourceCode(t *testing.T, templateData templateDataT) string {
	tmpl, err := template.New("").Parse(testTemplate)
	if err != nil {
		t.Fatal("Failed to render parsed template, error: \n" + err.Error())
	}

	var sourceCode bytes.Buffer
	if err := tmpl.Execute(&sourceCode, templateData); err != nil {
		t.Fatal("Failed to execute template\n" + err.Error())
	}

	return sourceCode.String()
}
