package services

import (
	"errors"

	"github.com/sakkurohilla/kineticops/backend/internal/auth"
	"github.com/sakkurohilla/kineticops/backend/internal/models"
	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
	"gorm.io/gorm"
)

// GetUserByID retrieves a user by their ID
func GetUserByID(userID int64) (*models.User, error) {
	var user models.User
	result := postgres.DB.First(&user, userID)

	if result.Error == gorm.ErrRecordNotFound {
		return nil, errors.New("user not found")
	} else if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

// GetUserByEmail retrieves a user by email
func GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	result := postgres.DB.Where("email = ?", email).First(&user)

	if result.Error == gorm.ErrRecordNotFound {
		return nil, errors.New("user not found")
	} else if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

// GetUserByUsername retrieves a user by username
func GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	result := postgres.DB.Where("username = ?", username).First(&user)

	if result.Error == gorm.ErrRecordNotFound {
		return nil, errors.New("user not found")
	} else if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

// GetAllUsers retrieves all users (admin only)
func GetAllUsers() ([]models.User, error) {
	var users []models.User
	result := postgres.DB.Find(&users)

	if result.Error != nil {
		return nil, result.Error
	}

	return users, nil
}

// UpdateUserProfile updates user's email and username
func UpdateUserProfile(userID int64, email, username string) error {
	// Check if email or username already taken by another user
	var count int64
	postgres.DB.Model(&models.User{}).
		Where("(email = ? OR username = ?) AND id != ?", email, username, userID).
		Count(&count)

	if count > 0 {
		return errors.New("email or username already taken")
	}

	updates := map[string]interface{}{}
	if email != "" {
		updates["email"] = email
	}
	if username != "" {
		updates["username"] = username
	}

	result := postgres.DB.Model(&models.User{}).Where("id = ?", userID).Updates(updates)
	return result.Error
}

// ProfileUpdateRequest represents extended profile update data
type ProfileUpdateRequest struct {
	Email     string `json:"email"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
	Company   string `json:"company"`
	Location  string `json:"location"`
	Role      string `json:"role"`
	Timezone  string `json:"timezone"`
}

// UpdateUserProfileExtended updates user's full profile
func UpdateUserProfileExtended(userID int64, req ProfileUpdateRequest) error {
	// Check if email or username already taken by another user
	if req.Email != "" || req.Username != "" {
		var count int64
		query := postgres.DB.Model(&models.User{}).Where("id != ?", userID)

		if req.Email != "" && req.Username != "" {
			query = query.Where("email = ? OR username = ?", req.Email, req.Username)
		} else if req.Email != "" {
			query = query.Where("email = ?", req.Email)
		} else if req.Username != "" {
			query = query.Where("username = ?", req.Username)
		}

		query.Count(&count)
		if count > 0 {
			return errors.New("email or username already taken")
		}
	}

	updates := map[string]interface{}{}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Username != "" {
		updates["username"] = req.Username
	}
	if req.FirstName != "" {
		updates["first_name"] = req.FirstName
	}
	if req.LastName != "" {
		updates["last_name"] = req.LastName
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}
	if req.Company != "" {
		updates["company"] = req.Company
	}
	if req.Location != "" {
		updates["location"] = req.Location
	}
	if req.Role != "" {
		updates["role"] = req.Role
	}
	if req.Timezone != "" {
		updates["timezone"] = req.Timezone
	}

	if len(updates) == 0 {
		return errors.New("no fields to update")
	}

	result := postgres.DB.Model(&models.User{}).Where("id = ?", userID).Updates(updates)
	return result.Error
}

// DeleteUser soft deletes a user
func DeleteUser(userID int64) error {
	result := postgres.DB.Delete(&models.User{}, userID)
	return result.Error
}

// ChangeUserPassword changes user's password after verifying current password
func ChangeUserPassword(userID int64, currentPassword, newPassword string) error {
	user, err := GetUserByID(userID)
	if err != nil {
		return err
	}

	// Verify current password
	if !auth.CheckPasswordHash(currentPassword, user.PasswordHash) {
		return errors.New("current password is incorrect")
	}

	// Hash new password
	hash, err := auth.HashPassword(newPassword)
	if err != nil {
		return err
	}

	// Update password
	result := postgres.DB.Model(&models.User{}).
		Where("id = ?", userID).
		Update("password_hash", hash)

	return result.Error
}
