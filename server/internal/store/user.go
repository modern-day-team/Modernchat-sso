package store

import (
	"gorm.io/gorm"

	"github.com/modern-day-team/Modernchat-sso/server/internal/model"
)

// UserStore users 表访问。
type UserStore struct{ db *gorm.DB }

// NewUserStore 构造。
func NewUserStore(db *gorm.DB) *UserStore { return &UserStore{db: db} }

// Create 新建用户。
func (s *UserStore) Create(u *model.User) error { return s.db.Create(u).Error }

// Count 用户总数（首个注册者判定用）。
func (s *UserStore) Count() (int64, error) {
	var n int64
	err := s.db.Model(&model.User{}).Count(&n).Error
	return n, err
}

// FindByUsername 按用户名查；不存在返回 gorm.ErrRecordNotFound。
func (s *UserStore) FindByUsername(username string) (*model.User, error) {
	var u model.User
	if err := s.db.Where("username = ?", username).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

// FindByEmail 按邮箱查；不存在返回 gorm.ErrRecordNotFound。
func (s *UserStore) FindByEmail(email string) (*model.User, error) {
	var u model.User
	if err := s.db.Where("email = ?", email).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

// FindByID 按主键查；不存在返回 gorm.ErrRecordNotFound。
func (s *UserStore) FindByID(id uint) (*model.User, error) {
	var u model.User
	if err := s.db.First(&u, id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

// BumpTokenVersion 使该用户已签发的全部 token 失效（登出/改密后调用）。
func (s *UserStore) BumpTokenVersion(id uint) error {
	return s.db.Model(&model.User{}).Where("id = ?", id).
		UpdateColumn("token_version", gorm.Expr("token_version + 1")).Error
}
