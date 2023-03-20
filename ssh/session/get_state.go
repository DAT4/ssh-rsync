package session

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"github.com/DAT4/ssh-rsync/state"
)

func (s *Session) GetState(sourcePath, targetPath string) (*state.State, error) {
	localModTimes, err := walkDir(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("Failed getting local filemodtimes: %s\n", err)
	}
	remoteModTimes, err := s.walkTargetDir(targetPath)
	if err != nil {
		return nil, fmt.Errorf("Failed getting remote filemodtimes: %s\n", err)
	}
	st := state.New(localModTimes, remoteModTimes)

	return st, nil
}

func walkDir(root string) (map[string]time.Time, error) {
	modTimes := make(map[string]time.Time)
	err := filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		modTimes[strings.TrimPrefix(path, root)] = info.ModTime()
		return nil
	})
	return modTimes, err
}

func (s *Session) walkTargetDir(root string) (map[string]time.Time, error) {
	session, err := s.cli.NewSession()
	if err != nil {
		return nil, fmt.Errorf("Failed creating new session: %s", err)
	}
	cmd := fmt.Sprintf("find %s -type f -printf '%%p++%%TY-%%Tm-%%Td %%TH:%%TM:%%TS %%TZ %%Tz\n'", root)
	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return nil, err
	}
	modTimes := make(map[string]time.Time)
	for _, line := range strings.Split(string(output), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		name, timestamp, err := getNameAndTime(line, root)
		if err != nil {
			return nil, fmt.Errorf("Failed while getting name and time from file: %s", err)
		}
		modTimes[name] = timestamp
	}
	return modTimes, nil
}

func getNameAndTime(line, root string) (string, time.Time, error) {
	parts := strings.Split(line, "++")
	if len(parts) != 2 {
		return "", time.Now(), fmt.Errorf("unexpected output from remote find command: %s\n", line)
	}
	layout := "2006-01-02 15:04:05 MST -0700"
	timestamp, err := time.Parse(layout, parts[1])
	if err != nil {
		return "", time.Now(), fmt.Errorf("Failed parsing time: %s", err)
	}
	return strings.TrimPrefix(parts[0], root), timestamp, nil
}
