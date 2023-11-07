package api

import (
	"os"
	"strings"
)

func PathExists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func FileExists(path string, isDirectory bool) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	if isDirectory && info.IsDir() {
		return true
	}
	if !isDirectory && !info.IsDir() {
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

func removeRanges(file string, rangeStart string, rangeEnd string, removeRangeEnd bool) string {
	clearFile := ""
	slice := file
	sliceLen := len(slice)
	endLen := 0
	if removeRangeEnd {
		endLen = len(rangeEnd)
	}

	for {
		commentPos := strings.Index(slice, rangeStart)
		if commentPos == -1 {
			break
		}
		clearFile += slice[0:commentPos]
		slice = slice[commentPos:sliceLen]

		sliceLen = len(slice)
		realEndPos := strings.Index(slice, rangeEnd) + endLen

		slice = slice[realEndPos:sliceLen]
		sliceLen = len(slice)
	}
	clearFile += slice[0:sliceLen]
	return clearFile
}

func RemoveComments(file string, singleLineComment string, multilineCommentStart string, multilineCommentEnd string) string {
	// single line comments
	withoutSingleLineComments := removeRanges(file, singleLineComment, "\n", false)
	clearFile := removeRanges(withoutSingleLineComments, multilineCommentStart, multilineCommentEnd, true)

	return clearFile
}

func FileContains(file string, content string) bool {
	return strings.Contains(file, content)
}
