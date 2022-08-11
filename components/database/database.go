package database

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"time"
)

const (
	maxOpenConnections = 60
	connMaxLifetime    = 120
	maxIdleConnections = 30
	connMaxIdleTime    = 20
)

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	Dbname   string
	SSLMode  string
	Driver   string
	CreateDb bool
}

func (c *Config) toPgConnection() string {
	dataSourceName := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		c.Host,
		c.Port,
		c.User,
		c.Dbname,
		c.Password,
		c.SSLMode,
	)
	return dataSourceName
}
func (c *Config) NewComponent() (*sqlx.DB, error) {
	db, err := sqlx.Connect(c.Driver, c.toPgConnection())
	if err != nil {
		return nil, errors.Wrap(err, "Database.Connect")
	}

	db.SetMaxOpenConns(maxOpenConnections)
	db.SetConnMaxLifetime(connMaxLifetime * time.Second)
	db.SetMaxIdleConns(maxIdleConnections)
	db.SetConnMaxIdleTime(connMaxIdleTime * time.Second)
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, err
}
