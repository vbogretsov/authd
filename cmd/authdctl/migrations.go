package main

import (
	"fmt"

	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"

	_ "github.com/jinzhu/gorm/dialects/mssql"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"

	"github.com/vbogretsov/authd/model"
)

var migrations = []*gormigrate.Migration{
	{
		ID: "0.1.0",
		Migrate: func(tx *gorm.DB) error {
			var err error = nil

			err = tx.AutoMigrate(
				&model.User{},
				&model.Confirmation{},
				&model.Refresh{}).Error
			if err != nil {
				return err
			}

			err = addUserFK(tx, &model.Confirmation{})
			if err != nil {
				return err
			}

			err = addUserFK(tx, &model.Refresh{})
			if err != nil {
				return err
			}

			return err
		},
		Rollback: func(tx *gorm.DB) error {
			var err error = nil

			err = tx.DropTable(tabname(tx, &model.Refresh{})).Error
			if err != nil {
				return err
			}

			err = tx.DropTable(tabname(tx, &model.Confirmation{})).Error
			if err != nil {
				return err
			}

			err = tx.DropTable(tabname(tx, &model.User{})).Error
			if err != nil {
				return err
			}

			return err
		},
	},
}

func addUserFK(tx *gorm.DB, v interface{}) error {
	userFK := fmt.Sprintf("%s(id)", tabname(tx, &model.User{}))
	userID := "user_id"

	return tx.Model(&model.Confirmation{}).
		AddForeignKey(userID, userFK, "CASCADE", "RESTRICT").Error
}

func tabname(tx *gorm.DB, v interface{}) string {
	return tx.NewScope(v).TableName()
}
