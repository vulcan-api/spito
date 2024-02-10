package package_conflict

import (
	"errors"
	"fmt"
)

type PackageConflictTracker struct {
	packagesInstalled map[string]bool
	packagesRemoved   map[string]bool
}

func NewPackageConflictTracker() PackageConflictTracker {
	return PackageConflictTracker{
		packagesInstalled: make(map[string]bool),
		packagesRemoved:   make(map[string]bool),
	}
}

func (packageTracker PackageConflictTracker) AddPackage(packageName string) error {

	if _, isPackageUninstalled := packageTracker.packagesRemoved[packageName]; isPackageUninstalled {
		return errors.New(fmt.Sprintf("[PACKAGE_CONFLICT] the package %s is required to be uninstalled by a dependency", packageName))
	}

	packageTracker.packagesInstalled[packageName] = true
	return nil
}

func (packageTracker PackageConflictTracker) RemovePackage(packageName string) error {

	if _, isPackageInstalled := packageTracker.packagesInstalled[packageName]; isPackageInstalled {
		return errors.New(fmt.Sprintf("[PACKAGE_CONFLICT] the package %s is required to be installed by a dependency", packageName))
	}

	packageTracker.packagesRemoved[packageName] = true
	return nil
}

func (packageTracker PackageConflictTracker) GetPackagesToInstall() []string {
	result := make([]string, len(packageTracker.packagesInstalled))
	i := 0
	for packageName, _ := range packageTracker.packagesInstalled {
		result[i] = packageName
		i++
	}
	return result
}

func (packageTracker PackageConflictTracker) GetPackagesToRemove() []string {
	result := make([]string, len(packageTracker.packagesRemoved))
	i := 0
	for packageName, _ := range packageTracker.packagesRemoved {
		result[i] = packageName
		i++
	}
	return result
}
