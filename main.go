package main

import (
	_ "embed"
	"errors"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	gochefcodegenutils "github.com/ihatiko/go-chef-code-gen-utils"
	"github.com/ihatiko/go-chef/tui"
	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"
	"golang.org/x/mod/semver"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"
)

const configDir = ".go-chef"
const configFile = "go-chef.toml"

type Config struct {
	Proxies        []string  `toml:"proxies"`
	CorePackage    string    `toml:"core-package"`
	CoreCommand    string    `toml:"core-command"`
	BaseCommand    string    `toml:"base-command"`
	BasePackage    string    `toml:"base-package"`
	NextUpdateTime time.Time `toml:"next-update-time"`
	Interval       string    `toml:"interval"`
}

//go:embed config.toml
var defaultConfig []byte

// Step 1 => check config in UserHomeDir
// Step 2 => if it does not exist config => Copy default config
// Step 3 => marshal config

func main() {
	config := getConfig()
	if config == nil {
		return
	}
	params := strings.Join(os.Args[1:], " ")

	// Core command - watch configs, version, and other
	if len(os.Args) > 1 && strings.ToLower(os.Args[1]) == config.BaseCommand {
		setCoreModules()
		return
	}

	updater := gochefcodegenutils.NewUpdater(config.Proxies)

	mainLastVersion, err := updater.GetLastVersion(config.BasePackage)
	if err != nil {
		slog.Error("Error getting last version:", slog.Any("error", err), slog.String("package", config.BasePackage))
	}
	composer := gochefcodegenutils.NewExecutor()
	RunTui(err, mainLastVersion, config, composer)
	build, _ := debug.ReadBuildInfo()
	currentVersion := build.Main.Version
	slog.Info("Current Version:", slog.String("version", currentVersion))

	updater.AutoUpdate(config.CorePackage)

	proxyCommand := fmt.Sprintf("%s %s", config.CoreCommand, params)
	result, err := composer.ExecDefaultCommand(proxyCommand)
	if err != nil {
		slog.Error("Error executing command: ", slog.Any("error", err), slog.String("command", params))
	}
	fmt.Println(result)
}
func RunTui(err error, mainLastVersion string, config *Config, composer *gochefcodegenutils.Executor) {
	build, _ := debug.ReadBuildInfo()
	if err == nil && semver.Compare(build.Main.Version, mainLastVersion) == -1 && time.Now().After(config.NextUpdateTime) {
		installCommand := fmt.Sprintf("go install %s@%s", config.BasePackage, mainLastVersion)
		p := tea.NewProgram(tui.Model{
			Title:   fmt.Sprintf("Available new version %s update now ?", mainLastVersion),
			Choices: []tui.Choice{tui.Yes, tui.Later},
		})

		// Run returns the model as a tea.Model.
		m, err := p.Run()
		if err != nil {
			fmt.Println("Ups please write issue or write to support:", err)
		} else {
			// Assert the final tea.Model to our local model and print the choice.
			if m, ok := m.(tui.Model); ok && m.Choice != "" {
				switch m.Choice {
				case tui.Yes:
					command, err := composer.ExecDefaultCommand(installCommand)
					if err != nil {
						slog.Error("Error executing command: ", slog.Any("error", err))
						return
					}
					fmt.Println(command)
				case tui.Later:
					interval, err := time.ParseDuration(config.Interval)
					if err != nil {
						slog.Error("Error parsing interval:", slog.Any("error", err))
						interval = time.Hour
					}
					config.NextUpdateTime = time.Now().Add(interval)
					configBytes, err := toml.Marshal(config)
					if err != nil {
						slog.Error("Error marshalling config:", slog.Any("error", err))
						return
					}
					dir, err := os.UserHomeDir()
					if err != nil {
						slog.Error("Error getting user cache dir", slog.Any("err", err))
						return
					}
					configPath := filepath.Join(dir, configDir, configFile)
					err = os.WriteFile(configPath, configBytes, fs.ModePerm)
					if err != nil {
						slog.Error("Error writing config file:", slog.Any("err", err))
						return
					}
				}
			}
		}
	}
}
func setCoreModules() {
	dir, err := os.UserHomeDir()
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
	dir, err := os.UserHomeDir()
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
			err := os.Mkdir(dirPath, os.ModePerm)
			if err != nil {
				slog.Error("Error creating config dir", slog.Any("err", err))
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
