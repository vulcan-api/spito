package vrctFs

import (
	"fmt"
	"os"
)

const revertTmpPath = "/tmp/spito-vrct/fs-revert"

const (
	removeFile = iota
	removeDirAll
	replaceContent
)

type RevertStep struct {
	path   string
	action int
	// oldContentPath field is optional
	oldContentPath *string
}

type RevertSteps struct {
	steps         []RevertStep
	revertTempDir string
}

func NewRevertSteps() (RevertSteps, error) {
	err := os.MkdirAll(revertTmpPath, os.ModePerm)
	if err != nil {
		return RevertSteps{}, err
	}

	dir, err := os.MkdirTemp(revertTmpPath, "")
	if err != nil {
		return RevertSteps{}, err
	}

	return RevertSteps{
		revertTempDir: dir,
	}, nil
}

func (r *RevertSteps) RemoveFile(path string) {
	r.steps = append(r.steps, RevertStep{
		path:   path,
		action: removeFile,
	})
}

func (r *RevertSteps) RemoveDirAll(path string) {
	r.steps = append(r.steps, RevertStep{
		path:   path,
		action: removeDirAll,
	})
}

func (r *RevertSteps) BackupOldContent(path string) error {
	oldFile, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	tempContentFile, err := os.CreateTemp(r.revertTempDir, "old-")
	if err != nil {
		return err
	}

	defer func() {
		if err := tempContentFile.Close(); err != nil {
			println("Failed to close file, it may cause memory leak\n", err.Error())
		}
	}()

	_, err = tempContentFile.Write(oldFile)
	if err != nil {
		return err
	}

	contentPath := r.revertTempDir + "/" + tempContentFile.Name()
	r.steps = append(r.steps, RevertStep{
		path:           path,
		action:         replaceContent,
		oldContentPath: &contentPath,
	})
	return nil
}

func (r *RevertStep) Apply() error {
	switch r.action {
	case removeFile:
		return os.Remove(r.path)
	case removeDirAll:
		return os.RemoveAll(r.path)
	case replaceContent:
		if r.oldContentPath == nil {
			return fmt.Errorf("oldContentPath cannot be null if RevertStep action is: \"replaceContent\"\n")
		}
		err := os.RemoveAll(r.path)
		if err != nil {
			return err
		}

		return os.Rename(*r.oldContentPath, r.path)
	default:
		return fmt.Errorf("unknown RevertStep action: %d\n", r.action)
	}
}

func (r *RevertSteps) Apply() error {
	for _, step := range r.steps {
		if err := step.Apply(); err != nil {
			return err
		}
	}
	return nil
}
