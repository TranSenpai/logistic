package model

type ArticleFilter struct {
	Page    int    `form:"page" binding:"omitempty,min=1"`         // Mặc định >= 1
	Limit   int    `form:"limit" binding:"omitempty,min=1,max=50"` // Khống chế giới hạn tải để chống DDoS DB
	ID      int64  `form:"id" binding:"omitempty,min=1"`
	Keyword string `form:"keyword"`
	SortBy  string `form:"sortBy" binding:"omitempty,oneof=createdAt views likeCount"` // Chỉ cho phép sort theo các cột này
	IsDesc  bool   `form:"isDesc"`
}

func (f *ArticleFilter) SetDefault() {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit <= 0 {
		f.Limit = 10
	}
}
