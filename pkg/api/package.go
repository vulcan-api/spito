package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/avorty/spito/pkg/shared"
	"github.com/go-git/go-git/v5"
	"github.com/oleiade/reflections"
	"github.com/schollz/progressbar/v3"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

/* #cgo LDFLAGS: -lalpm
   //#include "install_packages.h"
   #include <alpm.h>
*/
import "C"

const (
	packageManager         = "pacman" // Currently we only support arch pacman
	installCommand         = "-S"
	installFromFileOption  = "-U"
	noConfirmOption        = "--noconfirm"
	removeCommand          = "-Rns"
	changeUserCommand      = "/usr/bin/sudo"
	changeUserOption       = "-u"
	commandOption          = "-c"
	aurHelper              = "yay"
	pacmanDatabaseLocation = "/var/lib/pacman"
	rootLocation           = "/"
	successStatus          = 0
	aurAPIRequestURL       = "https://aur.archlinux.org/rpc/v5/info"
	aurCloneTemplate       = "https://aur.archlinux.org/%s.git"
	defaultCacheLocation   = "~/.cache"
	makepkgCommand         = "makepkg"
	makePkgOptions         = "-s"
)

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
	BuildDate     string
	InstallDate   string
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

type AurPackage struct {
	Name string
}

type AurResponseLayout struct {
	Results []AurPackage
}

func getListOfAURPackages(packages ...string) ([]string, error) {

	requestValues := url.Values{
		"arg[]": packages,
	}
	requestUrl := aurAPIRequestURL + "?" + requestValues.Encode()
	response, err := http.Get(requestUrl)
	if err != nil {
		return []string{}, err
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return []string{}, err
	}
	err = response.Body.Close()
	if err != nil {
		return []string{}, err
	}

	var jsonBody AurResponseLayout
	err = json.Unmarshal(body, &jsonBody)

	if err != nil {
		return []string{}, err
	}

	result := []string{}
	for _, aurPackage := range jsonBody.Results {
		result = append(result, aurPackage.Name)
	}
	return result, nil
}

func installPackageFromFile(packageName string, workingDirectory string) error {
	shared.ChangeToRoot()
	const pacmanPackageFileExtension = ".tar.zst"
	err := filepath.Walk(workingDirectory, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasPrefix(info.Name(), packageName) ||
			!strings.HasSuffix(info.Name(), pacmanPackageFileExtension) {
			return nil
		}

		packageManagerCommand :=
			exec.Command(packageManager, installFromFileOption, noConfirmOption, filepath.Join(workingDirectory, info.Name()))
		return packageManagerCommand.Run()
	})
	return err
}

func installAurPackages(packages []string, bar *progressbar.ProgressBar) error {
	cachePath := filepath.Join(
		shared.GetEnvWithDefaultValue("XDG_CACHE_HOME", defaultCacheLocation),
		"spito")

	shared.ChangeToUser()
	err := shared.ExpandTilde(&cachePath)
	if err != nil {
		return err
	}
	err = os.MkdirAll(cachePath, shared.DirectoryPermissions)
	if err != nil {
		return err
	}

	for _, pkg := range packages {
		repoPath := filepath.Join(cachePath, pkg)
		if doesExist, _ := shared.PathExists(repoPath); doesExist {
			err = os.RemoveAll(repoPath)
			if err != nil {
				return err
			}
		}
		bar.Describe(fmt.Sprintf("Cloning AUR package %s...", pkg))
		_, err = git.PlainClone(repoPath, false, &git.CloneOptions{
			URL: fmt.Sprintf(aurCloneTemplate, pkg),
		})
		if err != nil {
			return err
		}

		bar.Describe(fmt.Sprintf("Building AUR package %s...", pkg))
		argv := []string{changeUserCommand, changeUserOption, shared.GetRegularUser().Username, makepkgCommand}
		makePkgCommand, err := os.StartProcess(changeUserCommand, argv, &os.ProcAttr{
			Dir: repoPath,
		})
		if err != nil {
			return err
		}
		_, err = makePkgCommand.Wait()
		if err != nil {
			return err
		}
		bar.Describe(fmt.Sprintf("Installing AUR package %s...", pkg))
		err = installPackageFromFile(pkg, repoPath)
		if err != nil {
			return err
		}
	}
	_ = bar.Add(1)
	return nil
}

