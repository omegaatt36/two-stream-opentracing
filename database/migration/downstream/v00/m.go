package v00

import (
	"opentracing-playground/models"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// User defines user's db model.
type User struct {
	models.Model
	Email    string `gorm:"type:varchar(256);not null;unique" json:"email"`
	Password string `gorm:"type:varchar(256);not null" json:"-"`
	Name     string `gorm:"type:varchar(128);default:'';not null" json:"name"`
}

// Init add user table.
var Init = gormigrate.Migration{
	ID: "add-user",
	Migrate: func(tx *gorm.DB) error {
		return tx.AutoMigrate(&User{})
	},
	Rollback: func(tx *gorm.DB) error {
		return nil
	},
}

// ModelSchemaList v0 Model Structs
var ModelSchemaList = []any{
	&User{},
}
