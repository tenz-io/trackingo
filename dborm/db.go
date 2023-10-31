package dborm

import (
	"context"
	"fmt"
	syslog "log"
	"sync"
	"time"

	"github.com/tenz-io/trackingo/util"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	ErrNotActive = fmt.Errorf("db manager is not active")
)

type Manager interface {
	GetDB(ctx context.Context) (*gorm.DB, error)
	Active() bool
}

type manager struct {
	cfg      *Config
	db       *gorm.DB
	dbLogger logger.Writer // log.Logger
	active   bool
	lock     sync.RWMutex
}

func NewManager(
	cfg *Config,
) (Manager, error) {
	m := &manager{
		cfg:      cfg,
		dbLogger: newDBLog(cfg.TrackingLogbase),
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
		Logger: logger.New(m.dbLogger, logger.Config{
			SlowThreshold:             time.Second,
			Colorful:                  false,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      false,
			LogLevel:                  util.If(m.cfg.EnableTracking, logger.Info, logger.Warn),
		}),
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
