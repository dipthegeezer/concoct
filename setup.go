package main
import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/packer/plugin"
	"./packer/api"
)

func packerMeta() api.Meta {
	config, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: \n\n%s\n", err)
	}
	log.Printf("Packer config: %+v", config)

	cacheDir := os.Getenv("PACKER_CACHE_DIR")
	if cacheDir == "" {
		cacheDir = "packer_cache"
	}

	cacheDir, err = filepath.Abs(cacheDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error preparing cache directory: \n\n%s\n", err)
	}

	log.Printf("Setting cache directory: %s", cacheDir)
	cache := &packer.FileCache{CacheDir: cacheDir}

	defer plugin.CleanupClients()

	m := api.Meta{}
	m.EnvConfig = packer.DefaultEnvironmentConfig()
	m.EnvConfig.Cache = cache
	m.EnvConfig.Components.Builder = config.LoadBuilder
	m.EnvConfig.Components.Hook = config.LoadHook
	m.EnvConfig.Components.PostProcessor = config.LoadPostProcessor
	m.EnvConfig.Components.Provisioner = config.LoadProvisioner

	return m
}

func loadConfig() (*config, error) {
	var config config
	config.PluginMinPort = 10000
	config.PluginMaxPort = 25000
	if err := config.Discover(); err != nil {
		return nil, err
	}

	mustExist := true
	configFilePath := os.Getenv("PACKER_CONFIG")
	if configFilePath == "" {
		var err error
		configFilePath, err = configFile()
		mustExist = false

		if err != nil {
			log.Printf("Error detecting default config file path: %s", err)
		}
	}

	if configFilePath == "" {
		return &config, nil
	}

	log.Printf("Attempting to open config file: %s", configFilePath)
	f, err := os.Open(configFilePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}

		if mustExist {
			return nil, err
		}

		log.Println("File doesn't exist, but doesn't need to. Ignoring.")
		return &config, nil
	}
	defer f.Close()

	if err := decodeConfig(f, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

