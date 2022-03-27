package config

import (
	"os"
	"path/filepath"
)

var (
	CAFile         = configFile("ca.pem")
	ServerCertFile = configFile("server.pem")
	ServerKeyFile  = configFile("server-key.pem")
	ClientCertFile = configFile("client.pem")
	ClientKeyFile  = configFile("client-key.pem")

	RootClientCertFile = configFile("root-client.pem")
	RootClientKeyFile  = configFile("root-client-key.pem")

	UnauthClientCertFile = configFile("unauthorized-client.pem")
	UnauthClientKeyFile  = configFile("unauthorized-client-key.pem")

	ACLModelFile  = configFile("model.conf")
	ACLPolicyFile = configFile("policy.csv")
)

func configFile(filename string) string {
	if dir := os.Getenv("CONFIG_DIR"); dir != "" {
		return filepath.Join(dir, filename)
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	filePath := filepath.Join(homeDir, ".golog", filename)

	//fmt.Printf("filePath: %s\n", filePath)

	return filePath
}
