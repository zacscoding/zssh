package main

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/zacscoding/zssh/pkg/host"
	"github.com/zacscoding/zssh/pkg/ssh"
	"gorm.io/gorm"
	"os"
)

var (
	ErrNoActiveHost = errors.New("no active host")
)

var (
	sshHostName string
)

func init() {
	sshShellCmd.PersistentFlags().StringVarP(&hostName, "name", "n", "", "the host name of identifier")
	sshExecCmd.PersistentFlags().StringVarP(&hostName, "name", "n", "", "the host name of identifier")

	sshCmd.AddCommand(sshShellCmd, sshExecCmd)
	rootCmd.AddCommand(sshCmd)
}

var sshCmd = &cobra.Command{
	Use:   "ssh",
	Short: "Commands for ssh",
}

var sshShellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Open the remote shell",
	RunE: func(cmd *cobra.Command, args []string) error {
		info, err := getServerInfoOrActive(sshHostName)
		if err != nil {
			return errors.Wrapf(err, "find the host(%s)", sshHostName)
		}

		cli, err := ssh.NewClient(&ssh.ClientParams{
			ServerInfo: info,
			StdIn:      os.Stdin,
			Stdout:     os.Stdout,
			Stderr:     os.Stderr,
		})
		if err != nil {
			return errors.Wrap(err, "create the ssh client")
		}
		if err := cli.OpenShell(); err != nil {
			return errors.Wrap(err, "open the shell")
		}
		log.Info().Msg("😎 Good bye")
		return nil
	},
}

var sshExecCmd = &cobra.Command{
	Use:   "exec",
	Short: "Execute command to the remote host",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		info, err := getServerInfoOrActive(sshHostName)
		if err != nil {
			return errors.Wrapf(err, "find the host(%s)", sshHostName)
		}

		cli, err := ssh.NewClient(&ssh.ClientParams{
			ServerInfo: info,
			StdIn:      stdin,
			Stdout:     stdout,
			Stderr:     stderr,
		})
		if err != nil {
			return errors.Wrap(err, "create the ssh client")
		}

		log.Info().Msgf("⚡ %s: %s", cli.ServerInfo.String(), args[0])
		if err := cli.Run(args[0]); err != nil {
			return errors.Wrap(err, "execute the command")
		}
		return nil
	},
}

func getServerInfoOrActive(hostname string) (*host.ServerInfo, error) {
	if hostname == "" {
		info, err := hostStore.FindActiveServerInfo(context.Background())
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, ErrNoActiveHost
			}
			return nil, err
		}
		return info, nil
	}

	info, err := hostStore.FindByName(context.Background(), hostname)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("host(%s) not found", hostname)
		}
		return nil, err
	}
	return info, nil
}
