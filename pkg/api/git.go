package api

import (
	"github.com/avorty/spito/pkg/vrct/vrctFs"
	"github.com/go-git/go-git/v5"
	"os"
	"path/filepath"
)

type GitApi struct {
	FsVrct *vrctFs.VRCTFs
}

func (g *GitApi) GitClone(repoUrl, destinationPath string) error {
	if err := os.MkdirAll("/tmp/spito", os.ModePerm); err != nil {
		return err
	}

	tmpDir, err := os.MkdirTemp("/tmp/spito", "git-clone-tmp-")
	if err != nil {
		return err
	}

	if _, err = git.PlainClone(tmpDir, false, &git.CloneOptions{URL: repoUrl}); err != nil {
		return err
	}

	if err := moveRecursively(g.FsVrct, tmpDir, destinationPath); err != nil {
		return err
	}

	return os.RemoveAll(tmpDir)
}

func moveRecursively(fsVrct *vrctFs.VRCTFs, from, to string) error {
	fromEntries, err := os.ReadDir(from)
	if err != nil {
		return err
	}

	for _, fromEntry := range fromEntries {
		fromPath := filepath.Join(from, fromEntry.Name())
		toPath := filepath.Join(to, fromEntry.Name())

		if fromEntry.IsDir() {
			if err := moveRecursively(fsVrct, fromPath, toPath); err != nil {
				return err
			}
			continue
		}
		fileContent, err := fsVrct.ReadFile(fromPath)
		if err != nil {
			return err
		}

		if err := fsVrct.CreateFile(toPath, fileContent, false); err != nil {
			return err
		}
	}

	return nil
}
