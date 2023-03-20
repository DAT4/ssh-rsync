package rsync

import (
	"fmt"
	"strings"

	"github.com/DAT4/ssh-rsync/ssh/client"
	"github.com/DAT4/ssh-rsync/ssh/session"
	"golang.org/x/crypto/ssh"
)

type Rsync struct {
	sourcepath, targetpath string
	cli                    *client.Client
}

func New(local, remote string, cli *client.Client) *Rsync {
	return &Rsync{
		sourcepath: fmt.Sprintf("%s/", strings.TrimSuffix(local, "/")),
		targetpath: fmt.Sprintf("%s/", strings.TrimSuffix(remote, "/")),
		cli:        cli,
	}
}

func (r *Rsync) targtfile(file string) string {
	return fmt.Sprintf("%s%s", r.targetpath, file)
}

func (r *Rsync) sourcefile(file string) string {
	return fmt.Sprintf("%s%s", r.sourcepath, file)
}

func (r *Rsync) Sync() error {
	return r.cli.Handle(func(c *ssh.Client) error {
		currentState, err := session.New(c).GetState(r.sourcepath, r.targetpath)
		if err != nil {
			return fmt.Errorf("Failed getting current state: %s", err)
		}
		if err := r.syncFiles(currentState.LocalChang, c); err != nil {
			return fmt.Errorf("Failed syncronizing local changes: %s", err)
		}
		if err := r.syncFiles(currentState.RemoteMis, c); err != nil {
			return fmt.Errorf("Failed syncronizing remote missed: %s", err)
		}
		return nil
	})
}

func (r *Rsync) syncFiles(files []string, c *ssh.Client) error {
	for _, file := range files {
		if err := session.New(c).SyncFile(r.sourcefile(file), r.targtfile(file)); err != nil {
			return fmt.Errorf("Failed syncing file %s: %s", file, err)
		}
	}
	return nil
}
