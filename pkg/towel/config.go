package towel

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

type Config struct {
	Version  int    `json:"version"`
	Email    string `json:"email"`
	ApiToken string `json:"api_token"`
	Server   string `json:"server"`
}

// loadConfig parses and validates the configuration file.
func loadConfig(configDir string) (Config, error) {
	configFile := configFile(configDir)
	file, err := os.Open(configFile)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return Config{},
				fmt.Errorf("opening configuration file %q: not found. Run 'jira-towel init' to create it", configFile)
		}
		return Config{},
			fmt.Errorf("opening configuration file %q: %s", configFile, err)
	}
	fileInfo, err := file.Stat()
	if err != nil {
		return Config{},
			fmt.Errorf("getting info about configuration file %q: %s", configFile, err)
	}
	perm := fileInfo.Mode().Perm()
	if perm != 0o600 && perm != 0o400 {
		return Config{},
			fmt.Errorf("configuration file %q must be readable only by the owner (0600), while it has permissions 0%o",
				configFile, perm)
	}

	buf, err := io.ReadAll(file)
	if err != nil {
		return Config{}, fmt.Errorf("reading configuration file %q: %s", configFile, err)
	}

	var config Config
	if err := json.Unmarshal(buf, &config); err != nil {
		return Config{}, err
	}

	if err := validateConfig(config); err != nil {
		return Config{}, fmt.Errorf("config file %q: %s", configFile, err)
	}

	return config, nil
}

// TODO actually validate something!
func validateConfig(config Config) error {
	return nil
}

// initConfig initialises the configuration file.
func initConfig(configDir string) error {
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		return err
	}

	configFile := configFile(configDir)
	_, err := os.Stat(configFile)
	if err == nil {
		return fmt.Errorf("file already exists: %s", configFile)
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	config := Config{
		Version:  1,
		Email:    "The email of the JIRA user with the API token; see README",
		ApiToken: "The Jira API token; see README",
		Server:   "URL to your JIRA instance",
	}

	buf, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return err
	}

	file, err := os.OpenFile(configFile, os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		return err
	}

	if _, err := file.Write(buf); err != nil {
		return err
	}
	fmt.Printf("init: created %s\n", configFile)

	return nil
}

// defaultConfigDir returns the OS-default configuration directory for
// jira-towel.
func defaultConfigDir() (string, error) {
	baseConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("retrieving the user configuration directory: %w", err)
	}
	return filepath.Join(baseConfigDir, "jira-towel"), nil
}

func configFile(configDir string) string {
	return filepath.Join(configDir, "jira-towel.json")
}
