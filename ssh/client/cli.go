package client

import (
	"fmt"

	"golang.org/x/crypto/ssh"
)

type Client struct {
	host   string
	config *ssh.ClientConfig
}

func New(username, host string, privateKey []byte) (*Client, error) {
	signer, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse private key: %s", err.Error())
	}
	return &Client{
		host: host,
		config: &ssh.ClientConfig{
			User: username,
			Auth: []ssh.AuthMethod{
				ssh.PublicKeys(signer),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		},
	}, nil
}

func (c *Client) Do(handler func(*ssh.Session) error) error {
	client, err := ssh.Dial("tcp", c.host, c.config)
	if err != nil {
		return fmt.Errorf("Failed to dial: %s", err)
	}
	defer client.Close()
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("Failed to create session: %s", err)
	}
	defer session.Close()
	if err := handler(session); err != nil {
		return fmt.Errorf("Failed to apply handler on session: %s", err)
	}
	return nil
}

func (c *Client) Handle(handler func(*ssh.Client) error) error {
	client, err := ssh.Dial("tcp", c.host, c.config)
	if err != nil {
		return fmt.Errorf("Failed to dial: %s", err)
	}
	defer client.Close()
	if err := handler(client); err != nil {
		return fmt.Errorf("Failed to apply handler on client: %s", err)
	}
	return nil
}
