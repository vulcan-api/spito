package api

import (
	"github.com/avorty/spito/pkg/vrct/vrctFs"
	"github.com/avorty/spito/pkg/path"
	"github.com/go-git/go-git/v5"
	"os"
)

type GitApi struct {
	FsVrct *vrctFs.VRCTFs
}

func (g *GitApi) GitClone(repoUrl, destinationPath string) error {
	if err := os.MkdirAll("/tmp/spito", path.DirectoryPermissions); err != nil {
		return err
	}

	tmpDir, err := os.MkdirTemp("/tmp/spito", "git-clone-tmp-")
	if err != nil {
		return err
	}

	if _, err = git.PlainClone(tmpDir, false, &git.CloneOptions{URL: repoUrl}); err != nil {
		return err
	}

	if err := g.FsVrct.Move(tmpDir, destinationPath); err != nil {
		return err
	}

	return os.RemoveAll(tmpDir)
}
