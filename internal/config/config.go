package config

import (
    "fmt"
    "os"
    "time"

    "gopkg.in/yaml.v3"
)

type Config struct {
    Server ServerConfig `yaml:"server"`
    JWT    JWTConfig    `yaml:"jwt"`
    DB     DBConfig     `yaml:"db"`
}

type ServerConfig struct {
    HTTPPort int `yaml:"http_port"`
}

type JWTConfig struct {
    Secret string `yaml:"secret"`
    TTL    string `yaml:"ttl"`

    ParsedTTL time.Duration `yaml:"-"`
}

type DBConfig struct {
    ConnString string `yaml:"conn_string"`
    MaxConns   int32  `yaml:"max_conns"`
}

func Load(path string) (*Config, error) {
    b, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

    var cfg Config
    if err := yaml.Unmarshal(b, &cfg); err != nil {
        return nil, err
    }

    if cfg.Server.HTTPPort == 0 {
        cfg.Server.HTTPPort = 8080
    }

    if cfg.JWT.Secret == "" {
        return nil, fmt.Errorf("jwt.secret is required")
    }

    if cfg.JWT.TTL == "" {
        cfg.JWT.TTL = "24h"
    }

    ttl, err := time.ParseDuration(cfg.JWT.TTL)
    if err != nil {
        return nil, fmt.Errorf("invalid jwt.ttl: %w", err)
    }
    cfg.JWT.ParsedTTL = ttl

    if cfg.DB.ConnString == "" {
        return nil, fmt.Errorf("db.conn_string is required")
    }
    if cfg.DB.MaxConns == 0 {
        cfg.DB.MaxConns = 10
    }

    return &cfg, nil
}
