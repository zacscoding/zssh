package host

import (
	"context"
	"gorm.io/gorm"
)

const (
	activeServerId = uint(1)
)

type Store interface {
	Save(ctx context.Context, info *ServerInfo) error
	FindByName(ctx context.Context, hostname string) (*ServerInfo, error)
	FindAll(ctx context.Context) ([]*ServerInfo, error)
	Update(ctx context.Context, info *ServerInfo) (int64, error)
	DeleteByName(ctx context.Context, hostname string) (int64, error)

	SaveOrUpdateActiveServerInfo(ctx context.Context, info *ServerInfo) error
	FindActiveServerInfo(ctx context.Context) (*ServerInfo, error)
}

// NewStore creates a new Store from given gorm.DB.
func NewStore(db *gorm.DB) Store {
	return &store{db: db}
}

type store struct {
	db *gorm.DB
}

func (hs *store) Save(ctx context.Context, info *ServerInfo) error {
	return hs.db.WithContext(ctx).Create(info).Error
}

func (hs *store) FindByName(ctx context.Context, hostname string) (*ServerInfo, error) {
	var info ServerInfo
	if err := hs.db.WithContext(ctx).Take(&info, "name = ?", hostname).Error; err != nil {
		return nil, err
	}
	return &info, nil
}

func (hs *store) FindAll(ctx context.Context) ([]*ServerInfo, error) {
	var servers []*ServerInfo
	if err := hs.db.WithContext(ctx).Find(&servers).Error; err != nil {
		return nil, err
	}
	return servers, nil
}

func (hs *store) Update(ctx context.Context, info *ServerInfo) (int64, error) {
	tx := hs.db.WithContext(ctx).Save(info)
	return tx.RowsAffected, tx.Error
}

func (hs *store) DeleteByName(ctx context.Context, hostname string) (int64, error) {
	tx := hs.db.WithContext(ctx).Where("name = ?", hostname).Delete(new(ServerInfo))
	return tx.RowsAffected, tx.Error
}

func (hs *store) SaveOrUpdateActiveServerInfo(ctx context.Context, info *ServerInfo) error {
	active := ActiveServerInfo{
		ID:           activeServerId,
		ServerInfo:   *info,
		ServerInfoID: info.ID,
	}
	return hs.db.WithContext(ctx).Save(&active).Error
}

func (hs *store) FindActiveServerInfo(ctx context.Context) (*ServerInfo, error) {
	var info ActiveServerInfo
	if err := hs.db.WithContext(ctx).
		Joins("ServerInfo").
		First(&info, activeServerId).
		Error; err != nil {
		return nil, err
	}
	return &info.ServerInfo, nil
}
