package dborm

import (
	"context"
	"fmt"
	"gorm.io/gorm/logger"
	syslog "log"
	"sync"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	ErrNotActive = fmt.Errorf("db manager is not active")
)

type Manager interface {
	GetDB(ctx context.Context) (*gorm.DB, error)
	Active() bool
}

type manager struct {
	cfg    *Config
	db     *gorm.DB
	active bool
	lock   sync.RWMutex
}

func NewManager(
	cfg *Config,
) (Manager, error) {
	m := &manager{
		cfg: cfg,
	}

	if err := m.connect(); err != nil {
		syslog.Println("[DB] connect database error: ", err)
		return m, nil
	}

	if err := m.applyPlugins(); err != nil {
		syslog.Println("[DB] apply plugins error: ", err)
		return m, nil
	}

	m.active = true
	return m, nil
}

func (m *manager) connect() (err error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	syslog.Println("[manager] connect database...")

	dsn := m.cfg.GetDSN()

	m.db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.New(
			emptyLog{},
			logger.Config{},
		),
	})

	if err != nil {
		return fmt.Errorf("open database error: %w", err)
	}

	// Setup connection pool
	sqlDB, err := m.db.DB()
	if err != nil {
		return err
	}

	sqlDB.SetMaxIdleConns(m.cfg.MaxIdleConn)
	sqlDB.SetMaxOpenConns(m.cfg.MaxOpenConn)
	sqlDB.SetConnMaxLifetime(time.Duration(m.cfg.MaxLifetime) * time.Second)

	return nil
}

func (m *manager) GetDB(ctx context.Context) (*gorm.DB, error) {
	if m == nil {
		return nil, fmt.Errorf("db manager is nil")
	}

	if !m.Active() {
		return nil, ErrNotActive
	}

	m.lock.RLock()
	defer m.lock.RUnlock()

	return m.db.WithContext(ctx), nil
}

func (m *manager) Active() bool {
	if m == nil {
		return false
	}
	return m.active
}

type emptyLog struct {
}

func (_ emptyLog) Printf(format string, args ...interface{}) {
	// ignore
}
