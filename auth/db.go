package auth

import (
	"github.com/jinzhu/gorm"

	"github.com/vbogretsov/authd/model"
)

func findUser(tx *gorm.DB, email string) (*model.User, error) {
	user := model.User{Email: email}

	res := tx.Where("email = ?", email).First(&user)
	if res.RecordNotFound() {
		return nil, nil
	}

	if res.Error != nil {
		return nil, res.Error
	}

	return &user, nil
}

func atomic(db *gorm.DB, action func(*gorm.DB) error) error {
	txn := db.Begin()

	err := txn.Error
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			txn.Rollback()
		}
	}()

	err = action(txn)
	if err != nil {
		return err
	}

	return txn.Commit().Error
}
