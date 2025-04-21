package pgconnect

// Repository provides a generic repository pattern for database operations
type Repository[T any] struct {
	db *DB
}

// NewRepository creates a new repository for the given model
func NewRepository[T any](db *DB) *Repository[T] {
	return &Repository[T]{db: db}
}

// Create inserts a new record
func (r *Repository[T]) Create(model *T) error {
	return r.db.Create(model).Error
}

// FindByID retrieves a record by ID
func (r *Repository[T]) FindByID(id interface{}, result *T) error {
	return r.db.First(result, id).Error
}

// FindAll retrieves all records
func (r *Repository[T]) FindAll(result *[]T) error {
	return r.db.Find(result).Error
}

// FindWhere finds records matching the given conditions
func (r *Repository[T]) FindWhere(result *[]T, query interface{}, args ...interface{}) error {
	return r.db.Where(query, args...).Find(result).Error
}

// FindOne finds a single record matching the given conditions
func (r *Repository[T]) FindOne(result *T, query interface{}, args ...interface{}) error {
	return r.db.Where(query, args...).First(result).Error
}

// Update updates a record
func (r *Repository[T]) Update(model *T) error {
	return r.db.Save(model).Error
}

// Delete deletes a record
func (r *Repository[T]) Delete(model *T) error {
	return r.db.Delete(model).Error
}

// Count counts records matching the given conditions
func (r *Repository[T]) Count(count *int64, query interface{}, args ...interface{}) error {
	db := r.db // Use the wrapper directly, not the internal r.db.DB field
	if query != nil {
		db = &DB{DB: db.DB.Where(query, args...)} // Wrap the new DB instance properly
	}
	var model T
	return db.DB.Model(&model).Count(count).Error
}

// Paginate retrieves records with pagination
func (r *Repository[T]) Paginate(result *[]T, page, pageSize int) error {
	offset := (page - 1) * pageSize
	return r.db.Offset(offset).Limit(pageSize).Find(result).Error
}
