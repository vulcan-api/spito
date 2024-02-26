package vrctFs

import (
	"fmt"
	"github.com/BaderBC/targz"
	"github.com/avorty/spito/pkg/path"
	"gopkg.in/mgo.v2/bson"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const revertTmpPath = "/tmp/spito-vrct/fs-revert"
const revertStepsBsonName = "revert_steps.bson"

const (
	removeFile = iota
	removeDirAll
	replaceContent
)

func GetSerializedRevertStepsDir() (string, error) {
	homeDir := path.UserHomeDir
	// cannot use shared.LocalStateSpitoPath, because it creates incorrect golang import loop
	dir := filepath.Join(homeDir, ".local/state/spito/revert-steps-serialized")

	// Ensure exist
	err := os.MkdirAll(dir, os.ModePerm)

	return dir, err
}

type RevertStep struct {
	Path   string `bson:"Path"`
	Action int    `bson:"Action"`
	// OldContentPath field is optional
	OldContentPath string `bson:"OldContentPath"`
}

type RevertSteps struct {
	Steps         []RevertStep `bson:"Steps"`
	RevertTempDir string       `bson:"-"`
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
		RevertTempDir: dir,
	}, nil
}

func (r *RevertSteps) RemoveFile(path string) {
	r.Steps = append(r.Steps, RevertStep{
		Path:   path,
		Action: removeFile,
	})
}

func (r *RevertSteps) RemoveDirAll(path string) {
	r.Steps = append(r.Steps, RevertStep{
		Path:   path,
		Action: removeDirAll,
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

	tempContentFile, err := os.CreateTemp(r.RevertTempDir, "old-")
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

	r.Steps = append(r.Steps, RevertStep{
		Path:           path,
		Action:         replaceContent,
		OldContentPath: tempContentFile.Name(),
	})
	return nil
}

func (r *RevertStep) Apply() error {
	switch r.Action {
	case removeFile:
		return os.Remove(r.Path)
	case removeDirAll:
		return os.RemoveAll(r.Path)
	case replaceContent:
		if r.OldContentPath == "" {
			return fmt.Errorf("oldContentPath cannot be null if RevertStep action is: \"replaceContent\"\n")
		}
		err := os.RemoveAll(r.Path)
		if err != nil {
			return err
		}

		return MoveFile(r.OldContentPath, r.Path)
	default:
		return fmt.Errorf("unknown RevertStep action: %d\n", r.Action)
	}
}

func (r *RevertSteps) Apply() error {
	for _, step := range r.Steps {
		if err := step.Apply(); err != nil {
			return err
		}
	}
	return nil
}

func (r *RevertSteps) DeleteRuntimeTemp() error {
	return os.RemoveAll(r.RevertTempDir)
}

// Serialize 1st return value is number which should be provided in order to deserialize propert RevertSteps
func (r *RevertSteps) Serialize() (int, error) {
	outBson, err := bson.Marshal(r)
	if err != nil {
		return 0, err
	}

	err = os.WriteFile(filepath.Join(r.RevertTempDir, revertStepsBsonName), outBson, os.ModePerm)
	if err != nil {
		return 0, err
	}

	serializedRevertStepsDir, err := GetSerializedRevertStepsDir()
	if err != nil {
		return 0, err
	}

	dirEntries, err := os.ReadDir(serializedRevertStepsDir)
	if err != nil {
		return 0, err
	}

	largestRevertNum := -1
	for _, entry := range dirEntries {
		entryName := strings.TrimSuffix(entry.Name(), ".tar.gz")
		num, err := strconv.Atoi(entryName)
		if err != nil {
			continue
		}

		if num > largestRevertNum {
			largestRevertNum = num
		}
	}

	revertNum := largestRevertNum + 1

	err = targz.Compress(
		filepath.Join(r.RevertTempDir, "*"),
		filepath.Join(serializedRevertStepsDir, fmt.Sprintf("%d.tar.gz", revertNum)),
	)

	return revertNum, err
}

func (r *RevertSteps) Deserialize(revertNum int) error {
	// Firstly extract .tar.gz to r.RevertTempDir

	if err := r.DeleteRuntimeTemp(); err != nil {
		return err
	}
	if err := os.Mkdir(r.RevertTempDir, os.ModePerm); err != nil {
		return err
	}

	revertStepsDir, err := GetSerializedRevertStepsDir()
	if err != nil {
		return err
	}

	revertNumDir := filepath.Join(r.RevertTempDir, strconv.Itoa(revertNum))
	revertTarGzPath := filepath.Join(revertStepsDir, fmt.Sprintf("%d.tar.gz", revertNum))

	if err = targz.Extract(revertTarGzPath, revertNumDir); err != nil {
		return err
	}
	if err := os.Remove(revertTarGzPath); err != nil {
		return err
	}

	err = moveAllChildren(revertNumDir, filepath.Dir(revertNumDir))
	if err != nil {
		return err
	}
	if err := os.Remove(revertNumDir); err != nil {
		return err
	}

	// Unmarshall bson into r

	bsonPath := filepath.Join(r.RevertTempDir, revertStepsBsonName)
	bsonContent, err := os.ReadFile(bsonPath)
	if err != nil {
		return err
	}

	// Save revertTempDir
	revertTempDir := r.RevertTempDir

	if err := bson.Unmarshal(bsonContent, r); err != nil {
		return err
	}
	r.RevertTempDir = revertTempDir

	return os.Remove(bsonPath)
}

func moveAllChildren(currentPath, destPath string) error {
	dirEntries, err := os.ReadDir(currentPath)
	if err != nil {
		return err
	}

	for _, entry := range dirEntries {
		destEntryPath := filepath.Join(destPath, entry.Name())
		currentEntryPath := filepath.Join(currentPath, entry.Name())

		if entry.IsDir() {
			if err := os.Mkdir(currentEntryPath, os.ModePerm); err != nil {
				return err
			}
			if err := moveAllChildren(currentEntryPath, destEntryPath); err != nil {
				return err
			}
			if err := os.Remove(currentEntryPath); err != nil {
				return err
			}
			continue
		}

		if err := os.Rename(currentEntryPath, destEntryPath); err != nil {
			return err
		}
	}

	return nil
}
