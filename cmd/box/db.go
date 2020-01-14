package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
)

// DBOptions contains database settings
type DBOptions struct {
	Host string `long:"host" env:"HOST" description:"hostname or IP" default:"127.0.0.1"`
	Port uint32 `long:"port" env:"PORT" description:"port" default:"5432"`
	User string `long:"user" env:"USER" description:"username"`
	Pass string `long:"pass" env:"PASS" description:"password"`
	Name string `long:"name" env:"NAME" description:"database name"`
	Opts string `long:"options" env:"OPTS" description:"extra connection options" default:"sslmode=disable"`
}

func (g *DBOptions) connStr() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?%s", g.User, g.Pass, g.Host, g.Port, g.Name, g.Opts)
}

// GetPool returns a database pool open with the provided parameters
func (g *DBOptions) GetPool() (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(g.connStr())
	if err != nil {
		return nil, err
	}
	db, err := pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		return nil, err
	}
	return db, nil
}
