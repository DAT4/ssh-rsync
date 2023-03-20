package session

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"golang.org/x/crypto/ssh"
)

type Session struct {
	cli *ssh.Client
}

func New(cli *ssh.Client) *Session {
	return &Session{
		cli: cli,
	}
}

func (s *Session) SyncFile(sourcePath, targetPath string) error {

	fd, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("Failed opening local file: %s", err)
	}
	defer fd.Close()

	if err := s.writeFile(fd, targetPath); err != nil {
		return fmt.Errorf("Failed writing file to ssh host: %s", err)
	}

	return s.overWriteTimeStamp(sourcePath, targetPath)
}

func (s *Session) writeFile(fd io.Reader, targetPath string) error {
	session, err := s.cli.NewSession()
	if err != nil {
		return fmt.Errorf("Failed getting new session from ssh client: %s", err)
	}
	defer session.Close()
	targetDir := filepath.Dir(targetPath)
	cmd := fmt.Sprintf("mkdir -p %s && cat > %s", targetDir, targetPath)

	stdin, err := session.StdinPipe()
	if err != nil {
		return fmt.Errorf("Failed getting stdinPipe from session: %s", err)
	}
	defer stdin.Close()

	stdout, err := session.StdoutPipe()
	if err != nil {
		return fmt.Errorf("Failed getting stdoutPipe from session: %s", err)
	}

	if err := session.Start(cmd); err != nil {
		msg := "Failed to start command (%s) on remote host: %%s"
		return fmt.Errorf(fmt.Sprintf(msg, cmd), err)
	}

	if _, err := io.Copy(stdin, fd); err != nil {
		return fmt.Errorf("Failed to copy file to stdin: %s", err)
	}

	if err := stdin.Close(); err != nil {
		return fmt.Errorf("Failed to close stdin: %s", err)
	}

	output, err := ioutil.ReadAll(stdout)
	if err != nil {
		return fmt.Errorf("Failed to read output: %s", err)
	}

	if err := session.Wait(); err != nil {
		return fmt.Errorf("Failed to wait for command to exit: %s", err)
	}

	if len(output) > 0 {
		fmt.Printf("Remote output: %s\n", output)
	}
	return nil
}

func (s *Session) overWriteTimeStamp(sourcePath, targetPath string) error {
	session, err := s.cli.NewSession()
	if err != nil {
		return fmt.Errorf("Failed getting new session from ssh client: %s", err)
	}

	defer session.Close()
	info, err := os.Stat(sourcePath)
	if err != nil {
		return fmt.Errorf("Failed retreiving info from source file: %s", err)
	}

	ts := info.ModTime().Format("2006-01-02 15:04:05.0000000000000")
	cmd := fmt.Sprintf("touch -d '%s' %s", ts, targetPath)

	_, err = session.CombinedOutput(cmd)
	if err != nil {
		return fmt.Errorf("Failed to run touch command: %s", err)
	}
	return nil
}
