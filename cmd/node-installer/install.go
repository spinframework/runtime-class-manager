package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spinframework/runtime-class-manager/internal/containerd"
	"github.com/spinframework/runtime-class-manager/internal/preset"
	"github.com/spinframework/runtime-class-manager/internal/shim"
)

// installCmd represents the install command.
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install containerd shims",
	Run: func(_ *cobra.Command, _ []string) {
		rootFs := afero.NewOsFs()
		hostFs := afero.NewBasePathFs(rootFs, config.Host.RootPath)

		distro, err := DetectDistro(config, hostFs)
		if err != nil {
			slog.Error("failed to detect containerd config", "error", err)
			os.Exit(1)
		}

		config.Runtime.ConfigPath = distro.ConfigPath
		if err = distro.Setup(preset.Env{ConfigPath: distro.ConfigPath, HostFs: hostFs}); err != nil {
			slog.Error("failed to run distro setup", "error", err)
			os.Exit(1)
		}

		config.Runtime.Options, err = RuntimeOptions()
		if err != nil {
			slog.Error("failed to get runtime options", "error", err)
			os.Exit(1)
		}

		if err := RunInstall(config, rootFs, hostFs, distro.Restarter); err != nil {
			slog.Error("failed to install", "error", err)
			os.Exit(1)
		}
	},
}

func init() {
	installCmd.Flags().StringVarP(&config.RCM.AssetPath, "asset-path", "a", "/assets", "Path to the asset to install")
	rootCmd.AddCommand(installCmd)
}

func RunInstall(config Config, rootFs, hostFs afero.Fs, restarter containerd.Restarter) error {
	// Get file or directory information.
	info, err := rootFs.Stat(config.RCM.AssetPath)
	if err != nil {
		return err
	}

	var files []fs.FileInfo
	// Check if the path is a directory.
	if info.IsDir() {
		files, err = afero.ReadDir(rootFs, config.RCM.AssetPath)
		if err != nil {
			return err
		}
	} else {
		// If the path is not a directory, add the file to the list of files.
		files = append(files, info)
		config.RCM.AssetPath = path.Dir(config.RCM.AssetPath)
	}

	containerdConfig := containerd.NewConfig(hostFs, config.Runtime.ConfigPath, restarter, config.Runtime.Options)
	shimConfig := shim.NewConfig(rootFs, hostFs, config.RCM.AssetPath, config.RCM.Path)

	anythingChanged := false
	for _, file := range files {
		fileName := file.Name()
		runtimeName := shim.RuntimeName(fileName)

		binPath, changed, err := shimConfig.Install(fileName)
		if err != nil {
			return fmt.Errorf("failed to install shim '%s': %w", runtimeName, err)
		}
		anythingChanged = anythingChanged || changed
		slog.Info("shim installed", "shim", runtimeName, "path", binPath, "new-version", changed)

		err = containerdConfig.AddRuntime(binPath)
		if err != nil {
			return fmt.Errorf("failed to write containerd config: %w", err)
		}
		slog.Info("shim configured", "shim", runtimeName, "path", config.Runtime.ConfigPath)
	}

	if !anythingChanged {
		slog.Info("nothing changed, nothing more to do")
		return nil
	}

	// Ensure D-Bus is installed and running if using systemd
	if _, err := containerd.ListSystemdUnits(); err == nil {
		err = containerd.InstallDbus()
		if err != nil {
			return fmt.Errorf("failed to install D-Bus: %w", err)
		}
	}

	slog.Info("restarting containerd")
	err = containerdConfig.RestartRuntime()
	if err != nil {
		return fmt.Errorf("failed to restart containerd: %w", err)
	}

	return nil
}

func RuntimeOptions() (map[string]string, error) {
	runtimeOptions := make(map[string]string)
	optionsJSON := os.Getenv("RUNTIME_OPTIONS")
	config.Runtime.Options = make(map[string]string)
	if optionsJSON != "" {
		err := json.Unmarshal([]byte(optionsJSON), &runtimeOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal runtime options JSON %s: %w", optionsJSON, err)
		}
	}
	return runtimeOptions, nil
}
