package main

import (
	"context"
	"log"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/urfave/cli/v2"
	"gorm.io/gorm"

	"opentracing-playground/app"
	"opentracing-playground/database"
	versionControlMigration "opentracing-playground/database/migration/upstream"
	initMigration "opentracing-playground/database/migration/upstream/v00"
)

type config struct {
	rollback bool
}

var cfg config

var migrationOptions = gormigrate.Options{
	UseTransaction: true,
}

func initMigrate(db *gorm.DB) error {
	m := gormigrate.New(db.Debug(), &migrationOptions, []*gormigrate.Migration{})

	m.InitSchema(func(tx *gorm.DB) error {
		err := tx.AutoMigrate(initMigration.ModelSchemaList...)
		if err != nil {
			return err
		}
		return nil
	})

	err := m.Migrate()
	if err != nil {
		return err
	}
	return nil
}
func upgradeLatestMigrate(db *gorm.DB) error {
	m := gormigrate.New(db, &migrationOptions, versionControlMigration.ModelSchemaList)
	err := m.Migrate()
	if err != nil {
		return err
	}
	return nil
}

// Main starts process in cli.
func Main(ctx context.Context, c *cli.Context) {
	db := database.GetDB(database.Default).Debug()

	if cfg.rollback {
		m := gormigrate.New(db, &migrationOptions, versionControlMigration.ModelSchemaList)
		if err := m.RollbackLast(); err != nil {
			log.Fatalf("Could not RollbackLast: %v", err)
		}

		log.Print("rollback to last")
		return
	}

	if err := initMigrate(db); err != nil {
		log.Fatalf("Could not migrate: %v", err)
	}
	if err := upgradeLatestMigrate(db); err != nil {
		log.Fatalf("Could not migrate: %v", err)
	}
	lastestMigration := versionControlMigration.
		ModelSchemaList[len(versionControlMigration.ModelSchemaList)-1]
	log.Printf("updated to version \"%s\"", lastestMigration.ID)
}

func main() {
	app := app.App{
		Main: Main,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "rollback-last",
				EnvVars:     []string{"ROLLBACK_LAST"},
				Value:       false,
				Destination: &cfg.rollback,
			},
		},
	}
	app.Run()
}
