package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	ServerAddress        string `env:"RUN_ADDRESS"`
	DatabaseDSN          string `env:"DATABASE_URI"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	JWTSecret            string `env:"JWT_SECRET"`

	LegacyServerAddress string `env:"SERVER_ADDRESS"`
	LegacyDatabaseDSN   string `env:"DATABASE_DSN"`
}

func ParseConfig() (config *Config) {

	defaultServerAddress := "localhost:8080"
	defaultDatabaseDSN := ""
	defaultAccrualSystemAddress := "http://localhost:8081"
	defaultJWTSecret := "test_auth"

	if !flag.Parsed() {
		serverAddressFlag := flag.String("a", defaultServerAddress, "server address")
		databaseDSNFlag := flag.String("d", defaultDatabaseDSN, "database connection string")
		accrualSystemAddressFlag := flag.String("r", defaultAccrualSystemAddress, "accrual system address")
		jwtSecretFlag := flag.String("j", defaultJWTSecret, "JWT secret key")
		flag.Parse()

		config = &Config{
			ServerAddress:        *serverAddressFlag,
			DatabaseDSN:          *databaseDSNFlag,
			AccrualSystemAddress: *accrualSystemAddressFlag,
			JWTSecret:            *jwtSecretFlag,
		}
	} else {
		config = &Config{
			ServerAddress:        defaultServerAddress,
			DatabaseDSN:          defaultDatabaseDSN,
			AccrualSystemAddress: defaultAccrualSystemAddress,
			JWTSecret:            defaultJWTSecret,
		}
	}

	envConfig := &Config{}
	if err := env.Parse(envConfig); err != nil {
		log.Printf("Failed to parse environment variables: %v", err)
	}

	if config.ServerAddress == defaultServerAddress {
		if envConfig.ServerAddress != "" {
			config.ServerAddress = envConfig.ServerAddress
		} else if envConfig.LegacyServerAddress != "" {
			config.ServerAddress = envConfig.LegacyServerAddress
		}
	}

	if config.DatabaseDSN == defaultDatabaseDSN {
		if envConfig.DatabaseDSN != "" {
			config.DatabaseDSN = envConfig.DatabaseDSN
		} else if envConfig.LegacyDatabaseDSN != "" {
			config.DatabaseDSN = envConfig.LegacyDatabaseDSN
		}
	}

	if config.AccrualSystemAddress == defaultAccrualSystemAddress && envConfig.AccrualSystemAddress != "" {
		config.AccrualSystemAddress = envConfig.AccrualSystemAddress
	}

	if config.JWTSecret == defaultJWTSecret && envConfig.JWTSecret != "" {
		config.JWTSecret = envConfig.JWTSecret
	}

	return config
}
