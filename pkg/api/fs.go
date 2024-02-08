package api

import (
	"github.com/avorty/spito/pkg/vrct/vrctFs"
	"os"
	"regexp"
	"strings"
)

type FsApi struct {
	FsVRCT *vrctFs.VRCTFs
}

func (f *FsApi) PathExists(path string) bool {
	_, err := f.FsVRCT.Stat(path)
	return !os.IsNotExist(err)
}

func (f *FsApi) FileExists(path string, isDirectory bool) bool {
	info, err := f.FsVRCT.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	if isDirectory && info.IsDir() || !isDirectory && !info.IsDir() {
		return true
	}

	return false
}

func (f *FsApi) ReadFile(path string) (string, error) {
	file, err := f.FsVRCT.ReadFile(path)
	if err != nil {
		return "", err
	}

	return string(file), nil
}

func (f *FsApi) ReadDir(path string) ([]os.DirEntry, error) {
	return f.FsVRCT.ReadDir(path)
}

func RemoveComments(fileContent, singleLineStart, multiLineStart, multiLineEnd string) string {
	var result strings.Builder
	isString := false
	isCharEscaped := false
	isMultilineComment := false
	isSingleLineComment := false

	isSingleLineSupported := singleLineStart != ""
	isMultiLineSupported := multiLineStart != ""

	for i := 0; i < len(fileContent); i++ {
		if !isCharEscaped {
			if strings.HasPrefix(fileContent[i:], multiLineStart) && !isString && !isSingleLineComment && isMultiLineSupported {
				isMultilineComment = true
			} else if strings.HasPrefix(fileContent[i:], singleLineStart) && !isString && !isMultilineComment && isSingleLineSupported {
				isSingleLineComment = true
			}

			if isMultilineComment && strings.HasPrefix(fileContent[i:], multiLineEnd) && !isString {
				i += 1
				isMultilineComment = false
				continue
			}

			switch fileContent[i] {
			case '"':
				isString = !isString
				break
			case '\\':
				isCharEscaped = !isCharEscaped
				break
			case '\n':
				isSingleLineComment = false
				break
			}
		} else if fileContent[i] != '\\' {
			isCharEscaped = false
		}

		if !(isMultilineComment && isMultiLineSupported) && !(isSingleLineComment && isSingleLineSupported) {
			result.WriteByte(fileContent[i])
		}
	}

	return result.String()
}

func (*FsApi) FileContains(fileContent string, content string) bool {
	return strings.Contains(fileContent, content)
}

func (*FsApi) Find(regex string, fileContent string) ([]int, error) {
	r, err := regexp.Compile(regex)
	if err != nil {
		return nil, err
	}
	return r.FindStringIndex(fileContent), nil
}

func (*FsApi) FindAll(regex string, fileContent string) ([][]int, error) {
	r, err := regexp.Compile(regex)
	if err != nil {
		return nil, err
	}
	return r.FindAllStringIndex(fileContent, -1), nil
}

func (f *FsApi) GetProperLines(regex string, fileContent string) ([]string, error) {
	indexesInLines, err := f.FindAll(regex, fileContent)
	if err != nil {
		return nil, err
	}

	fileLen := len(fileContent)

	var properLines []string
	for _, line := range indexesInLines {
		if line != nil {
			dataBefore := fileContent[0:line[0]]
			dataAfter := fileContent[line[1]:fileLen]
			startingLineEnd := strings.LastIndex(dataBefore, "\n")
			endingLineEnd := strings.Index(dataAfter, "\n")
			if startingLineEnd == -1 {
				startingLineEnd = 0
			} else {
				startingLineEnd++
			}
			if endingLineEnd == -1 {
				endingLineEnd = fileLen
			} else {
				endingLineEnd += line[1]
			}
			properLines = append(properLines, fileContent[startingLineEnd:endingLineEnd])
		}
	}

	return properLines, nil
}

func (f *FsApi) Apply() (int, error) {
	return f.FsVRCT.Apply()
}

func (f *FsApi) CreateFile(path, content string, optional bool) error {
	return f.FsVRCT.CreateFile(path, []byte(content), optional)
}

type CreateConfigOptions struct {
	Optional   bool
	Options    string
	ConfigType vrctFs.FileType
}

func (f *FsApi) CreateConfig(path, content string, options CreateConfigOptions) error {
	return f.FsVRCT.CreateConfig(path, []byte(content), []byte(options.Options), options.Optional, options.ConfigType)
}
func (f *FsApi) UpdateConfig(path, content string, options CreateConfigOptions) error {
	return f.FsVRCT.UpdateConfig(path, []byte(content), []byte(options.Options), options.Optional, options.ConfigType)
}

func (f *FsApi) CompareConfigs(received, desired []byte, configType uint) error {
	return vrctFs.CompareConfigs(received, desired, vrctFs.FileType(configType))
}
