package middleware

import (
	"bytedance_douyin/biz/handler"
	"bytedance_douyin/biz/model/vo"
	"bytedance_douyin/biz/utils"
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"strings"
)

func Password() []app.HandlerFunc {
	return []app.HandlerFunc{
		func(ctx context.Context, c *app.RequestContext) {
			password := c.Query("password")
			if strings.Compare(password, "") == 0 {
				password = c.PostForm("password")
			}
			//验证密码长度是否小于6
			if len(password) < 6 {
				c.JSON(consts.StatusOK, handler.UserBaseInfo{
					Status: vo.Status{
						StatusCode: -1,
						StatusMsg:  "密码长度至少为6位",
					},
				})
				c.Abort()
			}
			//将密码进行MD5加密
			passwordMD5 := utils.GetMD5(password)
			c.Set("password", passwordMD5)
			c.Next(ctx)
		},
	}
}
