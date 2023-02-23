package vo

type Comment struct {
	Status
	Comment *CommentInfo `json:"comment"` // 评论成功返回评论内容，不需要重新拉取整个列表
}

type CommentList struct {
	Status
	CommentList []CommentInfo `json:"comment_list"` // 评论列表
}

// CommentInfo
type CommentInfo struct {
	Content    string   `json:"content"`     // 评论内容
	CreateDate string   `json:"create_date"` // 评论发布日期，格式 mm-dd
	ID         int64    `json:"id"`          // 评论id
	User       UserInfo `json:"user"`        // 评论用户信息
}
