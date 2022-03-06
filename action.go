package main

import (
	"errors"
	"github.com/hashicorp/go-multierror"
	"strings"
)

type SSHAction interface {
	Get(remoteFile, localFile string) error
	Put(localFile, remoteFile string) error
	Shell(cmd string) error
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
	var result = new(multierror.Error)
	for _, host := range hosts {
		c, err := newSSHClient(host, username, password, port, timeout)
		if err != nil {
			result = multierror.Append(result, err)
			continue
		}
		s.list = append(s.list, c)
	}
	return result.ErrorOrNil()
}

func (s *simpleSSHAction) Get(remoteFile, localFile string) error {
	var result = new(multierror.Error)
	for _, client := range s.list {
		result = multierror.Append(result, client.getFile(remoteFile, localFile))
		result = multierror.Append(result, client.exit())
	}
	return result.ErrorOrNil()
}

func (s *simpleSSHAction) Put(localFile, remoteFile string) error {
	var result = new(multierror.Error)
	for _, client := range s.list {
		result = multierror.Append(result, client.putFile(localFile, remoteFile))
		result = multierror.Append(result, client.exit())
	}
	return result.ErrorOrNil()
}

func (s *simpleSSHAction) Shell(cmd string) error {
	var result = new(multierror.Error)
	for _, client := range s.list {
		result = multierror.Append(result, client.execShell(cmd))
		result = multierror.Append(result, client.exit())
	}
	return result.ErrorOrNil()
}
