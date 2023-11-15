package api

import (
	"errors"
	"fmt"
	"github.com/oleiade/reflections"
	"os/exec"
	"strings"
)

const packageManager = "pacman" // Currently we only support arch pacman

type Package struct {
	Name          string
	Version       string
	Description   string
	Architecture  string
	URL           string
	Licenses      []string
	Groups        []string
	Provides      []string
	DependsOn     []string
	OptionalDeps  []string
	RequiredBy    []string
	OptionalFor   []string
	ConflictsWith []string
	Replaces      []string
	InstalledSize []string
	Packager      string
	BuildDate     string //TODO: consider some kind of date type
	InstallDate   string //TODO: consider some kind of date type
	InstallReason string
	InstallScript bool
	ValidatedBy   string
}

func iFErrPrint(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
	}
}

func getPackageInfoString(packageName string, packageManager string) (string, error) {
	cmd := exec.Command(packageManager, "-Qi", packageName)
	cmd.Env = append(cmd.Environ(), "LANG=C")
	data, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (p *Package) setField(key string, value string) {
	fieldType, _ := reflections.GetFieldType(p, key)
	if value == "None" {
		return
	}

	switch fieldType {
	case "string":
		err := reflections.SetField(p, key, value)
		iFErrPrint(err)

		break
	case "[]string":
		values := strings.Split(value, " ")
		err := reflections.SetField(p, key, values)
		iFErrPrint(err)

		break
	case "bool":
		err := reflections.SetField(p, key, value == "Yes")
		iFErrPrint(err)

		break
	default:
		err := errors.New("Handling type " + fieldType + " in Package is not implemented")
		panic(err)
	}
}

func GetPackage(name string) (Package, error) {
	p := Package{}

	packageInfoString, err := getPackageInfoString(name, packageManager)
	if err != nil {
		return Package{}, err
	}
	packageInfo := strings.Split(packageInfoString, "\n")
	packageInfo = packageInfo[:len(packageInfo)-2] // Delete empty elements

	var multiLineValue string
	var multiLineKey string

	for index, line := range packageInfo {
		sides := strings.Split(line, ":")

		// Not only trim, we also change e.g. "Depends On" to "DependsOn"
		key := strings.ReplaceAll(sides[0], " ", "")

		// Handling potential ":" in value
		values := sides[1:]
		value := strings.Trim(strings.Join(values, ":"), " ")

		isNextLineValueOnly := false
		// -2 because we later use index + 1
		if index <= len(packageInfo)-2 {
			isNextLineValueOnly = packageInfo[index+1][0] == ' '
		}

		// if next line is still value of our key
		if isNextLineValueOnly {
			if len(multiLineKey) == 0 {
				multiLineKey = key
				multiLineValue = value
			} else {
				multiLineValue += line
			}
			continue
		}

		if len(multiLineKey) != 0 {
			p.setField(multiLineKey, multiLineValue)

			multiLineKey = ""
			multiLineValue = ""
			continue
		}

		p.setField(key, value)
	}
	return p, nil
}
