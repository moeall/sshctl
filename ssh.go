package main

import (
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type sshClient struct {
	ssh  *ssh.Client
	sftp *sftp.Client
}

func newSSHClient(host, username, password string, port, timeout int) (*sshClient, error) {
	var client = &sshClient{}
	var err error
	client.ssh, err = ssh.Dial(
		"tcp",
		fmt.Sprintf("%s:%d", host, port),
		&ssh.ClientConfig{
			User:            username,
			Auth:            []ssh.AuthMethod{ssh.Password(password)},
			Timeout:         time.Duration(timeout) * time.Second,
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		})
	if err != nil {
		return nil, fmt.Errorf("ssh.Dial->%w", err)
	}

	client.sftp, err = sftp.NewClient(client.ssh)
	if err != nil {
		return nil, fmt.Errorf("sftp.NewClient->%w", err)
	}
	return client, nil
}

func (c *sshClient) putFile(localFile, remoteFile string) error {
	srcFile, err := os.Open(localFile)
	if err != nil {
		return fmt.Errorf("os.Open->%w", err)
	}
	defer func() { _ = srcFile.Close() }()

	err = c.sftp.MkdirAll(filepath.Dir(remoteFile))
	if err != nil {
		return fmt.Errorf("sftp.MkdirAll->%w", err)
	}

	dstFile, err := c.sftp.Create(remoteFile)
	if err != nil {
		return fmt.Errorf("sftp.Create->%w", err)
	}
	defer func() { _ = dstFile.Close() }()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("io.Copy->%w", err)
	}
	return nil

}

func (c *sshClient) getFile(remoteFile, localFile string) error {
	srcFile, err := c.sftp.Open(remoteFile)
	if err != nil {
		return fmt.Errorf("sftp.Open->%w", err)
	}
	defer func() { _ = srcFile.Close() }()

	pwd, _ := os.Getwd()
	err = os.MkdirAll(filepath.Dir(filepath.Join(pwd, localFile)), 0666)
	if err != nil {
		return fmt.Errorf("os.MkdirAll->%w", err)
	}

	dstFile, err := os.Create(localFile)
	if err != nil {
		return fmt.Errorf("os.Create->%w", err)
	}
	defer func() { _ = dstFile.Close() }()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("io.Copy->%w", err)
	}
	return nil
}

func (c *sshClient) execShell(cmd string) error {
	session, err := c.ssh.NewSession()
	if err != nil {
		return fmt.Errorf("ssh.NewSession->%w", err)
	}
	defer func() { _ = session.Close() }()

	session.Stdin = strings.NewReader(cmd)
	session.Stdout = os.Stdout
	err = session.Shell()
	if err != nil {
		return fmt.Errorf("session.Shell->%w", err)
	}

	return session.Wait()
}

func (c *sshClient) exit() error {
	var result = new(multierror.Error)
	result = multierror.Append(result, c.sftp.Close())
	result = multierror.Append(result, c.ssh.Close())

	err := result.ErrorOrNil()
	if result.Unwrap() == io.EOF {
		return nil
	}
	return err
}
