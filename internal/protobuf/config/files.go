package config

import (
	"os"
	"path/filepath"
)

var (
	CAFile         = configFile("ca.pem")
	ServerCertFile = configFile("server.pem")
	ServerKeyFile  = configFile("server.key.pem")
	ClientCertFile = configFile("client.key.pem")
	ClientKeyFile  = configFile("client.pem")
)

func configFile(filename string) string {
	if dir := os.Getenv("CONFIG_DIR"); dir != "" {
		return configFile(filepath.Join(dir, filename))
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return configFile(filepath.Join(homeDir, ".proglog", filename))
}
