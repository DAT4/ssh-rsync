package rsync

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sshcli/client"
	"sshcli/files"
	"sshcli/state"
	"strings"

	"golang.org/x/crypto/ssh"
)

type Rsync struct {
	LocalDir, RemoteDir string
	cli                 *client.Client
}

func New(local, remote string, cli *client.Client) *Rsync {
	return &Rsync{
		LocalDir:  fmt.Sprintf("%s/", strings.TrimSuffix(local, "/")),
		RemoteDir: fmt.Sprintf("%s/", strings.TrimSuffix(remote, "/")),
		cli:       cli,
	}
}

func (r *Rsync) rpath(file string) string {
	return fmt.Sprintf("%s%s", r.RemoteDir, file)
}

func (r *Rsync) lpath(file string) string {
	return fmt.Sprintf("%s%s", r.LocalDir, file)
}

func (r *Rsync) Sync() error {
	var st state.State
	err := r.cli.Do(func(s *ssh.Session) error {
		localModTimes, err := files.WalkDir(r.LocalDir)
		if err != nil {
			return fmt.Errorf("Failed getting local filemodtimes: %s\n", err)
		}
		remoteModTimes, err := files.GetRemoteFileModTimes(s, r.RemoteDir)
		if err != nil {
			return fmt.Errorf("Failed getting remote filemodtimes: %s\n", err)
		}
		st = state.New(localModTimes, remoteModTimes)
		st.Print()

		return nil
	})
	if err != nil {
		return err
	}
	return r.cli.Map(func(c *ssh.Client) error {
		if err := r.syncFiles(st.LocalChang, c); err != nil {
			return err
		}
		if err := r.syncFiles(st.RemoteMis, c); err != nil {
			return err
		}
		return nil
	})
}

func (r *Rsync) syncFiles(files []string, c *ssh.Client) error {
	for _, file := range files {

		if err := r.syncFile(file, c); err != nil {
			return fmt.Errorf("Failed syncing file %s: %s", file, err)
		}

		if err := r.overWriteTimeStamp(file, c); err != nil {
			return fmt.Errorf("Failed overriding timestamp on file %s: %s", file, err)
		}

	}

	return nil
}

func (r *Rsync) syncFile(file string, c *ssh.Client) error {
	fmt.Printf("transfering file: %s\n", file)
	s, err := c.NewSession()
	if err != nil {
		return fmt.Errorf("Failed to start new ssh session: %s", err)
	}

	defer s.Close()

	sourcePath := r.lpath(file)
	targetPath := r.rpath(file)

	fd, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("Failed opening local file: %s", err)
	}
	defer fd.Close()

	targetDir := filepath.Dir(targetPath)
	cmd := fmt.Sprintf("mkdir -p %s && cat > %s", targetDir, targetPath)

	stdin, err := s.StdinPipe()
	if err != nil {
		return fmt.Errorf("Failed getting stdinPipe from session: %s", err)
	}
	defer stdin.Close()

	stdout, err := s.StdoutPipe()
	if err != nil {
		return fmt.Errorf("Failed getting stdoutPipe from session: %s", err)
	}

	if err := s.Start(cmd); err != nil {
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

	if err := s.Wait(); err != nil {
		return fmt.Errorf("Failed to wait for command to exit: %s", err)
	}

	if len(output) > 0 {
		fmt.Printf("Remote output: %s\n", output)
	}
	return nil
}

func (r *Rsync) overWriteTimeStamp(file string, c *ssh.Client) error {

	s, err := c.NewSession()
	if err != nil {
		return fmt.Errorf("Failed to start new ssh session: %s", err)
	}

	defer s.Close()

	sourcePath := r.lpath(file)
	targetPath := r.rpath(file)

	info, err := os.Stat(sourcePath)
	if err != nil {
		return fmt.Errorf("Failed retreiving info from source file: %s", err)
	}

	ts := strings.Split(info.ModTime().String(), " ")
	timeStr := fmt.Sprintf("%s %s %s", ts[0], ts[1], ts[2])

	cmd := fmt.Sprintf("touch -d '%s' %s", timeStr, targetPath)

	fmt.Println(cmd)

	out, err := s.CombinedOutput(cmd)
	if err != nil {
		return fmt.Errorf("Failed to run touch command: %s", err)
	}

	fmt.Println(out)

	return nil
}
