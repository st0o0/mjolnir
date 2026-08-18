package config

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
)

func FixOwnership(runDir string) error {
	u, err := user.Lookup("nut")
	if err != nil {
		return fmt.Errorf("lookup nut user: %w", err)
	}

	uid, err := strconv.Atoi(u.Uid)
	if err != nil {
		return fmt.Errorf("parse nut uid: %w", err)
	}
	gid, err := strconv.Atoi(u.Gid)
	if err != nil {
		return fmt.Errorf("parse nut gid: %w", err)
	}

	return filepath.WalkDir(runDir, func(path string, _ os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		return os.Chown(path, uid, gid)
	})
}
