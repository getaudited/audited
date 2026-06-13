package domain

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound                            = errors.New("user not found")
	ErrAuthenticationFailedCredentialsMismatch = errors.New("authentication failed due to credentials mismatch")
)

type User struct {
	id        ID
	email     Email
	password  Password
	role      UserRole
	createdAt time.Time
}

func NewUser(email Email, password Password, role UserRole) (*User, error) {
	if email.Empty() {
		return nil, errors.New("email cannot be empty")
	}

	if password.Empty() {
		return nil, errors.New("password cannot be empty")
	}

	if !role.Valid() {
		return nil, errors.New("user role must be 'admin' or 'member'")
	}

	return &User{
		id:        NewID(),
		email:     email,
		password:  password,
		role:      role,
		createdAt: time.Now(),
	}, nil
}

func (u *User) ID() ID {
	return u.id
}

func (u *User) Email() Email {
	return u.email
}

func (u *User) Password() Password {
	return u.password
}

func (u *User) Role() UserRole {
	return u.role
}

func (u *User) CreatedAt() time.Time {
	return u.createdAt
}

func MarshallToUser(id, email, password, role string, createdAt time.Time) *User {
	return &User{
		id:        ID(id),
		email:     Email{email},
		password:  Password{[]byte(password)},
		role:      UserRole(role),
		createdAt: createdAt,
	}
}

type Email struct {
	value string
}

func NewEmail(email string) (Email, error) {
	_, err := mail.ParseAddress(email)
	if err != nil {
		return Email{}, fmt.Errorf("email is invalid: %v", err)
	}

	return Email{
		value: email,
	}, nil
}

func (e Email) Empty() bool {
	return strings.TrimSpace(e.value) == ""
}

func (e Email) String() string {
	return e.value
}

type Password struct {
	value []byte
}

func NewPassword(plainTextPassword string) (Password, error) {
	// TODO: extract bcrypt out of the doamin
	hash, err := bcrypt.GenerateFromPassword([]byte(plainTextPassword), bcrypt.DefaultCost)
	if err != nil {
		return Password{}, fmt.Errorf("error generating password: %w", err)
	}

	return Password{
		value: hash,
	}, nil
}

func (p Password) String() string {
	return string(p.value)
}

func (p Password) Empty() bool {
	return string(p.value) == ""
}

func (p Password) IsEqual(plainTextPassword string) bool {
	// TODO: extract bcrypt out of the domain
	err := bcrypt.CompareHashAndPassword(p.value, []byte(plainTextPassword))
	return err == nil
}

type UserRole string

func (r UserRole) Valid() bool {
	return r == UserRoleAdmin || r == UserRoleMember
}

func (r UserRole) String() string {
	return string(r)
}

var (
	UserRoleAdmin  = UserRole("admin")
	UserRoleMember = UserRole("member")
)

type UserRepository interface {
	FindByEmail(ctx context.Context, email Email) (*User, error)
	FindByID(ctx context.Context, id ID) (*User, error)
	Save(ctx context.Context, user *User) error
}
