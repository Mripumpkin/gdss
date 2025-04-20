package config

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/spf13/viper"
)

// SettingsConf holds the application configuration.
type SettingsConf struct {
	App App `mapstructure:"app"`
}

// App contains application-specific settings.
type App struct {
	Secret string `mapstructure:"jwt_secret_token"`
	Salt   string `mapstructure:"salt"`
}

var Settings SettingsConf

// Provider defines read-only methods for accessing configuration parameters.
type Provider interface {
	ConfigFileUsed() string
	Get(key string) interface{}
	GetBool(key string) bool
	GetDuration(key string) time.Duration
	GetFloat64(key string) float64
	GetInt(key string) int
	GetInt64(key string) int64
	GetSizeInBytes(key string) uint
	GetString(key string) string
	GetStringMap(key string) map[string]interface{}
	GetStringMapString(key string) map[string]string
	GetStringMapStringSlice(key string) map[string][]string
	GetStringSlice(key string) []string
	GetTime(key string) time.Time
	InConfig(key string) bool
	IsSet(key string) bool
}

var defaultConfig *viper.Viper
var once sync.Once

// Config returns the default configuration provider.
func Config() Provider {
	return defaultConfig
}

// LoadConfigProvider loads and returns a configured viper instance.
func LoadConfigProvider() (Provider, error) {
	once.Do(func() {
		var err error
		defaultConfig, err = readViperConfig()
		if err != nil {
			defaultConfig = viper.New() // Fallback to empty config
			defaultConfig.SetDefault("loglevel", "info")
		}
	})
	if defaultConfig == nil {
		return nil, fmt.Errorf("failed to initialize configuration")
	}
	return defaultConfig, nil
}

func init() {
	var err error
	defaultConfig, err = readViperConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error loading config: %w", err))
	}
}

// readViperConfig initializes a viper instance with configuration from files and environment variables.
func readViperConfig() (*viper.Viper, error) {
	v := viper.New()
	v.SetEnvPrefix("gdss")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Default configuration paths and file
	configPaths := []string{".", "/srv/gdss_dev", "/etc/gdss"}
	configName := "settings"
	configType := "toml"

	// Allow override via environment variables
	if envPath := os.Getenv("GDSS_CONFIG_PATH"); envPath != "" {
		configPaths = append([]string{envPath}, configPaths...)
	}
	if envName := os.Getenv("GDSS_CONFIG_NAME"); envName != "" {
		configName = envName
	}
	if envType := os.Getenv("GDSS_CONFIG_TYPE"); envType != "" {
		configType = envType
	}

	// Set configuration paths and file details
	for _, path := range configPaths {
		v.AddConfigPath(path)
	}
	v.SetConfigName(configName)
	v.SetConfigType(configType)

	// Read main configuration
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file %s.%s: %w", configName, configType, err)
	}

	// Set default values for logging
	v.SetDefault("json_logs", false)
	v.SetDefault("loglevel", "info")
	v.SetDefault("log_file", "") // Empty means stderr
	v.SetDefault("log_max_size_mb", 100)
	v.SetDefault("log_max_backups", 3)
	v.SetDefault("log_max_age_days", 28)
	v.SetDefault("log_compress", false)

	// Unmarshal to Settings struct
	if err := v.Unmarshal(&Settings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Load development configuration if applicable
	if runLevel := v.GetString("run_level"); runLevel == "development" {
		v.SetConfigName("settings.local")
		v.SetConfigType("toml")
		v.AddConfigPath(".")
		if err := v.MergeInConfig(); err != nil {
			return nil, fmt.Errorf("failed to merge development config: %w", err)
		}
	}

	return v, nil
}
