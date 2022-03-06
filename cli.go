package main

import (
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	appName    = "sshctl"
	appVersion = "v0.1.0"

	remoteHost = "remote"
	username   = "username"
	password   = "password"
	timeout    = "timeout"
	port       = "port"
)

func init() {
	rootCmd.PersistentFlags().StringP(username, "u", "root", "ssh username")
	rootCmd.PersistentFlags().StringP(password, "p", "", "ssh password")
	rootCmd.PersistentFlags().IntP(timeout, "t", 0, "Task timeout (in seconds)")
	rootCmd.PersistentFlags().StringSliceP(remoteHost, "r", []string{}, "ssh server remote addr")
	rootCmd.PersistentFlags().Int(port, 22, "ssh server port")
	_ = rootCmd.MarkPersistentFlagRequired(remoteHost)
	_ = rootCmd.MarkPersistentFlagRequired(password)

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(getFileCmd)
	rootCmd.AddCommand(putFileCmd)
	rootCmd.AddCommand(execShellCmd)
}

func execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   appName,
	Short: appName + " is an SSH helper",
	Long:  appName + ` can upload files to the remote host or download files from the remote host locally,or execute script commands on the remote host without interaction with users`,
}

var versionCmd = &cobra.Command{
	Use:     "version",
	Aliases: []string{"v", "ver"},
	Short:   "Print the version number of " + appName,
	Long:    fmt.Sprintf(`All software has versions. This is %s's`, appName),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s version %s %s/%s", appName, appVersion, runtime.GOOS, runtime.GOARCH)
	},
}

var getFileCmd = &cobra.Command{
	Use:     "get",
	Aliases: []string{"download", "down"},
	Short:   "Download files from remote host to local",
	Long:    `Usage: get remoteFile localFile`,
	Args:    cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		t := time.Now()
		defer func(t time.Time) { fmt.Printf("It takes %s\n", time.Since(t)) }(t)

		op, err := getSftpClient(cmd)
		if err != nil {
			fmt.Println(err)
			return
		}

		err = op.Get(args[0], args[1])
		if err != nil {
			fmt.Println(err)
		}
	},
}

var putFileCmd = &cobra.Command{
	Use:     "put",
	Aliases: []string{"upload", "up"},
	Short:   "Last local file to remote host",
	Long:    `Usage: put localFile remoteFile`,
	Args:    cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		t := time.Now()
		defer func(t time.Time) { fmt.Printf("It takes %s\n", time.Since(t)) }(t)

		op, err := getSftpClient(cmd)
		if err != nil {
			fmt.Println(err)
			return
		}

		err = op.Put(args[0], args[1])
		if err != nil {
			fmt.Println(err)
		}
	},
}

var execShellCmd = &cobra.Command{
	Use:     "sh",
	Aliases: []string{"exec", "shell", "bash"},
	Short:   "Put the script on the remote host for execution",
	Long:    `Usage: sh script`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		op, err := getSftpClient(cmd)
		if err != nil {
			fmt.Println(err)
			return
		}

		err = op.Shell(args[0])
		if err != nil {
			fmt.Println(err)
		}
	},
}

func getSftpClient(cmd *cobra.Command) (SSHAction, error) {
	var result = new(multierror.Error)

	sshPort, err := strconv.Atoi(cmd.Flag(port).Value.String())
	result = multierror.Append(result, err)
	taskTimeout, err := strconv.Atoi(cmd.Flag(timeout).Value.String())
	result = multierror.Append(result, err)

	if err := result.ErrorOrNil(); err != nil {
		return nil, fmt.Errorf("getSftpClient-%s->%w", timeout, err)
	}

	str := cmd.Flag(remoteHost).Value.String()
	hosts := strings.Split(str[1:len(str)-1], ",")

	return NewSSHAction(
		cmd.Flag(username).Value.String(),
		cmd.Flag(password).Value.String(),
		hosts,
		sshPort,
		taskTimeout,
	)
}
