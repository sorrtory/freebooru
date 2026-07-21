package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

type ConfigParser struct {
	log *slog.Logger
	// configDir      string
	// configFile     string

	// storageFile    string
	// tagsDir        string
	// collectionsDir string
}

type ConfigStruct any

type CheckStatus int

const (
	NOT_EXIST CheckStatus = iota
	READ_ERROR
	CONFIG_ERROR
	OK
)

type Parser interface {
	// Check the configuration file and return the status
	Check() (CheckStatus, error)
	// Read the configuration file and return the configuration struct
	Read() (ConfigStruct, error)
	// Write the configuration struct to the configuration file
	Write(ConfigStruct) error
	// Get the default configuration struct
	GetDefaultConfig() ConfigStruct
}

func NewConfigParser(log *slog.Logger) (*ConfigParser, error) {
	if log == nil {
		return nil, fmt.Errorf("logger is nil")
	}

	// // Set up pathes for config dir
	// configDir, err := getConfigDir()
	// if err != nil {
	// 	panic(err)
	// }

	// // app config
	// configFile := filepath.Join(configDir, FREEBOORU_CONFIG_DIR, FREEBOORU_CONFIG_FILE)
	// appConfigParser, err := NewAppConfigParser(log, configFile)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to create app config parser: %v", err)
	// }

	// // storage config
	// cfgParser.storageFile = filepath.Join(configDir, FREEBOORU_CONFIG_DIR, FREEBOORU_STORAGE_FILE)

	// // tags dir
	// cfgParser.tagsDir = filepath.Join(configDir, FREEBOORU_CONFIG_DIR, FREEBOORU_TAGS_DIR)

	// // collections dir
	// cfgParser.collectionsDir = filepath.Join(configDir, FREEBOORU_CONFIG_DIR, FREEBOORU_COLLECTIONS_DIR)

	return &ConfigParser{
		log: log,
	}, nil
}

func (cp *ConfigParser) Parse(parser Parser) (any, error) {
	status, err := parser.Check()
	switch status {
		case NOT_EXIST:
			cp.log.Info("Configuration file does not exist, creating default configuration")
			defaultConfig := parser.GetDefaultConfig()
			if err := parser.Write(defaultConfig); err != nil {
				return nil, fmt.Errorf("failed to write default configuration: %v", err)
			}
		case CONFIG_ERROR:
			return nil, fmt.Errorf("configuration file is invalid: %v", err)
		case OK:
			return parser.Read()
	}

	// // Set up pathes for config dir
	// configDir, err := getConfigDir()
	// if err != nil {
	// 	panic(err)
	// }
	// // Ensure the file structure exists
	// if err := cp.ensureFileStructure(); err != nil {
	// 	return fmt.Errorf("failed to ensure file structure: %v", err)
	// }

	// Read the freebooru main configuration file

	// cp.log.Debug("Configuration parsed successfully")
	return nil, fmt.Errorf("never reached here")
}

// func (cp *ConfigParser) ensureFileStructure() error {

// 	// Create the freebooru config directory if it doesn't exist
// 	err := cp.createDirIfNotExists(cp.configDir)
// 	if err != nil {
// 		return fmt.Errorf("failed to create config dir: %v", err)
// 	}

// 	// Create the config file if it doesn't exist
// 	err = cp.createFileIfNotExists(cp.configFile)
// 	if err != nil {
// 		return fmt.Errorf("failed to create config file: %v", err)
// 	}

// 	// Create the storage file if it doesn't exist
// 	err = cp.createFileIfNotExists(cp.storageFile)
// 	if err != nil {
// 		return fmt.Errorf("failed to create storage file: %v", err)
// 	}

// 	// Create the tags directory if it doesn't exist
// 	err = cp.createDirIfNotExists(cp.tagsDir)
// 	if err != nil {
// 		return fmt.Errorf("failed to create tags dir: %v", err)
// 	}

// 	// Create the collections directory if it doesn't exist
// 	err = cp.createDirIfNotExists(cp.collectionsDir)
// 	if err != nil {
// 		return fmt.Errorf("failed to create collections dir: %v", err)
// 	}

// 	return nil
// }
