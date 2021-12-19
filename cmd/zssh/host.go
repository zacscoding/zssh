package main

import (
	"context"
	"fmt"
	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/zacscoding/zssh/pkg/host"
	"gorm.io/gorm"
	"os"
	"strconv"
)

const (
	defaultHostPort = 22
)

var (
	hostName        string
	hostUser        string
	hostAddress     string
	hostPort        = defaultHostPort
	hostPassword    string
	hostKeyPath     string
	hostDescription string
)

func init() {
	hostGetCmd.PersistentFlags().StringVarP(&hostName, "name", "n", "", "the host name of identifier")
	hostDeleteCmd.PersistentFlags().StringVarP(&hostName, "name", "n", "", "the host name of identifier")

	hostCmd.AddCommand(hostAddCmd, hostSelectCmd, hostActiveCmd, hostGetCmd, hostGetsCmd, hostUpdateCmd, hostDeleteCmd)
	rootCmd.AddCommand(hostCmd)
}

var hostCmd = &cobra.Command{
	Use:   "host",
	Short: "Handle host info",
}

var hostAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Adds a new host info",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := readHostPrompt(); err != nil {
			if isUserCancelError(err) {
				log.Info().Msg("😎 Good bye")
				return nil
			}
			return errors.Wrap(err, "read server info")
		}
		h := host.ServerInfo{
			Name:        hostName,
			User:        hostUser,
			Address:     hostAddress,
			Port:        hostPort,
			Password:    hostPassword,
			KeyPath:     hostKeyPath,
			Description: hostDescription,
		}
		if err := hostStore.Save(context.Background(), &h); err != nil {
			return errors.Wrap(err, "save the host")
		}
		log.Info().Msgf("✅ success to add a host\n%s", h.ToJSON(true))
		return nil
	},
}

var hostSelectCmd = &cobra.Command{
	Use:   "select",
	Short: "Select a default host",
	RunE: func(cmd *cobra.Command, args []string) error {
		info, err := selectHostPrompt()
		if err != nil {
			if isUserCancelError(err) {
				log.Info().Msg("😎 Good bye")
				return nil
			}
			return errors.Wrap(err, "select the host")
		}

		if err := hostStore.SaveOrUpdateActiveServerInfo(context.Background(), info); err != nil {
			return errors.Wrap(err, "save or update active host")
		}
		return nil
	},
}

var hostActiveCmd = &cobra.Command{
	Use:   "active",
	Short: "Get active a host",
	RunE: func(cmd *cobra.Command, args []string) error {
		info, err := hostStore.FindActiveServerInfo(context.Background())
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				log.Info().Msgf("Could not find active host 🤔. Activate host with 'zssh host select' command.")
				return nil
			}
			return errors.Wrap(err, "find the activated host")
		}
		log.Info().Msgf("✅ Active host: %s", info.String())
		return nil
	},
}

var hostGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a host",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		hostname := args[0]
		info, err := hostStore.FindByName(context.Background(), hostname)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return errors.Wrapf(err, "host(%s) not find", hostname)
			}
			return errors.Wrapf(err, "find the host(%s)", hostname)
		}
		log.Info().Msg(info.ToJSON(true))
		return nil
	},
}

var hostGetsCmd = &cobra.Command{
	Use:   "gets",
	Short: "Get host all",
	RunE: func(cmd *cobra.Command, args []string) error {
		hosts, err := hostStore.FindAll(context.Background())
		if err != nil {
			return errors.Wrap(err, "find all hosts")
		}
		log.Info().Msgf("⚡ Total hosts: #%d", len(hosts))
		for _, info := range hosts {
			log.Info().Msgf("  🔹 %s", info.ToJSON(false))
		}
		return nil
	},
}

var hostUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the host",
	RunE: func(cmd *cobra.Command, args []string) error {
		info, err := selectHostPrompt()
		if err != nil {
			if isUserCancelError(err) {
				log.Info().Msg("😎 Good bye")
				return nil
			}
			errors.Wrap(err, "select the host")
		}

		hostName = info.Name
		hostUser = info.User
		hostAddress = info.Address
		hostPort = info.Port
		hostPassword = info.Password
		hostKeyPath = info.KeyPath
		hostDescription = info.Description

		if err := readHostPrompt(); err != nil {
			if isUserCancelError(err) {
				log.Info().Msg("😎 Good bye")
				return nil
			}
			return errors.Wrap(err, "read server info")
		}
		update := host.ServerInfo{
			ID:          info.ID,
			Name:        hostName,
			User:        hostUser,
			Address:     hostAddress,
			Port:        hostPort,
			Password:    hostPassword,
			KeyPath:     hostKeyPath,
			Description: hostDescription,
		}
		if _, err := hostStore.Update(context.Background(), &update); err != nil {
			return errors.Wrap(err, "update the host")
		}
		log.Info().Msgf("✅ success to update the host\n%s", update.ToJSON(true))
		return nil
	},
}

var hostDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete the host",
	RunE: func(cmd *cobra.Command, args []string) error {
		info, err := selectHostPrompt()
		if err != nil {
			if isUserCancelError(err) {
				log.Info().Msg("😎 Good bye")
				return nil
			}
			errors.Wrap(err, "select the host")
		}

		ok, err := confirmPrompt(fmt.Sprintf("remove %s?", info.String()))
		if err != nil {
			if isUserCancelError(err) {
				log.Info().Msg("😎 Good bye")
				return nil
			}
			errors.Wrap(err, "confirm to delete")
		}
		if !ok {
			log.Info().Msgf("Cancel to delete the host(%s)", info.String())
			return nil
		}

		deleted, err := hostStore.DeleteByName(context.Background(), info.Name)
		if err != nil {
			return errors.Wrapf(err, "delete the host(%s)", hostName)
		}
		if deleted == 0 {
			return errors.Wrapf(err, "host(%s) not found", hostName)
		}
		log.Info().Msgf("Success to delete the host(%s)", hostName)
		return nil
	},
}

func selectHostPrompt() (*host.ServerInfo, error) {
	ctx := context.Background()
	hosts, err := hostStore.FindAll(ctx)
	if err != nil {
		log.Error().Msgf("failed to find hosts. reason: %v", err)
		os.Exit(1)
	}
	if len(hosts) == 0 {
		log.Error().Msg("empty hosts")
		os.Exit(1)
	}

	var hostNames []string
	for _, info := range hosts {
		hostNames = append(hostNames, info.Name)
	}

	p := promptui.Select{
		Label: "Select host",
		Items: hostNames,
		Size:  10,
	}

	_, selected, err := p.Run()
	if err != nil {
		os.Exit(1)
	}

	var (
		selectedHostName = selected
		selectedHost     *host.ServerInfo
	)

	for _, info := range hosts {
		if info.Name == selectedHostName {
			selectedHost = info
			break
		}
	}
	return selectedHost, nil
}

func readHostPrompt() error {
	inputs := []struct {
		label  string
		valueP interface{}
		mask   rune
	}{
		{label: "name", valueP: &hostName},
		{label: "user", valueP: &hostUser},
		{label: "address", valueP: &hostAddress},
		{label: "port", valueP: &hostPort},
		{label: "password", valueP: &hostPassword, mask: '*'},
		{label: "keypath", valueP: &hostKeyPath},
		{label: "description", valueP: &hostDescription},
	}

	for _, input := range inputs {
		switch p := input.valueP.(type) {
		case *string:
			prompt := promptui.Prompt{
				Label:   input.label,
				Mask:    input.mask,
				Default: *p,
			}
			result, err := prompt.Run()
			if err != nil {
				return err
			}
			*p = result
		case *int:
			prompt := promptui.Prompt{
				Label:   input.label,
				Mask:    input.mask,
				Default: strconv.Itoa(*p),
				Validate: func(input string) error {
					_, err := strconv.ParseInt(input, 10, 64)
					if err != nil {
						return errors.New("invalid number")
					}
					return nil
				},
			}

			result, err := prompt.Run()
			if err != nil {
				return err
			}
			v, _ := strconv.ParseInt(result, 10, 32)
			*p = int(v)
		}
	}
	return nil
}
