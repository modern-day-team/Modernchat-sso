package store

import (
	"errors"
	"slices"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/modern-day-team/Modernchat-sso/server/internal/model"
)

// ErrLastAdmin 至少需要保留一个站长（角色变更的原子约束）。
var ErrLastAdmin = errors.New("last admin")

// UserStore users 表访问。
type UserStore struct{ db *gorm.DB }

// NewUserStore 构造。
func NewUserStore(db *gorm.DB) *UserStore { return &UserStore{db: db} }

// Create 新建用户。
func (s *UserStore) Create(u *model.User) error { return s.db.Create(u).Error }

// Save 保存用户（改密/封禁等全字段更新）。
func (s *UserStore) Save(u *model.User) error { return s.db.Save(u).Error }

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

// UserListQuery 用户列表查询条件（对齐主仓库 admin 列表的过滤维度）。
type UserListQuery struct {
	Page     int
	PageSize int
	Kw       string // 匹配用户名/昵称
	Role     string // user | admin | 空
	Status   string // active | banned | 空
}

// List 分页用户列表。
func (s *UserStore) List(q UserListQuery) ([]model.User, int64, error) {
	page, size := q.Page, q.PageSize
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	tx := s.db.Model(&model.User{})
	if kw := strings.TrimSpace(q.Kw); kw != "" {
		like := "%" + kw + "%"
		tx = tx.Where("username LIKE ? OR nickname LIKE ?", like, like)
	}
	if q.Role == "user" || q.Role == "admin" {
		tx = tx.Where("role = ?", q.Role)
	}
	if q.Status == "active" {
		tx = tx.Where("banned_until IS NULL OR banned_until <= ?", time.Now())
	} else if q.Status == "banned" {
		tx = tx.Where("banned_until > ?", time.Now())
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var users []model.User
	if err := tx.Order("id ASC").Offset((page - 1) * size).Limit(size).Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

// SetRole 变更角色；降职最后一个站长时返回 ErrLastAdmin。
// 事务内 FOR UPDATE 锁住全部 admin 行（主仓库同款原子保护）；
// 不 bump token_version——Auth 每请求查库，角色变更即时生效。
func (s *UserStore) SetRole(id uint, role string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var adminIDs []uint
		if err := tx.Model(&model.User{}).Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("role = ?", "admin").Order("id").Pluck("id", &adminIDs).Error; err != nil {
			return err
		}
		if role != "admin" && slices.Contains(adminIDs, id) && len(adminIDs) <= 1 {
			return ErrLastAdmin
		}
		return tx.Model(&model.User{}).Where("id = ?", id).Update("role", role).Error
	})
}
