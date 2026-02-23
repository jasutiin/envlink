package auth

import (
	"fmt"

	"gorm.io/gorm"
)

type AuthRepository interface {
	Login(email, password string) (uint, error)
	Register(email, password string) (uint, error)
}

type authRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) AuthRepository {
	return &authRepository{db: db}
}

func (repository *authRepository) Login(email, password string) (uint, error) {
	_ = repository.db
	_ = email
	_ = password
	return 0, nil
}

func (repository *authRepository) Register(email, password string) (uint, error) {
	db := repository.db
	user := User{Email: email, Password: password}
	result := db.Create(&user)

	if result.Error != nil {
		fmt.Println("there was an error with creating a new user")
	}

	return user.ID, nil
}