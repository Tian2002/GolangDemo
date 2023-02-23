package middleware

import (
	"bytedance_douyin/biz/utils"
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"strings"
)

func Token() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		token := c.Query("token")
		if strings.Compare(token, "") == 0 {
			token = c.PostForm("token")
		}
		userId, err := utils.ParseToken(token)
		if err != nil {
			c.AbortWithMsg(err.Error(), consts.StatusOK)
		}
		c.Set("userId", userId)
		c.Next(ctx)
	}
}
