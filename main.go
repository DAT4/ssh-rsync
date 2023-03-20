package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/DAT4/ssh-rsync/rsync"
	"github.com/DAT4/ssh-rsync/ssh/client"
)

func main() {
	key, err := ioutil.ReadFile("/home/martin/.ssh/id_ed25519")
	if err != nil {
		panic("Failed to read private key: " + err.Error())
	}
	client, err := client.New("martin", "docker.lan:22", key)
	if err != nil {
		panic(err)
	}

	r := rsync.New("/home/martin/rsyncTestPath", "/home/martin/rsyncTestPath", client)

	if err := r.Sync(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("success")
}
