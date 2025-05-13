# How to Use pgconnect

This guide provides detailed explanations and examples for using all the functionality provided by the pgconnect library.

## Table of Contents

1. [Configuration](#configuration)
2. [Connection Management](#connection-management)
3. [Transaction Support](#transaction-support)
4. [Model Migration](#model-migration)
5. [Repository Pattern](#repository-pattern)
6. [Pagination](#pagination)
7. [Integration with Gin](#integration-with-gin)
8. [Complete Examples](#complete-examples)

## Configuration

The `Config` struct holds all the necessary configuration parameters for your PostgreSQL database connection.

```go
// Create config with defaults
config := pgconnect.DefaultConfig()

// Customize as needed
config.Host = "postgres-server"        // Default: "localhost"
config.Port = "5432"                   // Default: "5432"
config.User = "appuser"                // Default: "postgres"
config.Password = "securepassword"     // Default: "postgres"
config.DatabaseName = "myapp"          // Default: "postgres"
config.SSLMode = "disable"             // Default: "disable"
config.TimeZone = "UTC"                // Default: "UTC"
config.MaxIdleConns = 10               // Default: 10
config.MaxOpenConns = 100              // Default: 100
config.LogLevel = logger.Info          // Default: logger.Silent
```

Available log levels:
- `logger.Silent` - No logging
- `logger.Error` - Log errors only
- `logger.Warn` - Log warnings and errors
- `logger.Info` - Log info, warnings, and errors

## Connection Management

### Creating a Connection

```go
// Create a new database connection
db, err := pgconnect.New(config)
if err != nil {
    log.Fatalf("Failed to connect to database: %v", err)
}
```

### Connection Checking

```go
// Check if connection is alive
if err := db.Ping(); err != nil {
    log.Fatalf("Database connection failed: %v", err)
}
```

### Closing Connections

```go
// Close the database connection when done
if err := db.Close(); err != nil {
    log.Printf("Error closing database connection: %v", err)
}
```

### Connection Pooling

The connection pool is automatically configured based on the `MaxIdleConns` and `MaxOpenConns` settings in your Config. The library handles connection pooling internally through GORM.

## Transaction Support

Use the `WithTransaction` method to perform multiple operations within a transaction:

```go
err := db.WithTransaction(func(tx *gorm.DB) error {
    // Operation 1
    user := User{Name: "John"}
    if err := tx.Create(&user).Error; err != nil {
        return err // Will cause rollback
    }
    
    // Operation 2
    profile := Profile{UserID: user.ID}
    if err := tx.Create(&profile).Error; err != nil {
        return err // Will cause rollback
    }
    
    return nil // Will commit the transaction
})

if err != nil {
    log.Printf("Transaction failed: %v", err)
}
```

## Model Migration

Run database migrations for your models:

```go
// Define your models
type User struct {
    gorm.Model
    Name     string
    Email    string `gorm:"uniqueIndex"`
    Password string
}

type Profile struct {
    gorm.Model
    UserID  uint
    Bio     string
    Avatar  string
}

// Auto migrate models
if err := db.AutoMigrate(&User{}, &Profile{}); err != nil {
    log.Fatalf("Failed to migrate database: %v", err)
}
```

## Repository Pattern

The Repository pattern provides a generic way to handle common database operations for your models.

### Creating a Repository

```go
// Create a repository for User model
userRepo := pgconnect.NewRepository[User](db)

// Create a repository for Profile model
profileRepo := pgconnect.NewRepository[Profile](db)
```

### Basic CRUD Operations

```go
// Create a new user
user := User{Name: "John", Email: "john@example.com"}
if err := userRepo.Create(&user); err != nil {
    log.Printf("Failed to create user: %v", err)
}

// Find user by ID
var foundUser User
if err := userRepo.FindByID(user.ID, &foundUser); err != nil {
    log.Printf("User not found: %v", err)
}

// Update user
foundUser.Name = "John Smith"
if err := userRepo.Update(&foundUser); err != nil {
    log.Printf("Failed to update user: %v", err)
}

// Delete user
if err := userRepo.Delete(&foundUser); err != nil {
    log.Printf("Failed to delete user: %v", err)
}

// Delete users by condition
if err := userRepo.DeleteWhere("age > ?", 30); err != nil {
    log.Printf("Failed to delete users over 30: %v", err)
}

```

### Advanced Queries

```go
// Find all users
var allUsers []User
if err := userRepo.FindAll(&allUsers); err != nil {
    log.Printf("Failed to fetch users: %v", err)
}

// Find users with specific condition
var activeUsers []User
if err := userRepo.FindWhere(&activeUsers, "status = ?", "active"); err != nil {
    log.Printf("Failed to fetch active users: %v", err)
}

// Find one user with condition
var admin User
if err := userRepo.FindOne(&admin, "role = ?", "admin"); err != nil {
    log.Printf("Admin not found: %v", err)
}

// Count records
var count int64
if err := userRepo.Count(&count, "status = ?", "active"); err != nil {
    log.Printf("Failed to count users: %v", err)
}

```

## Pagination

Implement pagination for your data retrieval:

```go
// Get users with pagination (page 1, 10 items per page)
var users []User
if err := userRepo.Paginate(&users, 1, 10); err != nil {
    log.Printf("Failed to paginate users: %v", err)
}

// Custom pagination with conditions
page := 2
pageSize := 15
var activeUsers []User

//  Using the repository's Paginate method
if err := userRepo.Paginate(&activeUsers, page, pageSize); err != nil {
    log.Printf("Pagination failed: %v", err)
}

// PaginateWhere combines filtering conditions with pagination for efficient retrieval
// of subsets of data. It applies WHERE conditions before pagination to avoid 
// retrieving all records when only specific ones are needed.

// Using the repository's PaginateWhere method
if err := userRepo.PaginateWhere(&activeUsers, page, pageSize, "status = ?", "active"); err != nil {
    log.Printf("Pagination with filters failed: %v", err)
}

// Find non-completed tasks with pagination
var pendingTasks []Task
if err := taskRepo.PaginateWhere(&pendingTasks, page, pageSize, "status != ?", "completed"); err != nil {
    log.Printf("Failed to paginate pending tasks: %v", err)
}

// Multiple conditions can be combined in the query
var recentActiveTasks []Task
if err := taskRepo.PaginateWhere(
    &recentActiveTasks, 
    page, 
    pageSize, 
    "status = ? AND created_at > ?", 
    "active", 
    time.Now().AddDate(0, 0, -7),
); err != nil {
    log.Printf("Filtered pagination failed: %v", err)
}

// Manual pagination with conditions (not recommended, use PaginateWhere instead)
offset := (page - 1) * pageSize
if err := db.Where("status = ?", "active").Offset(offset).Limit(pageSize).Find(&activeUsers).Error; err != nil {
    log.Printf("Manual pagination failed: %v", err)
}



```

## Integration with Gin

Here's an example of integrating pgconnect with a Gin web application:

```go
package main

import (
    "log"
    "net/http"
    "strconv"

    "github.com/JorgeSaicoski/pgconnect"
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

// User model
type User struct {
    gorm.Model
    Name  string `json:"name"`
    Email string `json:"email" gorm:"uniqueIndex"`
}

// Global variables
var db *pgconnect.DB
var userRepo *pgconnect.Repository[User]

func main() {
    // Connect to database
    config := pgconnect.DefaultConfig()
    var err error
    db, err = pgconnect.New(config)
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    
    // Migrate models
    if err := db.AutoMigrate(&User{}); err != nil {
        log.Fatalf("Failed to migrate: %v", err)
    }
    
    // Initialize repository
    userRepo = pgconnect.NewRepository[User](db)
    
    // Setup Gin router
    r := gin.Default()
    
    // Routes
    r.GET("/users", getUsers)
    r.GET("/users/:id", getUserByID)
    r.POST("/users", createUser)
    r.PUT("/users/:id", updateUser)
    r.DELETE("/users/:id", deleteUser)
    
    // Start server
    r.Run(":8080")
}

func getUsers(c *gin.Context) {
    // Get page and limit parameters
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
    
    var users []User
    if err := userRepo.Paginate(&users, page, limit); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, users)
}

func getUserByID(c *gin.Context) {
    id := c.Param("id")
    var user User
    
    if err := userRepo.FindByID(id, &user); err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        return
    }
    
    c.JSON(http.StatusOK, user)
}

func createUser(c *gin.Context) {
    var user User
    if err := c.BindJSON(&user); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    if err := userRepo.Create(&user); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusCreated, user)
}

func updateUser(c *gin.Context) {
    id := c.Param("id")
    var user User
    
    if err := userRepo.FindByID(id, &user); err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        return
    }
    
    if err := c.BindJSON(&user); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    if err := userRepo.Update(&user); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, user)
}

func deleteUser(c *gin.Context) {
    id := c.Param("id")
    var user User
    
    if err := userRepo.FindByID(id, &user); err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        return
    }
    
    if err := userRepo.Delete(&user); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}
```

## Complete Examples

### Basic Usage

```go
package main

import (
    "log"
    "time"

    "github.com/JorgeSaicoski/pgconnect"
    "gorm.io/gorm"
)

type Post struct {
    gorm.Model
    Title    string    `json:"title"`
    Content  string    `json:"content"`
    AuthorID uint      `json:"authorId"`
    Published bool     `json:"published" gorm:"default:false"`
    PublishAt time.Time `json:"publishAt,omitempty"`
}

func main() {
    // Connect to database
    config := pgconnect.DefaultConfig()
    config.DatabaseName = "blog"
    
    db, err := pgconnect.New(config)
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    defer db.Close()
    
    // Migrate model
    if err := db.AutoMigrate(&Post{}); err != nil {
        log.Fatalf("Migration failed: %v", err)
    }
    
    // Create repository
    postRepo := pgconnect.NewRepository[Post](db)
    
    // Create a post
    now := time.Now()
    post := Post{
        Title:     "Getting Started with pgconnect",
        Content:   "This is a sample post about using pgconnect...",
        AuthorID:  1,
        Published: false,
        PublishAt: now.Add(24 * time.Hour),
    }
    
    if err := postRepo.Create(&post); err != nil {
        log.Fatalf("Failed to create post: %v", err)
    }
    log.Printf("Created post with ID: %d", post.ID)
    
    // Find posts scheduled for future publishing
    var scheduledPosts []Post
    tomorrow := time.Now().Add(48 * time.Hour)
    if err := postRepo.FindWhere(&scheduledPosts, "published = ? AND publish_at < ?", false, tomorrow); err != nil {
        log.Fatalf("Failed to find scheduled posts: %v", err)
    }
    
    log.Printf("Found %d scheduled posts", len(scheduledPosts))
    
    // Update posts to published status
    for i := range scheduledPosts {
        scheduledPosts[i].Published = true
        if err := postRepo.Update(&scheduledPosts[i]); err != nil {
            log.Printf("Failed to update post %d: %v", scheduledPosts[i].ID, err)
        }
    }
}
```

### Complex Example with Transactions and Relations

```go
package main

import (
    "log"
    "time"

    "github.com/JorgeSaicoski/pgconnect"
    "gorm.io/gorm"
)

type User struct {
    gorm.Model
    Name     string  `json:"name"`
    Email    string  `json:"email" gorm:"uniqueIndex"`
    Password string  `json:"-"` // Not exposed in JSON
    Posts    []Post  `json:"posts,omitempty" gorm:"foreignKey:AuthorID"`
}

type Post struct {
    gorm.Model
    Title    string    `json:"title"`
    Content  string    `json:"content"`
    AuthorID uint      `json:"authorId"`
    Author   User      `json:"author,omitempty" gorm:"foreignKey:AuthorID"`
    Tags     []Tag     `json:"tags,omitempty" gorm:"many2many:post_tags;"`
}

type Tag struct {
    gorm.Model
    Name  string `json:"name" gorm:"uniqueIndex"`
    Posts []Post `json:"posts,omitempty" gorm:"many2many:post_tags;"`
}

func main() {
    // Connect to database
    config := pgconnect.DefaultConfig()
    config.DatabaseName = "blog"
    
    db, err := pgconnect.New(config)
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    defer db.Close()
    
    // Migrate models
    if err := db.AutoMigrate(&User{}, &Post{}, &Tag{}); err != nil {
        log.Fatalf("Migration failed: %v", err)
    }
    
    // Create repositories
    userRepo := pgconnect.NewRepository[User](db)
    postRepo := pgconnect.NewRepository[Post](db)
    tagRepo := pgconnect.NewRepository[Tag](db)
    
    // Create a user, post, and tags within a transaction
    err = db.WithTransaction(func(tx *gorm.DB) error {
        // Create user
        user := User{
            Name:     "Alice Smith",
            Email:    "alice@example.com",
            Password: "securepassword", // Would normally be hashed
        }
        
        if err := tx.Create(&user).Error; err != nil {
            return err
        }
        
        // Create tags
        goTag := Tag{Name: "Go"}
        if err := tx.Create(&goTag).Error; err != nil {
            return err
        }
        
        dbTag := Tag{Name: "Database"}
        if err := tx.Create(&dbTag).Error; err != nil {
            return err
        }
        
        // Create post with tags
        post := Post{
            Title:    "Working with PostgreSQL in Go",
            Content:  "This post explains how to connect to PostgreSQL...",
            AuthorID: user.ID,
            Tags:     []Tag{goTag, dbTag},
        }
        
        if err := tx.Create(&post).Error; err != nil {
            return err
        }
        
        return nil
    })
    
    if err != nil {
        log.Fatalf("Transaction failed: %v", err)
    }
    
    // Retrieve posts with author and tags (using the repository pattern)
    var posts []Post
    if err := postRepo.FindAll(&posts); err != nil {
        log.Fatalf("Failed to fetch posts: %v", err)
    }
    
    // To load relations, we need to use the underlying GORM functions
    for i := range posts {
        var post Post
        if err := db.DB.Preload("Author").Preload("Tags").First(&post, posts[i].ID).Error; err != nil {
            log.Printf("Failed to load relations for post %d: %v", posts[i].ID, err)
            continue
        }
        
        log.Printf("Post: %s by %s", post.Title, post.Author.Name)
        log.Printf("Tags: ")
        for _, tag := range post.Tags {
            log.Printf("- %s", tag.Name)
        }
    }
}
```

This guide should help you use all the functionality available in the pgconnect library effectively.