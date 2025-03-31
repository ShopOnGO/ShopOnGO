package user

import (
	"errors"
)

type UserService struct {
	UserRepository *UserRepository
}

func NewUserService(userRepository *UserRepository) *UserService {
	return &UserService{UserRepository: userRepository}
}

func (s *UserService) CreateUser(user *User) (*User, error) {
	if user.Email == "" {
		return nil, errors.New("email is required")
	}
	return s.UserRepository.Create(user)
}

func (s *UserService) FindUserByEmail(email string) (*User, error) {
	if email == "" {
		return nil, errors.New("email cannot be empty")
	}
	return s.UserRepository.FindByEmail(email)
}

func (s *UserService) UpdateUser(user *User) (*User, error) {
	if user.ID == 0 {
		return nil, errors.New("user ID is required for update")
	}
	return s.UserRepository.Update(user)
}

func (s *UserService) DeleteUser(id uint) error {
	if id == 0 {
		return errors.New("user ID is required for deletion")
	}
	return s.UserRepository.Delete(id)
}
