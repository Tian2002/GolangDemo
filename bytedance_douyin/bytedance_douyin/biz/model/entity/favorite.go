package entity

type Favorite struct {
	ID          int64 `json:"id"`
	UserID      int64 `json:"user_id"`
	VideoUserID int64 `gorm:"video_user_id"`
	VideoID     int64 `json:"video_id"`
}

func (Favorite) TableName() string {
	return "favorite"
}
