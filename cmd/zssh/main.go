package main

import (
	"fmt"
	"github.com/manifoldco/promptui"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/zacscoding/zssh/pkg/database"
	"github.com/zacscoding/zssh/pkg/host"
	"io"
	"os"
	"path/filepath"
	"time"
)

var workspace string

var (
	commit  = "HEAD"
	version = "latest"
)

var (
	stdin  io.Reader = os.Stdin
	stdout io.Writer = os.Stdout
	stderr io.Writer = os.Stderr
)

var hostStore host.Store

var rootCmd = &cobra.Command{
	Use:     "zssh",
	Short:   "SSH Command line utilities :)",
	Version: fmt.Sprintf("%s (%s)", version, commit),
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		stdin = cmd.InOrStdin()
		stdout = cmd.OutOrStdout()
		stderr = cmd.OutOrStderr()
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: stdout, TimeFormat: time.RFC3339})
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&workspace, "workspace", "", "workspace path(default: $HOME/.zssh)")
	cobra.OnInitialize(onInitialize)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

func onInitialize() {
	if workspace == "" {
		p, err := getDefaultWorkspace()
		if err != nil {
			panic(err)
		}
		workspace = p
	}
	if err := checkWorkspace(workspace); err != nil {
		panic(err)
	}

	db, err := database.NewSQLiteDB(filepath.Join(workspace, "zssh.db"))
	if err != nil {
		panic(err)
	}
	if err := db.Migrator().AutoMigrate(new(host.ServerInfo), new(host.ActiveServerInfo)); err != nil {
		panic(err)
	}
	hostStore = host.NewStore(db)
}

func checkWorkspace(configPath string) error {
	stat, err := os.Stat(configPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}

		return os.MkdirAll(configPath, os.ModePerm)
	}
	if !stat.IsDir() {
		return fmt.Errorf("config path(%s) is not a directory", configPath)
	}
	return nil
}

func getDefaultWorkspace() (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".zssh"), nil
}

func confirmPrompt(label string) (bool, error) {
	prompt := promptui.Prompt{
		Label: label + " [yN]",
		Validate: func(input string) error {
			switch input {
			case "y", "N":
				return nil
			}
			return errors.New("Enter [yN]")
		},
	}

	result, err := prompt.Run()
	if err != nil {
		return false, err
	}
	return result == "y", nil
}

func isUserCancelError(err error) bool {
	switch err {
	case promptui.ErrEOF, promptui.ErrAbort, promptui.ErrInterrupt:
		return true
	default:
		return false
	}
}
