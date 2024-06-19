package model

type User struct {
	NpHandle   string `gorm:"column:np_handle"`
	IconDB     []byte `gorm:"column:icon"`
	Icon       string `gorm:"-"`
	HeartCount int64  `gorm:"column:heartCount"`
	Yay2DB     []byte `gorm:"column:yay2"`
	Boo2DB     []byte `gorm:"column:boo2"`
	Yay2       string `gorm:"-"`
	Boo2       string `gorm:"-"`
}

func (User) TableName() string {
	return "user"
}
