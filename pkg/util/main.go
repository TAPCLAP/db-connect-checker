package util

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strconv"

	"github.com/tapclap/db-connect-checker/pkg/types"
)

// FileReader interface for reading files
type FileReader interface {
	ReadFile(filename string) ([]byte, error)
}

// OsFileReader implements FileReader using os package
type OsFileReader struct{}

func (o OsFileReader) ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

// defaultFileReader is the default implementation used in production
var defaultFileReader FileReader = OsFileReader{}

func GetAllMysqlConfigsFromEnvs() []types.MysqlConfig {
	configs := []types.MysqlConfig{}
	for i := 0; true; i++ {
		config, err := getMysqlConfigFromEnvsByIndex(i)
		if err != nil {
			break
		}
		configs = append(configs, config)
	}

	config := getMysqlConfigFromEnvs()
	if config.Name != "" && config.User != "" && config.Pass != "" && config.Host != "" && config.Port != "" {
		configs = append(configs, config)
	}

	if len(configs) > 0 {
		fmt.Println("Discovered MySQL configurations from environment variables:")
		for _, config := range configs {
			fmt.Printf(" - %s@%s:%s/%s\n", config.User, config.Host, config.Port, config.Name)
		}
	}

	return configs
}

func getMysqlConfigFromEnvsByIndex(index int) (types.MysqlConfig, error) {
	var config types.MysqlConfig
	config.Name = GetEnvString(fmt.Sprintf("MYSQL_NAME_%d", index), "")
	config.User = GetEnvString(fmt.Sprintf("MYSQL_USER_%d", index), "")
	config.Pass = GetEnvString(fmt.Sprintf("MYSQL_PASS_%d", index), "")
	config.Host = GetEnvString(fmt.Sprintf("MYSQL_HOST_%d", index), "")
	config.Port = GetEnvString(fmt.Sprintf("MYSQL_PORT_%d", index), "3306")
	config.TLS = GetEnvBool(fmt.Sprintf("MYSQL_TLS_%d", index), false)

	capath := GetEnvString(fmt.Sprintf("MYSQL_TLS_CA_FILE_%d", index), "/etc/ssl/certs/ca-certificates.crt")
	if config.TLS {
		config.TLSConfig = mysqlTLSConfig(capath, defaultFileReader)
	}

	if config.Name == "" || config.User == "" || config.Pass == "" || config.Host == "" || config.Port == "" {
		return types.MysqlConfig{}, fmt.Errorf("no MySQL config found for index %d", index)
	}
	return config, nil
}

func getMysqlConfigFromEnvs() types.MysqlConfig {
	var config types.MysqlConfig
	config.Name = GetEnvString("MYSQL_NAME", "")
	config.User = GetEnvString("MYSQL_USER", "")
	config.Pass = GetEnvString("MYSQL_PASS", "")
	config.Host = GetEnvString("MYSQL_HOST", "")
	config.Port = GetEnvString("MYSQL_PORT", "3306")
	config.TLS = GetEnvBool("MYSQL_TLS", false)

	capath := GetEnvString("MYSQL_TLS_CA_FILE", "/etc/ssl/certs/ca-certificates.crt")
	if config.TLS {
		config.TLSConfig = mysqlTLSConfig(capath, defaultFileReader)
	}
	return config
}

func mysqlTLSConfig(capath string, reader FileReader) *tls.Config {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	rootCertPool := x509.NewCertPool()
	pem, err := reader.ReadFile(capath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading CA file: %v\n", err)
		os.Exit(1)

	}
	if ok := rootCertPool.AppendCertsFromPEM(pem); !ok {
		fmt.Fprintf(os.Stderr, "Error appending CA cert\n")
		os.Exit(1)
	}
	tlsConfig.RootCAs = rootCertPool
	return tlsConfig

}

func GetEnvString(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func GetEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value == "true"
}

func GetEnvNumber(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	num, err := strconv.Atoi(value)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error converting env %s value %s to number: %v\n", key, value, err)
		os.Exit(1)
	}
	return num
}
