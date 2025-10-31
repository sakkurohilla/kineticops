package postgres

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/sakkurohilla/kineticops/backend/internal/models"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	_ "github.com/lib/pq"
)

var DB *gorm.DB
var SqlxDB *sqlx.DB

func Init() error {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		viper.GetString("POSTGRES_HOST"),
		viper.GetString("POSTGRES_PORT"),
		viper.GetString("POSTGRES_USER"),
		viper.GetString("POSTGRES_PASSWORD"),
		viper.GetString("POSTGRES_DB"),
	)
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn), // minimize log noise
	})
	if err != nil {
		log.Printf("Postgres connection failed: %v", err)
		return err
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
	if migrateErr := DB.AutoMigrate(&models.Host{}, &models.Metric{}, &models.HostMetric{}, &models.Alert{}, &models.AlertRule{}); migrateErr != nil {
		log.Printf("Postgres auto-migrate warning: %v", migrateErr)
		// do not fail startup on migrate error, but return the original err (nil or previous)
	} else {
		// Check which tables were created or already existed
		nowHost := DB.Migrator().HasTable(&models.Host{})
		nowMetric := DB.Migrator().HasTable(&models.Metric{})
		nowHostMetric := DB.Migrator().HasTable(&models.HostMetric{})
		nowAlert := DB.Migrator().HasTable(&models.Alert{})
		nowAlertRule := DB.Migrator().HasTable(&models.AlertRule{})

		if !hasHost && nowHost {
			log.Println("Postgres auto-migrate: created table 'hosts'")
		} else if nowHost {
			log.Println("Postgres: table 'hosts' exists")
		}

		if !hasMetric && nowMetric {
			log.Println("Postgres auto-migrate: created table 'metrics'")
		} else if nowMetric {
			log.Println("Postgres: table 'metrics' exists")
		}

		if !hasHostMetric && nowHostMetric {
			log.Println("Postgres auto-migrate: created table 'host_metrics'")
		} else if nowHostMetric {
			log.Println("Postgres: table 'host_metrics' exists")
		}

		if !hasAlert && nowAlert {
			log.Println("Postgres auto-migrate: created table 'alerts'")
		} else if nowAlert {
			log.Println("Postgres: table 'alerts' exists")
		}

		if !hasAlertRule && nowAlertRule {
			log.Println("Postgres auto-migrate: created table 'alert_rules'")
		} else if nowAlertRule {
			log.Println("Postgres: table 'alert_rules' exists")
		}
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
		log.Printf("Sqlx connection failed: %v", err)
		return err
	}
	
	log.Println("PostgreSQL connected (GORM + sqlx)")
	return err
}
