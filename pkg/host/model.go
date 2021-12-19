package host

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const (
	TableNameServerInfo       = "hosts"
	TableNameActiveServerInfo = "active_host"
)

type ServerInfo struct {
	ID          uint   `json:"id" gorm:"column:id;primarykey"`
	Name        string `json:"name" gorm:"column:name;unique"`
	User        string `json:"user" gorm:"column:user_name"`
	Address     string `json:"address" gorm:"column:address"`
	Port        int    `json:"port" gorm:"column:port"`
	Password    string `json:"password" gorm:"column:password"`
	KeyPath     string `json:"keypath" gorm:"column:keypath"`
	Description string `json:"description" gorm:"column:description"`

	CreatedAt time.Time `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"column:updated_at"`
}

func (info ServerInfo) TableName() string {
	return TableNameServerInfo
}

// HasCredentials returns a true if ServerInfo has password or key path in this host, otherwise false.
func (info *ServerInfo) HasCredentials() bool {
	if info.Password == "" && info.KeyPath == "" {
		return false
	}
	return true
}

func (info *ServerInfo) String() string {
	return fmt.Sprintf("%s (%s:%d)", info.Name, info.Address, info.Port)
}

func (info *ServerInfo) MarshalJSON() ([]byte, error) {
	v := struct {
		ID          uint      `json:"id"`
		Name        string    `json:"name"`
		User        string    `json:"user"`
		Address     string    `json:"address"`
		Port        int       `json:"port"`
		Password    string    `json:"password"`
		KeyPath     string    `json:"keypath"`
		Description string    `json:"description"`
		CreatedAt   time.Time `json:"createdAt"`
		UpdatedAt   time.Time `json:"updatedAt"`
	}{
		ID:          info.ID,
		Name:        info.Name,
		User:        info.User,
		Address:     info.Address,
		Port:        info.Port,
		Password:    strings.Repeat("*", len(info.Password)),
		KeyPath:     info.KeyPath,
		Description: info.Description,
		CreatedAt:   info.CreatedAt,
		UpdatedAt:   info.UpdatedAt,
	}
	return json.Marshal(v)
}

func (info *ServerInfo) ToJSON(pretty bool) string {
	var (
		b   []byte
		err error
	)
	if pretty {
		b, err = json.MarshalIndent(info, "", "  ")
	} else {
		b, err = json.Marshal(info)
	}
	if err != nil {
		return err.Error()
	}
	return string(b)
}

type ActiveServerInfo struct {
	ID           uint `gorm:"column:id;primarykey"`
	ServerInfo   ServerInfo
	ServerInfoID uint
}

func (info ActiveServerInfo) TableName() string {
	return TableNameActiveServerInfo
}