func installRegularPackages(packages ...string) error {
	packageManagerCommand := exec.Command(packageManager, installCommand, noConfirmOption, strings.Join(packages, " "))
	packageManagerCommand.Stdout = os.Stdout
	packageManagerCommand.Stderr = os.Stderr
	packageManagerCommand.Stdin = os.Stdin
	return packageManagerCommand.Run()
	/*shared.ChangeToRoot()

	alpmHandle := C.alpm_initialize(C.CString(rootLocation), C.CString(pacmanDatabaseLocation), err)
	if alpmHandle == nil {
		return fmt.Errorf("couldn't initialize the alpm library")
	}

	packageDatabases := C.alpm_get_syncdbs(alpmHandle)
	if packageDatabases == nil {
		C.alpm_release(alpmHandle)
		return fmt.Errorf("couldn't fetch package databases")
	}

	C.alpm_trans_init(alpmHandle, 0)
	for _, packageToInstall := range packages {
		packageToInstallCString := C.CString(packageToInstall)
		database := (*packageDatabases).data
		pkg := C.alpm_db_get_pkg((*C.alpm_db_t)(database), packageToInstallCString)
		for packageDatabases != nil && pkg == nil {
			fmt.Println(database, pkg, packageDatabases)
			database = (*packageDatabases).data
			pkg = C.alpm_db_get_pkg((*C.alpm_db_t)(database), packageToInstallCString)
			packageDatabases = C.alpm_list_next(packageDatabases)
		}
		if pkg == nil {
			C.alpm_release(alpmHandle)
			return fmt.Errorf("couldn't find the package %s in the pacman database", packageToInstall)
		}
		C.alpm_add_pkg(alpmHandle, pkg)
	}

	result := C.alpm_trans_prepare(alpmHandle, nil)
	if result != successStatus {
		C.alpm_release(alpmHandle)
		return errors.New("couldn't prepare the pacman transaction")
	}

	result = C.alpm_trans_commit(alpmHandle, nil)
	if result != successStatus {
		C.alpm_release(alpmHandle)
		return errors.New("couldn't install pacman packages")
	}

	result = C.alpm_trans_release(alpmHandle)
	if result != successStatus {
		C.alpm_release(alpmHandle)
		return errors.New("couldn't release the pacman transaction")
	}

	C.alpm_release(alpmHandle)
	return nil*/
}

func InstallPackages(packageStrings ...string) error {

	/* Determine packages to install/update */
	shared.ChangeToRoot()
	var packagesToInstall []string //[]*C.char
	for _, packageString := range packageStrings {
		packageName, version, _ := strings.Cut(packageString, "@")
		packageToBeInstalled, err := GetPackage(packageName)

		var expectedVersion string
		if len(version) > 0 {
			expectedVersion = version[1:]
		} else {
			expectedVersion = ""
		}

		doesPackageNeedToBeUpgraded := err == nil && packageToBeInstalled.Version < expectedVersion
		isPackageNotInstalled := err != nil

		if version == "" || version == "*" || isPackageNotInstalled || doesPackageNeedToBeUpgraded {
			packagesToInstall = append(packagesToInstall, packageName /*C.CString(packageName)*/)
		}
	}

	/* Get list of AUR packages */
	aurPackagesToInstall, err := getListOfAURPackages(packagesToInstall...)
	if err != nil {
		return err
	}

	/* Exclude AUR packages from the packagesToInstall slice */
	packagesToInstall = slices.DeleteFunc(packagesToInstall, func(pkg string) bool {
		return slices.Index(aurPackagesToInstall, pkg) != -1
	})

	if len(aurPackagesToInstall) > 0 {
		aurBar := progressbar.NewOptions(len(aurPackagesToInstall),
			progressbar.OptionSetDescription("Installing AUR packages..."),
			progressbar.OptionSetPredictTime(false),
			progressbar.OptionSetElapsedTime(false),
			progressbar.OptionShowCount(),
		)
		err = installAurPackages(aurPackagesToInstall, aurBar)
		if err != nil {
			return err
		}
		fmt.Println()
	}

	if len(packagesToInstall) == 0 {
		return nil
	}

	err = installRegularPackages(packagesToInstall...)
	shared.ChangeToUser()
	return err
}

func RemovePackages(packagesToRemove ...string) error {
	shared.ChangeToRoot()
	pacmanCommand := exec.Command(packageManager, removeCommand, noConfirmOption, strings.Join(packagesToRemove, " "))
	err := pacmanCommand.Run()
	shared.ChangeToUser()
	return err
}
