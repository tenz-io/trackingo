package dborm

import (
	"context"
	"fmt"
	"github.com/tenz-io/trackingo/common"
	"github.com/tenz-io/trackingo/logger"
	"github.com/tenz-io/trackingo/monitor"
	"gorm.io/gorm"
	"time"
)

type recordCtxKeyType string

const (
	metricsRecordCtxKey recordCtxKeyType = "_metrics_record_ctx_key"
	beginTimeCtxKey     recordCtxKeyType = "_begin_time_record_ctx_key"
)

func (m *manager) applyPlugins() (err error) {
	err = m.db.Callback().Query().Before("*").Register("start_query_metrics", m.enter("db_query"))
	if err != nil {
		return fmt.Errorf("register start_metrics error: %w", err)
	}

	err = m.db.Callback().Create().Before("*").Register("start_create_metrics", m.enter("db_create"))
	if err != nil {
		return fmt.Errorf("register start_metrics error: %w", err)
	}

	err = m.db.Callback().Update().Before("*").Register("start_update_metrics", m.enter("db_update"))
	if err != nil {
		return fmt.Errorf("register start_metrics error: %w", err)
	}

	err = m.db.Callback().Delete().Before("*").Register("start_delete_metrics", m.enter("db_delete"))
	if err != nil {
		return fmt.Errorf("register start_metrics error: %w", err)
	}

	err = m.db.Callback().Row().Before("*").Register("start_row_metrics", m.enter("db_row"))
	if err != nil {
		return fmt.Errorf("register start_metrics error: %w", err)
	}

	err = m.db.Callback().Raw().Before("*").Register("start_raw_metrics", m.enter("db_raw"))
	if err != nil {
		return fmt.Errorf("register start_metrics error: %w", err)
	}

	err = m.db.Callback().Query().After("*").Register("end_query_metrics", m.exit("db_query"))
	if err != nil {
		return fmt.Errorf("register end_metrics error: %w", err)
	}

	err = m.db.Callback().Create().After("*").Register("end_create_metrics", m.exit("db_create"))
	if err != nil {
		return fmt.Errorf("register end_metrics error: %w", err)
	}

	err = m.db.Callback().Update().After("*").Register("end_update_metrics", m.exit("db_update"))
	if err != nil {
		return fmt.Errorf("register end_metrics error: %w", err)
	}

	err = m.db.Callback().Delete().After("*").Register("end_delete_metrics", m.exit("db_delete"))
	if err != nil {
		return fmt.Errorf("register end_metrics error: %w", err)
	}

	return nil
}

// enter is a callback function that will be called when the gorm
func (m *manager) enter(dsCmd string) func(db *gorm.DB) {

	if !m.cfg.EnableTracking {
		return func(db *gorm.DB) {}
	}

	return func(db *gorm.DB) {
		ctx := db.Statement.Context
		beginTime := time.Now()
		ctx = context.WithValue(ctx, beginTimeCtxKey, beginTime)
		rec := monitor.BeginRecord(ctx, dsCmd)
		ctx = context.WithValue(ctx, metricsRecordCtxKey, rec)
		db.Statement.Context = ctx
		logger.TrafficEntryFromContext(ctx).DataWith(&logger.Traffic{
			Typ: logger.TrafficTypRequest,
			Cmd: dsCmd,
		}, logger.Fields{
			"sql": db.Statement.SQL.String(),
			"val": db.Statement.Vars,
		})
	}
}

// exit is a callback function that will be called when the gorm
func (m *manager) exit(dsCmd string) func(db *gorm.DB) {
	if !m.cfg.EnableTracking {
		return func(db *gorm.DB) {}
	}

	return func(db *gorm.DB) {
		ctx := db.Statement.Context
		rec, ok := ctx.Value(metricsRecordCtxKey).(*monitor.Recorder)
		if ok {
			rec.EndWithError(db.Error)
		}

		beginTime, ok := ctx.Value(beginTimeCtxKey).(time.Time)
		if ok {
			logger.TrafficEntryFromContext(ctx).DataWith(&logger.Traffic{
				Typ:  logger.TrafficTypRequestResp,
				Cmd:  dsCmd,
				Code: common.ErrorCode(db.Error),
				Msg:  common.ErrorMsg(db.Error),
				Cost: time.Since(beginTime),
			}, logger.Fields{
				"sql": db.Statement.SQL.String(),
				"val": db.Statement.Vars,
			})

		}

	}

}
