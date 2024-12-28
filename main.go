package main

import (
	_ "embed"
	"errors"
	"fmt"
	gochefcodegenutils "github.com/ihatiko/go-chef-code-gen-utils"
	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
)

const configDir = ".go-chef"
const configFile = "go-chef.toml"

type Config struct {
	Proxies      []string `toml:"proxies"`
	ProxyPackage string   `toml:"proxy-package"`
	ProxyCommand string   `toml:"proxy-command"`
	BaseCommand  string   `toml:"base-command"`
}

//go:embed config.toml
var defaultConfig []byte

// Step 1 => check config in UserConfigDir
// Step 2 => if it does not exist config => Copy default config
// Step 3 => marshal config

func main() {
	config := getConfig()
	if config == nil {
		return
	}
	params := strings.Join(os.Args[1:], " ")

	// Core command - watch configs, version, and another
	if len(os.Args) > 1 && strings.ToLower(os.Args[1]) == config.BaseCommand {
		setCoreModules()
		return
	}

	updater := gochefcodegenutils.NewUpdater(config.Proxies)
	updater.AutoUpdate(config.ProxyPackage)
	composer := gochefcodegenutils.NewExecutor()
	proxyCommand := fmt.Sprintf("%s %s", config.ProxyCommand, params)
	result, err := composer.ExecDefaultCommand(proxyCommand)
	if err != nil {
		slog.Error("Error executing command: ", slog.Any("error", err), slog.String("command", params))
	}
	fmt.Println(result)
}

func setCoreModules() {
	dir, err := os.UserConfigDir()
	if err != nil {
		slog.Error("Error getting user cache dir", slog.Any("err", err))
		return
	}
	configPath := filepath.Join(dir, configDir, configFile)
	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "get current version module",
		Run: func(cmd *cobra.Command, args []string) {
			build, _ := debug.ReadBuildInfo()
			fmt.Println(build.Main.Version)
		},
	}
	var configPathCmd = &cobra.Command{
		Use:   "config-path",
		Short: "get current config path",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(configPath)
		},
	}
	var resetConfigCmd = &cobra.Command{
		Use:   "reset-config",
		Short: "reset current config",
		Run: func(cmd *cobra.Command, args []string) {
			err := os.WriteFile(configPath, defaultConfig, os.ModePerm)
			if err != nil {
				return
			}
			slog.Info("reset config", slog.String("path", configPath))
		},
	}
	var showConfigCmd = &cobra.Command{
		Use:   "show-config",
		Short: "show current config",
		Run: func(cmd *cobra.Command, args []string) {
			bytes, err := os.ReadFile(configPath)
			if err != nil {
				return
			}
			fmt.Println(string(bytes))
		},
	}
	mainCommand := &cobra.Command{}
	mainCommand.AddCommand(versionCmd, configPathCmd, resetConfigCmd, showConfigCmd)
	mainCommand.SetArgs(os.Args[2:])
	if err := mainCommand.Execute(); err != nil {
		slog.Error("Error executing command", slog.Any("error", err))
	}
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func getConfig() *Config {
	dir, err := os.UserConfigDir()
	if err != nil {
		slog.Error("Error getting user cache dir", slog.Any("err", err))
		return nil
	}
	configPath := filepath.Join(dir, configDir, configFile)
	state, err := exists(configPath)
	if err != nil {
		slog.Error("Error checking if config file exists", slog.Any("err", err))
		return nil
	}
	if !state {
		dirPath := filepath.Join(dir, configDir)
		dirState, err := exists(dirPath)
		if err != nil {
			slog.Error("Error checking if config file exists", slog.Any("err", err))
			return nil
		}
		if !dirState {
			slog.Info("Config dir does not exist, creating", slog.Any("dir", dirPath))
			err := os.Mkdir(configDir, os.ModePerm)
			if err != nil {
				return nil
			}
		}
		slog.Info("try create system config file")
		err = os.WriteFile(configPath, defaultConfig, os.ModePerm)
		if err != nil {
			slog.Error("Error creating config file", slog.Any("err", err))
			return nil
		}
		slog.Info("config file created", slog.Any("file", configPath))
	}
	cfgBytes, err := os.ReadFile(configPath)
	if err != nil {
		slog.Error("Error reading config file", slog.Any("err", err), slog.String("file", configPath))
		return nil
	}
	config := new(Config)
	err = toml.Unmarshal(cfgBytes, config)
	if err != nil {
		slog.Error("Error unmarshalling default config", slog.Any("error", err))
		return nil
	}
	return config
}
