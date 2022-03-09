package main

import (
	"errors"
	"fmt"
	"github.com/hashicorp/go-multierror"
	"strings"
	"time"
)

type SSHAction interface {
	Get(remoteFile, localFile string) error
	Put(localFile, remoteFile string) error
	Shell(script []byte) error
}

func NewSSHAction(username, password string, hosts []string, port, timeout int) (SSHAction, error) {
	switch {
	case `` == strings.TrimSpace(username),
		`` == strings.TrimSpace(password),
		0 == len(hosts),
		0 >= port || port > 65535:
		return nil, errors.New("invalid parameters")
	}

	client := &simpleSSHAction{}
	return client, client.init(username, password, hosts, port, timeout)
}

type simpleSSHAction struct {
	list []*sshClient
}

func (s *simpleSSHAction) init(username, password string, hosts []string, port, timeout int) error {
	var g multierror.Group
	for _, addr := range hosts {
		host := addr
		g.Go(func() error {
			c, err := newSSHClient(host, username, password, port, timeout)
			if err == nil {
				s.list = append(s.list, c)
			}
			return err
		})
	}
	return g.Wait().ErrorOrNil()
}

func (s *simpleSSHAction) Get(remoteFile, localFile string) error {
	var g multierror.Group
	for _, c := range s.list {
		client := c
		g.Go(func() error {
			t := time.Now()
			defer func(t time.Time) {
				fmt.Printf("Get file %s form host %s takes %s\n", remoteFile, client.ssh.RemoteAddr().String(), time.Since(t))
			}(t)

			err := client.getFile(remoteFile, localFile)
			if err == nil {
				err = client.exit()
			}
			return err
		})
	}
	return g.Wait().ErrorOrNil()
}

func (s *simpleSSHAction) Put(localFile, remoteFile string) error {
	var g multierror.Group
	for _, c := range s.list {
		client := c
		g.Go(func() error {
			t := time.Now()
			defer func(t time.Time) {
				fmt.Printf("Put file %s to host %s takes %s\n", localFile, client.ssh.RemoteAddr().String(), time.Since(t))
			}(t)

			err := client.putFile(localFile, remoteFile)
			if err == nil {
				err = client.exit()
			}
			return err
		})
	}
	return g.Wait().ErrorOrNil()
}

func (s *simpleSSHAction) Shell(script []byte) error {
	var es = new(multierror.Error)
	for _, client := range s.list {
		es = multierror.Append(es, client.execShell(script))
		es = multierror.Append(es, client.exit())
	}
	return es.ErrorOrNil()
}
