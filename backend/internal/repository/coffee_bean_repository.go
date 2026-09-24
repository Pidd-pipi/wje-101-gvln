package repository

import (
	"gorm.io/gorm"

	"github.com/wjecoffeetaste/wjecoffeetaste/internal/model"
)

// CoffeeBeanRepository handles bean persistence.
type CoffeeBeanRepository struct{ db *gorm.DB }

// NewCoffeeBeanRepository creates the repository.
func NewCoffeeBeanRepository(db *gorm.DB) *CoffeeBeanRepository { return &CoffeeBeanRepository{db: db} }

// Create inserts a bean.
func (r *CoffeeBeanRepository) Create(b *model.CoffeeBean) error { return translate(r.db.Create(b).Error) }

// FindByID locates a bean by id.
func (r *CoffeeBeanRepository) FindByID(id uint) (*model.CoffeeBean, error) {
	var b model.CoffeeBean
	if err := translate(r.db.First(&b, id).Error); err != nil {
		return nil, err
	}
	return &b, nil
}

// Update persists a bean.
func (r *CoffeeBeanRepository) Update(b *model.CoffeeBean) error { return translate(r.db.Save(b).Error) }

// Delete removes a bean by id.
func (r *CoffeeBeanRepository) Delete(id uint) error {
	res := r.db.Delete(&model.CoffeeBean{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// BeanWithNoteCount is a coffee bean paired with the number of notes referencing it.
type BeanWithNoteCount struct {
	model.CoffeeBean
	NoteCount int64 `json:"note_count" gorm:"->"`
}

// List filters beans by origin/process/keyword and includes the bound note count.
func (r *CoffeeBeanRepository) List(origin, process, keyword string, page, pageSize int) ([]BeanWithNoteCount, int64, error) {
	var items []BeanWithNoteCount
	var total int64
	q := r.db.Model(&model.CoffeeBean{})
	if origin != "" {
		q = q.Where("origin = ?", origin)
	}
	if process != "" {
		q = q.Where("process_method = ?", process)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("name LIKE ? OR flavor_tags LIKE ?", like, like)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Select("coffee_beans.*, (SELECT COUNT(*) FROM tasting_notes WHERE tasting_notes.coffee_bean_id = coffee_beans.id) AS note_count").
		Order("id ASC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}
