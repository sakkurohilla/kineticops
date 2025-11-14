package postgres

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/sakkurohilla/kineticops/backend/internal/logging"
	"github.com/sakkurohilla/kineticops/backend/internal/models"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB
var SqlxDB *sqlx.DB

func Init() error {
	// Build DSN using environment variables (viper reads env and .env files)
	host := viper.GetString("POSTGRES_HOST")
	if host == "" {
		host = viper.GetString("POSTGRES_HOST")
	}
	port := viper.GetString("POSTGRES_PORT")
	user := viper.GetString("POSTGRES_USER")
	pass := viper.GetString("POSTGRES_PASSWORD")
	dbname := viper.GetString("POSTGRES_DB")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host,
		port,
		user,
		pass,
		dbname,
	)
	var err error
	// Try up to 3 times to connect (helps with transient startup ordering when DB comes up).
	var lastErr error
	for i := 0; i < 3; i++ {
		DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Warn), // minimize log noise
		})
		if err == nil {
			lastErr = nil
			break
		}
		lastErr = err
		logging.Warnf("Postgres connection attempt %d failed: %v", i+1, err)
		time.Sleep(time.Duration(i+1) * time.Second)
	}
	if lastErr != nil {
		// Provide guidance to operator: echo masked connection details
		logging.Errorf("Postgres connection failed after retries: %v", lastErr)
		logging.Errorf("Postgres DSN host=%s port=%s user=%s db=%s", host, port, user, dbname)
		return lastErr
	}

	sqlDB, err := DB.DB()
	if err == nil {
		sqlDB.SetMaxOpenConns(20)
		sqlDB.SetMaxIdleConns(10)
	}
	// Ensure essential tables exist (auto-migrate minimal schemas).
	// We'll check existence before and after AutoMigrate so we can log what changed.
	hasHost := DB.Migrator().HasTable(&models.Host{})
	hasMetric := DB.Migrator().HasTable(&models.Metric{})
	hasHostMetric := DB.Migrator().HasTable(&models.HostMetric{})
	hasAlert := DB.Migrator().HasTable(&models.Alert{})
	hasAlertRule := DB.Migrator().HasTable(&models.AlertRule{})

	// Auto-migrate core models including alerts/rules so API endpoints don't 500
	// when those tables are missing in the target database.
	// Run safe idempotent adjustments before AutoMigrate to avoid noisy errors
	// when GORM attempts to change constraints. Specifically ensure the old
	// unique constraint on host_metrics is dropped if present.
	if err := DB.Exec(`ALTER TABLE host_metrics DROP CONSTRAINT IF EXISTS "uni_host_metrics_host_id";`).Error; err != nil {
		// Do not fail startup for this; log at debug level via Printf
		logging.Warnf("Postgres: safe drop constraint attempt returned: %v", err)
	}

	// Note: host_metrics schema is managed via SQL migrations to avoid
	// destructive changes that GORM may attempt (it can issue DROP CONSTRAINT
	// without IF EXISTS). Exclude HostMetric from AutoMigrate and let the
	// SQL migration files handle any constraint changes safely.
	// NOTE: GORM AutoMigrate is intentionally disabled here to avoid it
	// issuing DDL that may be destructive or non-idempotent in existing
	// deployments (for example, DROP CONSTRAINT without IF EXISTS). The
	// project uses explicit SQL migration files under backend/migrations
	// to manage schema changes safely. We'll only verify that expected
	// tables exist and log their presence.

	nowHost := DB.Migrator().HasTable(&models.Host{})
	nowMetric := DB.Migrator().HasTable(&models.Metric{})
	nowHostMetric := DB.Migrator().HasTable(&models.HostMetric{})
	nowAlert := DB.Migrator().HasTable(&models.Alert{})
	nowAlertRule := DB.Migrator().HasTable(&models.AlertRule{})

	if !hasHost && nowHost {
		logging.Infof("Postgres: table 'hosts' was created (migration tool should have run)")
	} else if nowHost {
		logging.Infof("Postgres: table 'hosts' exists")
	}

	if !hasMetric && nowMetric {
		logging.Infof("Postgres: table 'metrics' was created (migration tool should have run)")
	} else if nowMetric {
		logging.Infof("Postgres: table 'metrics' exists")
	}

	if !hasHostMetric && nowHostMetric {
		logging.Infof("Postgres: table 'host_metrics' was created (migration tool should have run)")
	} else if nowHostMetric {
		logging.Infof("Postgres: table 'host_metrics' exists")
	}

	if !hasAlert && nowAlert {
		logging.Infof("Postgres: table 'alerts' was created (migration tool should have run)")
	} else if nowAlert {
		logging.Infof("Postgres: table 'alerts' exists")
	}

	if !hasAlertRule && nowAlertRule {
		logging.Infof("Postgres: table 'alert_rules' was created (migration tool should have run)")
	} else if nowAlertRule {
		logging.Infof("Postgres: table 'alert_rules' exists")
	}

	// Initialize sqlx connection for agent service
	sqlxDSN := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		viper.GetString("POSTGRES_HOST"),
		viper.GetString("POSTGRES_PORT"),
		viper.GetString("POSTGRES_USER"),
		viper.GetString("POSTGRES_PASSWORD"),
		viper.GetString("POSTGRES_DB"),
	)

	SqlxDB, err = sqlx.Connect("postgres", sqlxDSN)
	if err != nil {
		logging.Errorf("Sqlx connection failed: %v", err)
		return err
	}

	logging.Infof("PostgreSQL connected (GORM + sqlx)")
	return err
}
