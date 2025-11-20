package services

import (
	"errors"

	"github.com/sakkurohilla/kineticops/backend/internal/models"
	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
	"gorm.io/gorm"
)

// GetUserSettings retrieves user settings, creates default if not exists
func GetUserSettings(userID int64) (*models.UserSettings, error) {
	var settings models.UserSettings
	result := postgres.DB.Where("user_id = ?", userID).First(&settings)

	if result.Error == gorm.ErrRecordNotFound {
		// Create default settings
		settings = models.UserSettings{
			UserID:               userID,
			CompanyName:          "KineticOps",
			Timezone:             "Asia/Kolkata",
			DateFormat:           "YYYY-MM-DD",
			EmailNotifications:   true,
			SlackNotifications:   false,
			WebhookNotifications: false,
			AlertEmail:           "",
			SlackWebhook:         "",
			CustomWebhook:        "",
			RequireMFA:           false,
			SessionTimeout:       30,
			PasswordExpiry:       90,
			MetricsRetention:     30,
			LogsRetention:        7,
			TracesRetention:      7,
			AutoRefresh:          true,
			RefreshInterval:      30,
			MaxDashboardWidgets:  20,
		}

		if err := postgres.DB.Create(&settings).Error; err != nil {
			return nil, err
		}

		return &settings, nil
	} else if result.Error != nil {
		return nil, result.Error
	}

	return &settings, nil
}

// UpdateUserSettings updates user settings
func UpdateUserSettings(userID int64, updates map[string]interface{}) error {
	// Check if settings exist
	var settings models.UserSettings
	result := postgres.DB.Where("user_id = ?", userID).First(&settings)

	if result.Error == gorm.ErrRecordNotFound {
		// Create new settings with provided values
		settings = models.UserSettings{UserID: userID}
		if err := postgres.DB.Create(&settings).Error; err != nil {
			return err
		}
	} else if result.Error != nil {
		return result.Error
	}

	// Update settings
	if len(updates) == 0 {
		return errors.New("no fields to update")
	}

	result = postgres.DB.Model(&models.UserSettings{}).
		Where("user_id = ?", userID).
		Updates(updates)

	return result.Error
}
