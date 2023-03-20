package state

import (
	"fmt"
	"time"
)

type State struct {
	RemoteMis, LocalMis, RemoteChange, LocalChang []string
}

func New(a, b map[string]time.Time) *State {
	missing, changed := getChanges(a, b)
	srvMiss, srvChang := getChanges(b, a)

	return &State{
		RemoteMis:    missing,
		LocalChang:   changed,
		LocalMis:     srvMiss,
		RemoteChange: srvChang,
	}
}

func (state *State) Print() {
	fmt.Println("Files missing on remote host:")
	for _, file := range state.RemoteMis {
		fmt.Println(file)
	}
	fmt.Println("Files changed on local host:")
	for _, file := range state.LocalChang {
		fmt.Println(file)
	}
	fmt.Println("Files missing on local host:")
	for _, file := range state.LocalMis {
		fmt.Println(file)
	}
	fmt.Println("Files changed on remote host:")
	for _, file := range state.RemoteChange {
		fmt.Println(file)
	}
}

func getChanges(local, remote map[string]time.Time) (missing, changed []string) {
	for file, modTime := range local {
		remoteModTime, ok := remote[file]
		if !ok {
			missing = append(missing, file)
		} else if modTime.After(remoteModTime) {
			changed = append(changed, file)
		}
	}
	return
}
