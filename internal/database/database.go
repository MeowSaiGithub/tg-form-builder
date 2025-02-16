package database

import (
	"fmt"
	"go-tg-support-ticket/form"
	"net/url"
	"strings"
)

type Adaptor interface {
	Open(dns string) error
	GetName() string

	Migrate(schema *form.Form) error

	InsertUserInputs(tableName string, fields []form.Field) error
}

type Config struct {
	Enable         bool             `mapstructure:"enable"`
	UseAdaptor     string           `mapstructure:"use_adaptor"`
	MySQLConfig    MySQLConfig      `mapstructure:"mysql"`
	MongoConfig    MongoConfig      `mapstructure:"mongo"`
	PostgresConfig PostgreSQLConfig `mapstructure:"postgres"`
}

type MySQLConfig struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Database string `mapstructure:"database"`
	DSN      string `mapstructure:"dsn"`
}

type MongoConfig struct {
	Addresses     []string `mapstructure:"addresses"`
	Database      string   `mapstructure:"database"`
	ReplicaSet    string   `mapstructure:"replica_set"`
	AuthMechanism string   `mapstructure:"auth_mechanism"`
	Username      string   `mapstructure:"username"`
	Password      string   `mapstructure:"password"`
	URI           string   `mapstructure:"uri"`
}

type PostgreSQLConfig struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Database string `mapstructure:"database"`
	DSN      string `mapstructure:"dsn"`
}

// ParseConfig validates and processes the configuration
func ParseConfig(cfg *Config) (string, error) {
	switch cfg.UseAdaptor {
	case "mysql":
		return parseMySQLConfig(&cfg.MySQLConfig)
	case "mongo":
		return parseMongoConfig(&cfg.MongoConfig)
	case "postgres":
		return parsePostgresConfig(&cfg.PostgresConfig)
	default:
		return "", fmt.Errorf("invalid use_adaptor value, must be 'mysql', 'mongo' or 'postgres'")
	}
}

// parseMySQLConfig validates MySQL config and builds DSN if needed
func parseMySQLConfig(mysql *MySQLConfig) (string, error) {
	if mysql.DSN == "" {
		// Validate required fields
		if mysql.Username == "" || mysql.Password == "" || mysql.Host == "" || mysql.Port == "" || mysql.Database == "" {
			return "", fmt.Errorf("mysql config error: missing required fields")
		}

		// Construct DSN
		mysql.DSN = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4",
			mysql.Username, mysql.Password, mysql.Host, mysql.Port, mysql.Database)
	}

	// Validate DSN format
	if !strings.Contains(mysql.DSN, "@tcp") {
		return "", fmt.Errorf("invalid MySQL DSN format")
	}

	return mysql.DSN, nil
}

// parseMongoConfig validates MongoDB config and builds URI if needed
func parseMongoConfig(mongo *MongoConfig) (string, error) {
	if mongo.URI == "" {
		// Validate required fields
		if len(mongo.Addresses) == 0 || mongo.Database == "" || mongo.Username == "" || mongo.Password == "" {
			return "", fmt.Errorf("mongo config error: missing required fields")
		}

		// Build MongoDB connection URI
		addresses := strings.Join(mongo.Addresses, ",")
		mongo.URI = fmt.Sprintf("mongodb://%s:%s@%s/%s",
			url.QueryEscape(mongo.Username), url.QueryEscape(mongo.Password), addresses, mongo.Database)

		// Add optional parameters
		queryParams := []string{}
		if mongo.AuthMechanism != "" {
			queryParams = append(queryParams, "authMechanism="+url.QueryEscape(mongo.AuthMechanism))
		}
		if mongo.ReplicaSet != "" {
			queryParams = append(queryParams, "replicaSet="+url.QueryEscape(mongo.ReplicaSet))
		}
		if len(queryParams) > 0 {
			mongo.URI += "?" + strings.Join(queryParams, "&")
		}
	}

	// Validate MongoDB URI format
	if !strings.HasPrefix(mongo.URI, "mongodb://") {
		return "", fmt.Errorf("invalid MongoDB URI format")
	}

	return mongo.URI, nil
}

// parsePostgresConfig validates PostgreSQL config and builds DSN if needed
func parsePostgresConfig(pg *PostgreSQLConfig) (string, error) {
	if pg.DSN == "" {
		// Validate required fields
		if pg.Username == "" || pg.Password == "" || pg.Host == "" || pg.Port == "" || pg.Database == "" {
			return "", fmt.Errorf("postgres config error: missing required fields")
		}

		// Construct DSN
		pg.DSN = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			pg.Username, pg.Password, pg.Host, pg.Port, pg.Database)
	}

	// Validate DSN format (simple check for "postgres://")
	if !strings.HasPrefix(pg.DSN, "postgres://") {
		return "", fmt.Errorf("invalid PostgreSQL DSN format")
	}

	return pg.DSN, nil
}
