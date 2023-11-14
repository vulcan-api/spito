package api

import (
	"os"
	"regexp"
	"strings"
)

func PathExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func FileExists(path string, isDirectory bool) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	if isDirectory && info.IsDir() || !isDirectory && !info.IsDir() {
		return true
	}

	return false
}

func ReadFile(path string) (string, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(file), nil
}

func ReadDir(path string) ([]os.DirEntry, error) {
	return os.ReadDir(path)
}

func removeRanges(fileContent string, rangeStart string, rangeEnd string, removeRangeEnd bool) string {
	cleanFile := ""
	slice := fileContent
	sliceLen := len(slice)
	endLen := 0
	if removeRangeEnd {
		endLen = len(rangeEnd)
	}

	commentPos := strings.Index(slice, rangeStart)
	for commentPos != -1 {
		cleanFile += slice[0:commentPos]
		slice = slice[commentPos:sliceLen]

		sliceLen = len(slice)
		realEndPos := strings.Index(slice, rangeEnd) + endLen
		if realEndPos == -1 {
			realEndPos = len(slice)
		}

		slice = slice[realEndPos:sliceLen]
		sliceLen = len(slice)
		commentPos = strings.Index(slice, rangeStart)
	}
	cleanFile += slice[0:sliceLen]
	return cleanFile
}

func RemoveComments(fileContent string, singleLineComment string, multilineCommentStart string, multilineCommentEnd string) string {
	cleanFile := fileContent
	if singleLineComment != "" {
		cleanFile = removeRanges(fileContent, singleLineComment, "\n", false)
	}
	if multilineCommentStart != "" && multilineCommentEnd != "" {
		cleanFile = removeRanges(cleanFile, multilineCommentStart, multilineCommentEnd, true)
	}

	return cleanFile
}

func FileContains(fileContent string, content string) bool {
	return strings.Contains(fileContent, content)
}

func Find(regex string, fileContent string) ([]int, error) {
	r, err := regexp.Compile(regex)
	if err != nil {
		return nil, err
	}
	return r.FindStringIndex(fileContent), nil
}

func FindAll(regex string, fileContent string) ([][]int, error) {
	r, err := regexp.Compile(regex)
	if err != nil {
		return nil, err
	}
	return r.FindAllStringIndex(fileContent, -1), nil
}

func GetProperLines(regex string, fileContent string) ([]string, error) {
	indexesInLines, err := FindAll(regex, fileContent)
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
