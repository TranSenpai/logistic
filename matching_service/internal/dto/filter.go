package model

type ArticleFilter struct {
	Page    int    `form:"page" binding:"omitempty,min=1"`
	Limit   int    `form:"limit" binding:"omitempty,min=1,max=50"`
	ID      int64  `form:"id" binding:"omitempty,min=1"`
	Keyword string `form:"keyword"`
	SortBy  string `form:"sortBy" binding:"omitempty,oneof=createdAt views likeCount"`
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