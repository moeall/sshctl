package main

import (
	"bytes"
	"fmt"
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
		return fmt.Errorf("os.Open->%s %w", localFile, err)
	}
	defer func() { _ = srcFile.Close() }()

	pwd, _ := c.sftp.Getwd()
	dir := filepath.Dir(filepath.Join(pwd, remoteFile))
	err = c.sftp.MkdirAll(strings.ReplaceAll(dir, string(os.PathSeparator), "/"))
	if err != nil {
		return fmt.Errorf("sftp.MkdirAll->%s %w", dir, err)
	}

	dstFile, err := c.sftp.Create(remoteFile)
	if err != nil {
		return fmt.Errorf("sftp.Create->%s %w", remoteFile, err)
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
		return fmt.Errorf("sftp.Open->%s %w", remoteFile, err)
	}
	defer func() { _ = srcFile.Close() }()

	pwd, _ := os.Getwd()
	dir := filepath.Dir(filepath.Join(pwd, localFile))
	err = os.MkdirAll(dir, 0666)
	if err != nil {
		return fmt.Errorf("os.MkdirAll->%s %w", dir, err)
	}

	dstFile, err := os.Create(localFile)
	if err != nil {
		return fmt.Errorf("os.Create->%s %w", localFile, err)
	}
	defer func() { _ = dstFile.Close() }()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("io.Copy->%w", err)
	}
	return nil
}

func (c *sshClient) execShell(script []byte) error {
	session, err := c.ssh.NewSession()
	if err != nil {
		return fmt.Errorf("ssh.NewSession->%w", err)
	}
	defer func() { _ = session.Close() }()

	session.Stdin = bytes.NewReader(script)
	session.Stdout = os.Stdout
	err = session.Shell()
	if err != nil {
		return fmt.Errorf("session.Shell->%w", err)
	}
	return session.Wait()
}

func (c *sshClient) exit() error {
	_ = c.sftp.Close()
	_ = c.ssh.Close()
	return nil
}
