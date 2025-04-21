# pgconnect

A lightweight Go package for PostgreSQL database connectivity using GORM and designed for easy integration with Gin applications.

## Features

- Simple connection configuration with sensible defaults
- Connection pooling management
- Transaction support
- Generic repository pattern for common database operations
- Easy integration with GORM

## Installation

```bash
go get github.com/JorgeSaicoski/pgconnect
```

## Usage

### Basic Connection

```go
package main

import (
	"log"

	"github.com/JorgeSaicoski/pgconnect"
)

func main() {
	// Use default configuration
	config := pgconnect.DefaultConfig()
	
	// Or customize as needed
	config.Host = "localhost"
	config.Port = "5432"
	config.User = "postgres"
	config.Password = "yourpassword"
	config.DatabaseName = "yourdb"
	
	db, err := pgconnect.New(config)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	
	// Don't forget to close the connection when done
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()
	
	// Check connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Database ping failed: %v", err)
	}
}
```

### Working with Models

```go
type User struct {
	ID       uint   `gorm:"primaryKey"`
	Name     string `gorm:"size:255;not null"`
	Email    string `gorm:"size:255;uniqueIndex"`
	Password string `gorm:"size:255;not null"`
}

// Auto migrate your models
if err := db.AutoMigrate(&User{}); err != nil {
	log.Fatalf("Failed to auto migrate: %v", err)
}
```

### Using the Repository Pattern

```go
type UserService struct {
	repo *pgconnect.Repository[User]
}

func NewUserService(db *pgconnect.DB) *UserService {
	return &UserService{
		repo: pgconnect.NewRepository[User](db),
	}
}

func (s *UserService) CreateUser(user *User) error {
	return s.repo.Create(user)
}

func (s *UserService) GetUserByID(id uint) (*User, error) {
	var user User
	if err := s.repo.FindByID(id, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserService) GetAllUsers() ([]User, error) {
	var users []User
	if err := s.repo.FindAll(&users); err != nil {
		return nil, err
	}
	return users, nil
}
```

### Using Transactions

```go
err := db.WithTransaction(func(tx *gorm.DB) error {
	// Perform multiple operations within a transaction
	user := User{Name: "John", Email: "john@example.com"}
	if err := tx.Create(&user).Error; err != nil {
		return err // Transaction will be rolled back
	}
	
	profile := Profile{UserID: user.ID, Bio: "Bio info"}
	if err := tx.Create(&profile).Error; err != nil {
		return err // Transaction will be rolled back
	}
	
	return nil // Transaction will be committed
})
```

## License

[MIT](LICENSE)
