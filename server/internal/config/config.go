package config

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App           AppConfig           `mapstructure:"app"`
	HTTP          HTTPConfig          `mapstructure:"http"`
	Database      DatabaseConfig      `mapstructure:"database"`
	Timeseries    TimeseriesConfig    `mapstructure:"timeseries"`
	Redis         RedisConfig         `mapstructure:"redis"`
	NATS          NATSConfig          `mapstructure:"nats"`
	JWT           JWTConfig           `mapstructure:"jwt"`
	AI            AIConfig            `mapstructure:"ai"`
	Push          PushConfig          `mapstructure:"push"`
	ObjectStorage ObjectStorageConfig `mapstructure:"object_storage"`
}

type AppConfig struct {
	Env  string `mapstructure:"env"`
	Name string `mapstructure:"name"`
}

type HTTPConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

func (c HTTPConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
	SSLMode  string `mapstructure:"sslmode"`
	Timezone string `mapstructure:"timezone"`
}

func (c DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		c.Host,
		c.Port,
		c.User,
		c.Password,
		c.Name,
		c.SSLMode,
		c.Timezone,
	)
}

func (c DatabaseConfig) URL() string {
	values := url.Values{}
	values.Set("sslmode", c.SSLMode)
	if c.Timezone != "" {
		values.Set("timezone", c.Timezone)
	}

	return (&url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(c.User, c.Password),
		Host:     fmt.Sprintf("%s:%d", c.Host, c.Port),
		Path:     c.Name,
		RawQuery: values.Encode(),
	}).String()
}

type TimeseriesConfig struct {
	Enabled        bool `mapstructure:"enabled"`
	DatabaseConfig `mapstructure:",squash"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type NATSConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	URL     string `mapstructure:"url"`
}

type JWTConfig struct {
	Issuer     string        `mapstructure:"issuer"`
	Secret     string        `mapstructure:"secret"`
	AccessTTL  time.Duration `mapstructure:"access_ttl"`
	RefreshTTL time.Duration `mapstructure:"refresh_ttl"`
}

type AIConfig struct {
	Provider    string        `mapstructure:"provider"`
	BaseURL     string        `mapstructure:"base_url"`
	APIKey      string        `mapstructure:"api_key"`
	Model       string        `mapstructure:"model"`
	Temperature float64       `mapstructure:"temperature"`
	Timeout     time.Duration `mapstructure:"timeout"`
}

type PushConfig struct {
	Provider    string        `mapstructure:"provider"`
	ExpoURL     string        `mapstructure:"expo_url"`
	AccessToken string        `mapstructure:"access_token"`
	Timeout     time.Duration `mapstructure:"timeout"`
}

type ObjectStorageConfig struct {
	Provider      string `mapstructure:"provider"`
	LocalDir      string `mapstructure:"local_dir"`
	Endpoint      string `mapstructure:"endpoint"`
	AccessKey     string `mapstructure:"access_key"`
	SecretKey     string `mapstructure:"secret_key"`
	UseSSL        bool   `mapstructure:"use_ssl"`
	Bucket        string `mapstructure:"bucket"`
	PublicBaseURL string `mapstructure:"public_base_url"`
}

func Load(path string) (Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")
	v.SetEnvPrefix("PETVERSE")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return Config{}, fmt.Errorf("read config: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("unmarshal config: %w", err)
	}

	return cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("app.env", "dev")
	v.SetDefault("app.name", "petverse-api")

	v.SetDefault("http.host", "0.0.0.0")
	v.SetDefault("http.port", 8080)
	v.SetDefault("http.read_timeout", "10s")
	v.SetDefault("http.write_timeout", "15s")
	v.SetDefault("http.shutdown_timeout", "10s")

	v.SetDefault("database.host", "127.0.0.1")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.user", "postgres")
	v.SetDefault("database.password", "postgres")
	v.SetDefault("database.name", "petverse")
	v.SetDefault("database.sslmode", "disable")
	v.SetDefault("database.timezone", "Asia/Shanghai")

	v.SetDefault("timeseries.enabled", false)
	v.SetDefault("timeseries.host", "127.0.0.1")
	v.SetDefault("timeseries.port", 5433)
	v.SetDefault("timeseries.user", "postgres")
	v.SetDefault("timeseries.password", "postgres")
	v.SetDefault("timeseries.name", "petverse_ts")
	v.SetDefault("timeseries.sslmode", "disable")
	v.SetDefault("timeseries.timezone", "Asia/Shanghai")

	v.SetDefault("redis.host", "127.0.0.1")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)

	v.SetDefault("nats.enabled", false)
	v.SetDefault("nats.url", "nats://127.0.0.1:4222")

	v.SetDefault("jwt.issuer", "petverse")
	v.SetDefault("jwt.secret", "change-me-in-production")
	v.SetDefault("jwt.access_ttl", "15m")
	v.SetDefault("jwt.refresh_ttl", "168h")

	v.SetDefault("ai.provider", "local")
	v.SetDefault("ai.base_url", "")
	v.SetDefault("ai.api_key", "")
	v.SetDefault("ai.model", "")
	v.SetDefault("ai.temperature", 0.2)
	v.SetDefault("ai.timeout", "20s")

	v.SetDefault("push.provider", "none")
	v.SetDefault("push.expo_url", "https://exp.host/--/api/v2/push/send")
	v.SetDefault("push.access_token", "")
	v.SetDefault("push.timeout", "10s")

	v.SetDefault("object_storage.provider", "local")
	v.SetDefault("object_storage.local_dir", "./uploads")
	v.SetDefault("object_storage.endpoint", "127.0.0.1:9000")
	v.SetDefault("object_storage.access_key", "minioadmin")
	v.SetDefault("object_storage.secret_key", "minioadmin")
	v.SetDefault("object_storage.use_ssl", false)
	v.SetDefault("object_storage.bucket", "petverse")
	v.SetDefault("object_storage.public_base_url", "")
}
