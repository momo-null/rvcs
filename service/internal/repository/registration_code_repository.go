package repository

import (
	"context"
	"time"

	"rvcs/internal/model"

	"gorm.io/gorm"
)

// RegistrationCodeRepository 注册码仓库接口
type RegistrationCodeRepository interface {
	Create(ctx context.Context, code *model.RegistrationCode) error
	GetByID(ctx context.Context, id uint) (*model.RegistrationCode, error)
	GetByCode(ctx context.Context, code string) (*model.RegistrationCode, error)
	List(ctx context.Context, filter map[string]interface{}, page, pageSize int) ([]*model.RegistrationCode, int64, error)
	Update(ctx context.Context, id uint, updates map[string]interface{}) error
	Delete(ctx context.Context, id uint) error
	DeleteExpired(ctx context.Context) (int64, error)
}

// registrationCodeRepository 注册码仓库实现
type registrationCodeRepository struct {
	db *gorm.DB
}

// NewRegistrationCodeRepository 创建注册码仓库
func NewRegistrationCodeRepository(db *gorm.DB) RegistrationCodeRepository {
	return &registrationCodeRepository{
		db: db,
	}
}

// Create 创建注册码
func (r *registrationCodeRepository) Create(ctx context.Context, code *model.RegistrationCode) error {
	return r.db.WithContext(ctx).Create(code).Error
}

// GetByID 根据ID获取注册码
func (r *registrationCodeRepository) GetByID(ctx context.Context, id uint) (*model.RegistrationCode, error) {
	var code model.RegistrationCode
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&code).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &code, nil
}

// GetByCode 根据注册码内容获取
func (r *registrationCodeRepository) GetByCode(ctx context.Context, code string) (*model.RegistrationCode, error) {
	var regCode model.RegistrationCode
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&regCode).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &regCode, nil
}

// List 查询注册码列表
func (r *registrationCodeRepository) List(ctx context.Context, filter map[string]interface{}, page, pageSize int) ([]*model.RegistrationCode, int64, error) {
	var codes []*model.RegistrationCode
	var total int64

	query := r.db.WithContext(ctx).Model(&model.RegistrationCode{})
	
	// 应用过滤条件
	for key, value := range filter {
		query = query.Where(key+" = ?", value)
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&codes).Error; err != nil {
		return nil, 0, err
	}

	return codes, total, nil
}

// Update 更新注册码
func (r *registrationCodeRepository) Update(ctx context.Context, id uint, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&model.RegistrationCode{}).Where("id = ?", id).Updates(updates).Error
}

// Delete 删除注册码
func (r *registrationCodeRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.RegistrationCode{}).Error
}

// DeleteExpired 删除过期注册码
func (r *registrationCodeRepository) DeleteExpired(ctx context.Context) (int64, error) {
	result := r.db.WithContext(ctx).Where("expires_at < ? AND status != ?", time.Now(), model.RegistrationCodeUsed).Delete(&model.RegistrationCode{})
	return result.RowsAffected, result.Error
}