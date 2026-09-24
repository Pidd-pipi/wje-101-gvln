package service

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/wjecoffeetaste/wjecoffeetaste/internal/constants"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/model"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/repository"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/util"
)

// BeanService handles coffee bean library.
type BeanService struct {
	repo     *repository.CoffeeBeanRepository
	noteRepo *repository.TastingNoteRepository
	logger   *slog.Logger
}

// NewBeanService creates a BeanService.
func NewBeanService(repo *repository.CoffeeBeanRepository, noteRepo *repository.TastingNoteRepository, logger *slog.Logger) *BeanService {
	return &BeanService{repo: repo, noteRepo: noteRepo, logger: logger}
}

// Create adds a bean (admin).
func (s *BeanService) Create(b *model.CoffeeBean) (*model.CoffeeBean, error) {
	if !constants.IsValidProcessMethod(b.ProcessMethod) {
		return nil, util.NewAppError(422, constants.CodeValidationError,
			fmt.Sprintf("CoffeeBean[process_method=%s] create failed: invalid process method", b.ProcessMethod))
	}
	if b.FlavorTags == "" {
		b.FlavorTags = "[]"
	}
	if err := s.repo.Create(b); err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return nil, util.NewAppError(409, constants.CodeConflict,
				fmt.Sprintf("CoffeeBean[name=%s] create failed: name exists", b.Name))
		}
		s.logger.Error(fmt.Sprintf(constants.LogBeanCreateFailed, b.Name), "error", err)
		return nil, fmt.Errorf("bean create: %w", err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogBeanCreateSuccess, b.Name), "id", b.ID)
	return b, nil
}

// Update edits a bean (admin).
func (s *BeanService) Update(id uint, b *model.CoffeeBean) (*model.CoffeeBean, error) {
	exist, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("bean update find: %w", err)
	}
	if b.Name != "" {
		exist.Name = b.Name
	}
	if b.Origin != "" {
		exist.Origin = b.Origin
	}
	if b.ProcessMethod != "" {
		if !constants.IsValidProcessMethod(b.ProcessMethod) {
			return nil, util.NewAppError(422, constants.CodeValidationError, "invalid process method")
		}
		exist.ProcessMethod = b.ProcessMethod
	}
	if b.FlavorTags != "" {
		exist.FlavorTags = b.FlavorTags
	}
	if b.Description != "" {
		exist.Description = b.Description
	}
	if exist.FlavorTags == "" {
		exist.FlavorTags = "[]"
	}
	if err := s.repo.Update(exist); err != nil {
		return nil, fmt.Errorf("bean update: %w", err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogBeanUpdateSuccess, id), "id", id)
	return exist, nil
}

// Delete removes a bean (admin). Beans still referenced by notes are rejected.
func (s *BeanService) Delete(id uint) error {
	exist, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return util.NewAppError(404, constants.CodeNotFound, fmt.Sprintf("CoffeeBean[id=%d] not found", id))
		}
		return fmt.Errorf("bean delete find: %w", err)
	}
	count, err := s.noteRepo.CountByBean(id)
	if err != nil {
		return fmt.Errorf("bean delete count notes: %w", err)
	}
	if count > 0 {
		return util.NewAppError(409, constants.CodeConflict,
			fmt.Sprintf("豆种「%s」仍被 %d 篇品鉴笔记引用，无法撤下；请先处理相关笔记后再操作", exist.Name, count))
	}
	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("bean delete: %w", err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogBeanDeleteSuccess, id), "id", id)
	return nil
}

// List filters beans and includes the bound note count of each bean.
func (s *BeanService) List(origin, process, keyword string, page, pageSize int) ([]repository.BeanWithNoteCount, int64, error) {
	items, total, err := s.repo.List(origin, process, keyword, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("bean list: %w", err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogBeanListSuccess, process), "total", total)
	return items, total, nil
}
