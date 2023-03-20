package files

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

func WalkDir(root string) (map[string]time.Time, error) {
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

func GetRemoteFileModTimes(session *ssh.Session, root string) (map[string]time.Time, error) {
	format := "%%p++%%TY-%%Tm-%%Td %%TH:%%TM:%%TS %%TZ %%Tz"
	cmd := fmt.Sprintf("find %s -type f -printf '"+format+"\n'", root)
	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return nil, err
	}
	modTimes := make(map[string]time.Time)
	for _, line := range strings.Split(string(output), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Split(line, "++")
		if len(parts) != 2 {
			return nil, fmt.Errorf("unexpected output from remote find command: %s\n", line)
		}
		layout := "2006-01-02 15:04:05 MST -0700"
		tim, err := time.Parse(layout, parts[1])
		if err != nil {
			return nil, fmt.Errorf("Failed parsing time: %s", err)
		}
		modTimes[strings.TrimPrefix(parts[0], root)] = tim
	}
	return modTimes, nil
}
