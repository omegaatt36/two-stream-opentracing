package user

import (
	"opentracing-playground/models"
)

// User defines user's db model.
type User struct {
	models.Model
	Email    string `gorm:"type:varchar(256);not null;unique"`
	Password string `gorm:"type:varchar(256);not null"`
	Name     string `gorm:"type:varchar(128);default:'';not null"`
}
