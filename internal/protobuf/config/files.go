package config

import (
	"os"
	"path/filepath"
)

var (
	CAFile               = configFile("ca.pem")
	ServerCertFile       = configFile("server.pem")
	ServerKeyFile        = configFile("server.key.pem")
	RootClientCertFile   = configFile("root-client.key.pem")
	RootClientKeyFile    = configFile("root-client.pem")
	NobodyClientCertFile = configFile("nobody-client.key.pem")
	NobodyClientKeyFile  = configFile("nobody-client.key.pem")
	ACLModelFile         = configFile("model.conf")
	ACLPolicyFile        = configFile("policy.csv")
)

func configFile(filename string) string {
	if dir := os.Getenv("CONFIG_DIR"); dir != "" {
		return configFile(filepath.Join(dir, filename))
	}
	workingDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return configFile(filepath.Join(workingDir, ".proglog", filename))
}
