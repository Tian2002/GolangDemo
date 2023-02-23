package server

import (
	"bytedance_douyin/biz/dao"
	"bytedance_douyin/biz/model/entity"
	"bytedance_douyin/biz/model/vo"
	"bytedance_douyin/biz/utils"
	"errors"
	"gorm.io/gorm"
	"strconv"
)

// Register Register函数用于注册用户，主要进行以下过程：查询用户是否存在、更新
func Register(username, password string) (int64, string, error) {
	user := entity.User{
		UserName: username,
		Password: password,
	}

	//合法性校验
	if len(user.Password) < 5 {
		return 0, "", errors.New("密码长度小于5")
	}

	//查询用户是否存在
	if err := dao.QueryUser(&user); err != nil {
		return 0, "", err
	}

	var token string
	dao.DB.Transaction(func(tx *gorm.DB) error {
		//添加用户到数据库
		if err := dao.AddInUser(&user, tx); err != nil {
			return err
		}
		var userInfo entity.UserInfo
		userInfo.UserName = user.UserName
		userInfo.ID = user.ID
		if err := dao.AddInUserInfo(&userInfo, tx); err != nil {
			return err
		}

		//颁发token
		var err error
		token, err = utils.GenerateToken(user.ID)
		if err != nil {
			return err
		}
		return nil
	})

	return user.ID, token, nil
}

func Login(username, password string) (int64, string, error) {
	//合法性校验
	if len(password) < 6 {
		return 0, "", errors.New("密码长度至少为6位")
	}

	user := &entity.User{
		UserName: username,
		Password: password,
	}

	//查询用户对比密码是否正确
	if err := dao.Login(user); err != nil {
		return 0, "", err
	}

	//颁发token
	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		return 0, "", err
	}

	return user.ID, token, nil
}

func GetUserInfo(queryId int64, token string) (vo.UserInfo, error) {
	userId, _ := strconv.ParseInt(token, 10, 64)

	//判断是登录还是进入其他人的主页 //优化  不管查谁都可以queryId 但是如果是查自己的话就不用查isFollow

	userInfo, err := GetUserInfoById(queryId)
	if err != nil {
		return userInfo, err
	}
	if queryId != userId { //查是否关注
		isFollow, _ := IsFavorite(userId, queryId)
		userInfo.IsFollow = isFollow
	}

	return userInfo, nil

}
