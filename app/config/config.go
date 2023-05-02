package config

import (
	"fmt"
	"os"
)

type AppConfig struct {
	AppInfo *AppInfo
}

type AppInfo struct {
	DatabaseURL string
}

func loadDatabaseURL(dbName string) (string, error) {
	mysqlHost := os.Getenv("MYSQL_HOST")
	mysqlUser := os.Getenv("MYSQL_USER")
	mysqlPassword := os.Getenv("MYSQL_PASSWORD")
	mysqlPort := os.Getenv("MYSQL_PORT")
	if mysqlHost == "" || mysqlUser == "" || mysqlPassword == "" || mysqlPort == "" {
		return "", fmt.Errorf("環境変数が不足しています。MYSQL_HOST: %s, MYSQL_USER: %s, MYSQL_PASSWORD: %s, MYSQL_PORT: %s", mysqlHost, mysqlUser, mysqlPassword, mysqlPort)
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true", mysqlUser, mysqlPassword, mysqlHost, mysqlPort, dbName), nil
}

func LoadConfig() (*AppConfig, error) {
	databaseURL, err := loadDatabaseURL(os.Getenv("MYSQL_DATABASE"))
	if err != nil {
		return nil, err
	}

	appInfo := &AppInfo{
		DatabaseURL: databaseURL,
	}

	config := AppConfig{
		AppInfo: appInfo,
	}

	return &config, nil
}

func LoadTestConfig() (*AppConfig, error) {
	databaseURL, err := loadDatabaseURL(os.Getenv("MYSQL_TEST_DATABASE"))
	if err != nil {
		return nil, err
	}

	appInfo := &AppInfo{
		DatabaseURL: databaseURL,
	}

	config := AppConfig{
		AppInfo: appInfo,
	}

	return &config, nil
}
