package util

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/tapclap/db-connect-checker/pkg/types"
)

// MockFileReader is a mock implementation of FileReader for testing
type MockFileReader struct {
	ReadFileFunc func(filename string) ([]byte, error)
}

func (m MockFileReader) ReadFile(filename string) ([]byte, error) {
	if m.ReadFileFunc != nil {
		return m.ReadFileFunc(filename)
	}
	return nil, errors.New("mock not configured")
}

// generateTestCertificate creates a valid PEM certificate for testing
func generateTestCertificate(t *testing.T) []byte {
	t.Helper()

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test CA"},
			CommonName:   "Test CA",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		t.Fatalf("Failed to create certificate: %v", err)
	}

	return pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: derBytes,
	})
}

func TestGetEnvString(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		setEnv       bool
		expected     string
	}{
		{
			name:         "returns environment value when set",
			key:          "TEST_KEY",
			defaultValue: "default",
			envValue:     "test_value",
			setEnv:       true,
			expected:     "test_value",
		},
		{
			name:         "returns default value when env not set",
			key:          "TEST_KEY_NOT_SET",
			defaultValue: "default",
			envValue:     "",
			setEnv:       false,
			expected:     "default",
		},
		{
			name:         "returns default value when env is empty string",
			key:          "TEST_KEY_EMPTY",
			defaultValue: "default",
			envValue:     "",
			setEnv:       true,
			expected:     "default",
		},
		{
			name:         "returns empty string when both are empty",
			key:          "TEST_KEY_BOTH_EMPTY",
			defaultValue: "",
			envValue:     "",
			setEnv:       false,
			expected:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up before test
			os.Unsetenv(tt.key)

			// Set environment variable if needed
			if tt.setEnv {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			result := GetEnvString(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("GetEnvString() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetEnvBool(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue bool
		envValue     string
		setEnv       bool
		expected     bool
	}{
		{
			name:         "returns true when env is 'true'",
			key:          "TEST_BOOL_KEY",
			defaultValue: false,
			envValue:     "true",
			setEnv:       true,
			expected:     true,
		},
		{
			name:         "returns false when env is 'false'",
			key:          "TEST_BOOL_KEY_FALSE",
			defaultValue: true,
			envValue:     "false",
			setEnv:       true,
			expected:     false,
		},
		{
			name:         "returns false when env is random string",
			key:          "TEST_BOOL_KEY_RANDOM",
			defaultValue: true,
			envValue:     "random",
			setEnv:       true,
			expected:     false,
		},
		{
			name:         "returns default value when env not set",
			key:          "TEST_BOOL_KEY_NOT_SET",
			defaultValue: true,
			envValue:     "",
			setEnv:       false,
			expected:     true,
		},
		{
			name:         "returns default false when env not set",
			key:          "TEST_BOOL_KEY_DEFAULT_FALSE",
			defaultValue: false,
			envValue:     "",
			setEnv:       false,
			expected:     false,
		},
		{
			name:         "returns default when env is empty string",
			key:          "TEST_BOOL_KEY_EMPTY",
			defaultValue: true,
			envValue:     "",
			setEnv:       true,
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up before test
			os.Unsetenv(tt.key)

			// Set environment variable if needed
			if tt.setEnv {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			result := GetEnvBool(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("GetEnvBool() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetMysqlConfigFromEnvsByIndex(t *testing.T) {
	tests := []struct {
		name     string
		index    int
		envVars  map[string]string
		wantErr  bool
		expected types.MysqlConfig
	}{
		{
			name:  "returns config when all env vars set for index 0",
			index: 0,
			envVars: map[string]string{
				"MYSQL_NAME_0": "testdb",
				"MYSQL_USER_0": "testuser",
				"MYSQL_PASS_0": "testpass",
				"MYSQL_HOST_0": "localhost",
				"MYSQL_PORT_0": "3306",
			},
			wantErr: false,
			expected: types.MysqlConfig{
				Name: "testdb",
				User: "testuser",
				Pass: "testpass",
				Host: "localhost",
				Port: "3306",
			},
		},
		{
			name:  "returns config with default port when port not set",
			index: 1,
			envVars: map[string]string{
				"MYSQL_NAME_1": "testdb",
				"MYSQL_USER_1": "testuser",
				"MYSQL_PASS_1": "testpass",
				"MYSQL_HOST_1": "localhost",
			},
			wantErr: false,
			expected: types.MysqlConfig{
				Name: "testdb",
				User: "testuser",
				Pass: "testpass",
				Host: "localhost",
				Port: "3306",
			},
		},
		{
			name:  "returns error when name is missing",
			index: 2,
			envVars: map[string]string{
				"MYSQL_USER_2": "testuser",
				"MYSQL_PASS_2": "testpass",
				"MYSQL_HOST_2": "localhost",
				"MYSQL_PORT_2": "3306",
			},
			wantErr: true,
		},
		{
			name:  "returns error when user is missing",
			index: 3,
			envVars: map[string]string{
				"MYSQL_NAME_3": "testdb",
				"MYSQL_PASS_3": "testpass",
				"MYSQL_HOST_3": "localhost",
				"MYSQL_PORT_3": "3306",
			},
			wantErr: true,
		},
		{
			name:  "returns error when pass is missing",
			index: 4,
			envVars: map[string]string{
				"MYSQL_NAME_4": "testdb",
				"MYSQL_USER_4": "testuser",
				"MYSQL_HOST_4": "localhost",
				"MYSQL_PORT_4": "3306",
			},
			wantErr: true,
		},
		{
			name:  "returns error when host is missing",
			index: 5,
			envVars: map[string]string{
				"MYSQL_NAME_5": "testdb",
				"MYSQL_USER_5": "testuser",
				"MYSQL_PASS_5": "testpass",
				"MYSQL_PORT_5": "3306",
			},
			wantErr: true,
		},
		{
			name:    "returns error when no env vars set",
			index:   6,
			envVars: map[string]string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up all possible env vars for this index
			envKeys := []string{
				"MYSQL_NAME_" + string(rune(tt.index+'0')),
				"MYSQL_USER_" + string(rune(tt.index+'0')),
				"MYSQL_PASS_" + string(rune(tt.index+'0')),
				"MYSQL_HOST_" + string(rune(tt.index+'0')),
				"MYSQL_PORT_" + string(rune(tt.index+'0')),
			}
			for _, key := range envKeys {
				os.Unsetenv(key)
			}

			// Set environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
				defer os.Unsetenv(key)
			}

			result, err := getMysqlConfigFromEnvsByIndex(tt.index)

			if tt.wantErr {
				if err == nil {
					t.Error("getMysqlConfigFromEnvsByIndex() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("getMysqlConfigFromEnvsByIndex() unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("getMysqlConfigFromEnvsByIndex() = %v, want %v", result, tt.expected)
				}
			}
		})
	}
}

func TestGetMysqlConfigFromEnvs(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected types.MysqlConfig
	}{
		{
			name: "returns config when all env vars set",
			envVars: map[string]string{
				"MYSQL_NAME": "maindb",
				"MYSQL_USER": "mainuser",
				"MYSQL_PASS": "mainpass",
				"MYSQL_HOST": "mainhost",
				"MYSQL_PORT": "3307",
			},
			expected: types.MysqlConfig{
				Name: "maindb",
				User: "mainuser",
				Pass: "mainpass",
				Host: "mainhost",
				Port: "3307",
			},
		},
		{
			name: "returns config with default port when port not set",
			envVars: map[string]string{
				"MYSQL_NAME": "maindb",
				"MYSQL_USER": "mainuser",
				"MYSQL_PASS": "mainpass",
				"MYSQL_HOST": "mainhost",
			},
			expected: types.MysqlConfig{
				Name: "maindb",
				User: "mainuser",
				Pass: "mainpass",
				Host: "mainhost",
				Port: "3306",
			},
		},
		{
			name:    "returns empty config when no env vars set",
			envVars: map[string]string{},
			expected: types.MysqlConfig{
				Name: "",
				User: "",
				Pass: "",
				Host: "",
				Port: "3306",
			},
		},
		{
			name: "returns partial config when some env vars set",
			envVars: map[string]string{
				"MYSQL_NAME": "partialdb",
				"MYSQL_USER": "partialuser",
			},
			expected: types.MysqlConfig{
				Name: "partialdb",
				User: "partialuser",
				Pass: "",
				Host: "",
				Port: "3306",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up all env vars
			envKeys := []string{
				"MYSQL_NAME",
				"MYSQL_USER",
				"MYSQL_PASS",
				"MYSQL_HOST",
				"MYSQL_PORT",
			}
			for _, key := range envKeys {
				os.Unsetenv(key)
			}

			// Set environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
				defer os.Unsetenv(key)
			}

			result := getMysqlConfigFromEnvs()

			if result != tt.expected {
				t.Errorf("getMysqlConfigFromEnvs() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetAllMysqlConfigsFromEnvs(t *testing.T) {
	tests := []struct {
		name          string
		envVars       map[string]string
		expectedCount int
		checkConfigs  func(t *testing.T, configs []types.MysqlConfig)
	}{
		{
			name: "returns single base config when only base env vars set",
			envVars: map[string]string{
				"MYSQL_NAME": "basedb",
				"MYSQL_USER": "baseuser",
				"MYSQL_PASS": "basepass",
				"MYSQL_HOST": "basehost",
				"MYSQL_PORT": "3306",
			},
			expectedCount: 1,
			checkConfigs: func(t *testing.T, configs []types.MysqlConfig) {
				if configs[0].Name != "basedb" {
					t.Errorf("Expected base config name to be 'basedb', got '%s'", configs[0].Name)
				}
			},
		},
		{
			name: "returns indexed and base configs when both set",
			envVars: map[string]string{
				"MYSQL_NAME_0": "indexdb0",
				"MYSQL_USER_0": "indexuser0",
				"MYSQL_PASS_0": "indexpass0",
				"MYSQL_HOST_0": "indexhost0",
				"MYSQL_PORT_0": "3306",
				"MYSQL_NAME":   "basedb",
				"MYSQL_USER":   "baseuser",
				"MYSQL_PASS":   "basepass",
				"MYSQL_HOST":   "basehost",
				"MYSQL_PORT":   "3307",
			},
			expectedCount: 2,
			checkConfigs: func(t *testing.T, configs []types.MysqlConfig) {
				if configs[0].Name != "indexdb0" {
					t.Errorf("Expected first config name to be 'indexdb0', got '%s'", configs[0].Name)
				}
				if configs[1].Name != "basedb" {
					t.Errorf("Expected second config name to be 'basedb', got '%s'", configs[1].Name)
				}
			},
		},
		{
			name: "returns multiple indexed configs and base config",
			envVars: map[string]string{
				"MYSQL_NAME_0": "indexdb0",
				"MYSQL_USER_0": "indexuser0",
				"MYSQL_PASS_0": "indexpass0",
				"MYSQL_HOST_0": "indexhost0",
				"MYSQL_PORT_0": "3306",
				"MYSQL_NAME_1": "indexdb1",
				"MYSQL_USER_1": "indexuser1",
				"MYSQL_PASS_1": "indexpass1",
				"MYSQL_HOST_1": "indexhost1",
				"MYSQL_PORT_1": "3307",
				"MYSQL_NAME":   "basedb",
				"MYSQL_USER":   "baseuser",
				"MYSQL_PASS":   "basepass",
				"MYSQL_HOST":   "basehost",
				"MYSQL_PORT":   "3308",
			},
			expectedCount: 3,
			checkConfigs: func(t *testing.T, configs []types.MysqlConfig) {
				if configs[0].Name != "indexdb0" {
					t.Errorf("Expected first config name to be 'indexdb0', got '%s'", configs[0].Name)
				}
				if configs[1].Name != "indexdb1" {
					t.Errorf("Expected second config name to be 'indexdb1', got '%s'", configs[1].Name)
				}
				if configs[2].Name != "basedb" {
					t.Errorf("Expected third config name to be 'basedb', got '%s'", configs[2].Name)
				}
			},
		},
		{
			name:          "returns empty slice when no env vars set",
			envVars:       map[string]string{},
			expectedCount: 0,
		},
		{
			name: "stops reading indexed configs at first gap",
			envVars: map[string]string{
				"MYSQL_NAME_0": "indexdb0",
				"MYSQL_USER_0": "indexuser0",
				"MYSQL_PASS_0": "indexpass0",
				"MYSQL_HOST_0": "indexhost0",
				// Missing index 1
				"MYSQL_NAME_2": "indexdb2",
				"MYSQL_USER_2": "indexuser2",
				"MYSQL_PASS_2": "indexpass2",
				"MYSQL_HOST_2": "indexhost2",
				"MYSQL_NAME":   "basedb",
				"MYSQL_USER":   "baseuser",
				"MYSQL_PASS":   "basepass",
				"MYSQL_HOST":   "basehost",
			},
			expectedCount: 2, // Only index 0 and base config
			checkConfigs: func(t *testing.T, configs []types.MysqlConfig) {
				if configs[0].Name != "indexdb0" {
					t.Errorf("Expected first config name to be 'indexdb0', got '%s'", configs[0].Name)
				}
				if configs[1].Name != "basedb" {
					t.Errorf("Expected second config name to be 'basedb', got '%s'", configs[1].Name)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up all possible env vars
			envKeys := []string{
				"MYSQL_NAME", "MYSQL_USER", "MYSQL_PASS", "MYSQL_HOST", "MYSQL_PORT",
			}
			for i := 0; i < 10; i++ {
				envKeys = append(envKeys,
					"MYSQL_NAME_"+string(rune(i+'0')),
					"MYSQL_USER_"+string(rune(i+'0')),
					"MYSQL_PASS_"+string(rune(i+'0')),
					"MYSQL_HOST_"+string(rune(i+'0')),
					"MYSQL_PORT_"+string(rune(i+'0')),
				)
			}
			for _, key := range envKeys {
				os.Unsetenv(key)
			}

			// Set environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
				defer os.Unsetenv(key)
			}

			result := GetAllMysqlConfigsFromEnvs()

			if len(result) != tt.expectedCount {
				t.Errorf("GetAllMysqlConfigsFromEnvs() returned %d configs, want %d", len(result), tt.expectedCount)
			}

			if tt.checkConfigs != nil {
				tt.checkConfigs(t, result)
			}
		})
	}
}

func TestMysqlTLSConfig(t *testing.T) {
	validCert := generateTestCertificate(t)

	tests := []struct {
		name       string
		capath     string
		reader     FileReader
		wantNil    bool
		checkError bool
	}{
		{
			name:   "returns valid TLS config with valid CA certificate",
			capath: "/path/to/ca.pem",
			reader: MockFileReader{
				ReadFileFunc: func(filename string) ([]byte, error) {
					return validCert, nil
				},
			},
			wantNil: false,
		},
		{
			name:   "returns valid TLS config with real file",
			capath: "/path/to/ca.pem",
			reader: MockFileReader{
				ReadFileFunc: func(filename string) ([]byte, error) {
					return validCert, nil
				},
			},
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mysqlTLSConfig(tt.capath, tt.reader)

			if tt.wantNil {
				if result != nil {
					t.Error("mysqlTLSConfig() expected nil but got config")
				}
			} else {
				if result == nil {
					t.Error("mysqlTLSConfig() expected config but got nil")
				}
				if result != nil {
					if !result.InsecureSkipVerify {
						t.Error("mysqlTLSConfig() expected InsecureSkipVerify to be true")
					}
					if result.RootCAs == nil {
						t.Error("mysqlTLSConfig() expected RootCAs to be set")
					}
				}
			}
		})
	}
}

func TestGetMysqlConfigFromEnvsByIndexWithTLS(t *testing.T) {
	validCert := generateTestCertificate(t)

	// Create a temporary mock by replacing the defaultFileReader
	oldReader := defaultFileReader
	defer func() { defaultFileReader = oldReader }()

	defaultFileReader = MockFileReader{
		ReadFileFunc: func(filename string) ([]byte, error) {
			return validCert, nil
		},
	}

	tests := []struct {
		name        string
		index       int
		envVars     map[string]string
		wantErr     bool
		checkConfig func(t *testing.T, config types.MysqlConfig)
	}{
		{
			name:  "returns config with TLS enabled",
			index: 0,
			envVars: map[string]string{
				"MYSQL_NAME_0":        "testdb",
				"MYSQL_USER_0":        "testuser",
				"MYSQL_PASS_0":        "testpass",
				"MYSQL_HOST_0":        "localhost",
				"MYSQL_PORT_0":        "3306",
				"MYSQL_TLS_0":         "true",
				"MYSQL_TLS_CA_FILE_0": "/path/to/ca.pem",
			},
			wantErr: false,
			checkConfig: func(t *testing.T, config types.MysqlConfig) {
				if !config.TLS {
					t.Error("Expected TLS to be true")
				}
				if config.TLSConfig == nil {
					t.Error("Expected TLSConfig to be set")
				}
				if config.TLSConfig != nil && config.TLSConfig.RootCAs == nil {
					t.Error("Expected RootCAs to be set")
				}
			},
		},
		{
			name:  "returns config with TLS disabled by default",
			index: 1,
			envVars: map[string]string{
				"MYSQL_NAME_1": "testdb",
				"MYSQL_USER_1": "testuser",
				"MYSQL_PASS_1": "testpass",
				"MYSQL_HOST_1": "localhost",
				"MYSQL_PORT_1": "3306",
			},
			wantErr: false,
			checkConfig: func(t *testing.T, config types.MysqlConfig) {
				if config.TLS {
					t.Error("Expected TLS to be false by default")
				}
				if config.TLSConfig != nil {
					t.Error("Expected TLSConfig to be nil when TLS is disabled")
				}
			},
		},
		{
			name:  "returns config with TLS explicitly disabled",
			index: 2,
			envVars: map[string]string{
				"MYSQL_NAME_2": "testdb",
				"MYSQL_USER_2": "testuser",
				"MYSQL_PASS_2": "testpass",
				"MYSQL_HOST_2": "localhost",
				"MYSQL_PORT_2": "3306",
				"MYSQL_TLS_2":  "false",
			},
			wantErr: false,
			checkConfig: func(t *testing.T, config types.MysqlConfig) {
				if config.TLS {
					t.Error("Expected TLS to be false")
				}
				if config.TLSConfig != nil {
					t.Error("Expected TLSConfig to be nil when TLS is disabled")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up all possible env vars for this index
			envKeys := []string{
				"MYSQL_NAME_" + string(rune(tt.index+'0')),
				"MYSQL_USER_" + string(rune(tt.index+'0')),
				"MYSQL_PASS_" + string(rune(tt.index+'0')),
				"MYSQL_HOST_" + string(rune(tt.index+'0')),
				"MYSQL_PORT_" + string(rune(tt.index+'0')),
				"MYSQL_TLS_" + string(rune(tt.index+'0')),
				"MYSQL_TLS_CA_FILE_" + string(rune(tt.index+'0')),
			}
			for _, key := range envKeys {
				os.Unsetenv(key)
			}

			// Set environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
				defer os.Unsetenv(key)
			}

			result, err := getMysqlConfigFromEnvsByIndex(tt.index)

			if tt.wantErr {
				if err == nil {
					t.Error("getMysqlConfigFromEnvsByIndex() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("getMysqlConfigFromEnvsByIndex() unexpected error: %v", err)
				}
				if tt.checkConfig != nil {
					tt.checkConfig(t, result)
				}
			}
		})
	}
}

func TestGetMysqlConfigFromEnvsWithTLS(t *testing.T) {
	validCert := generateTestCertificate(t)

	// Create a temporary mock by replacing the defaultFileReader
	oldReader := defaultFileReader
	defer func() { defaultFileReader = oldReader }()

	defaultFileReader = MockFileReader{
		ReadFileFunc: func(filename string) ([]byte, error) {
			return validCert, nil
		},
	}

	tests := []struct {
		name        string
		envVars     map[string]string
		checkConfig func(t *testing.T, config types.MysqlConfig)
	}{
		{
			name: "returns config with TLS enabled",
			envVars: map[string]string{
				"MYSQL_NAME":        "maindb",
				"MYSQL_USER":        "mainuser",
				"MYSQL_PASS":        "mainpass",
				"MYSQL_HOST":        "mainhost",
				"MYSQL_PORT":        "3307",
				"MYSQL_TLS":         "true",
				"MYSQL_TLS_CA_FILE": "/path/to/ca.pem",
			},
			checkConfig: func(t *testing.T, config types.MysqlConfig) {
				if !config.TLS {
					t.Error("Expected TLS to be true")
				}
				if config.TLSConfig == nil {
					t.Error("Expected TLSConfig to be set")
				}
				if config.TLSConfig != nil && config.TLSConfig.RootCAs == nil {
					t.Error("Expected RootCAs to be set")
				}
			},
		},
		{
			name: "returns config with TLS disabled by default",
			envVars: map[string]string{
				"MYSQL_NAME": "maindb",
				"MYSQL_USER": "mainuser",
				"MYSQL_PASS": "mainpass",
				"MYSQL_HOST": "mainhost",
			},
			checkConfig: func(t *testing.T, config types.MysqlConfig) {
				if config.TLS {
					t.Error("Expected TLS to be false by default")
				}
				if config.TLSConfig != nil {
					t.Error("Expected TLSConfig to be nil when TLS is disabled")
				}
			},
		},
		{
			name: "returns config with TLS explicitly disabled",
			envVars: map[string]string{
				"MYSQL_NAME": "maindb",
				"MYSQL_USER": "mainuser",
				"MYSQL_PASS": "mainpass",
				"MYSQL_HOST": "mainhost",
				"MYSQL_TLS":  "false",
			},
			checkConfig: func(t *testing.T, config types.MysqlConfig) {
				if config.TLS {
					t.Error("Expected TLS to be false")
				}
				if config.TLSConfig != nil {
					t.Error("Expected TLSConfig to be nil when TLS is disabled")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up all env vars
			envKeys := []string{
				"MYSQL_NAME",
				"MYSQL_USER",
				"MYSQL_PASS",
				"MYSQL_HOST",
				"MYSQL_PORT",
				"MYSQL_TLS",
				"MYSQL_TLS_CA_FILE",
			}
			for _, key := range envKeys {
				os.Unsetenv(key)
			}

			// Set environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
				defer os.Unsetenv(key)
			}

			result := getMysqlConfigFromEnvs()

			if tt.checkConfig != nil {
				tt.checkConfig(t, result)
			}
		})
	}
}
