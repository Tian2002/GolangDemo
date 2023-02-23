package server

import (
	"bytedance_douyin/biz/dao"
	"bytedance_douyin/biz/model/entity"
	"bytedance_douyin/biz/model/vo"
	"bytedance_douyin/biz/redis"
	"bytedance_douyin/biz/utils"
	"errors"
	"fmt"
	"log"
	"strconv"
)

var userIdVar string
var countVar int

func FavoriteAction(token, actionType, videoID string, count int) error {
	ctx := redis.Ctx
	//验证token
	userId, err := utils.ParseToken(token)
	if err != nil {
		return err
	}
	//验证action是否有效
	var valid bool
	if valid, err = utils.ActionValid(actionType); err != nil {
		return errors.New("无效的操作")
	}
	//错误未处理
	videoId, err2 := strconv.ParseInt(videoID, 10, 64)
	if err2 != nil {
		return err2
	}
	videoIdStr := videoID
	userIdStr := strconv.FormatInt(userId, 10)
	setUserId(userIdStr)
	countVar = count

	//创建redis缓存集合
	//查看redis服务器中是否有 key=videoId 的set集合，若没有则创建
	if n, err := redis.RdbVideoLike.Exists(ctx, videoIdStr).Result(); n == 0 {
		if err != nil {
			log.Printf("查询是否存在 key=videoId 失败：%v", err)
			return err
		}
		if _, err := redis.RdbVideoLike.SAdd(ctx, videoIdStr, -1).Result(); err != nil {
			log.Printf("创建 Set（key=videoId） 失败：%v", err)
			return err
		}
	}
	//查看redis服务器中是否有 key=userId 的set集合，若没有则创建
	if n, err := redis.RdbUserLike.Exists(ctx, userIdStr).Result(); n == 0 {
		if err != nil {
			log.Printf("查询是否存在 key=userId 失败：%v", err)
			return err
		}
		if _, err := redis.RdbVideoLike.SAdd(ctx, userIdStr, -1).Result(); err != nil {
			log.Printf("创建 Set（key=videoId） 失败：%v", err)
			return err
		}
	}
	if n, err := redis.RdbUserDoNotLike.Exists(ctx, userIdStr).Result(); n == 0 {
		if err != nil {
			log.Printf("查询是否存在DoNotLike key=userId 失败：%v", err)
			return err
		}
		if _, err := redis.RdbVideoLike.SAdd(ctx, userIdStr, -1).Result(); err != nil {
			log.Printf("创建DoNotLike Set（key=videoId） 失败：%v", err)
			return err
		}
	}

	//若是点赞，把视频id添加到RdbUserLike key=userId 的set集合,把视频id从RdbUserDoNotLike key=userId 的set集合移除
	//若是取消点赞，把视频id从 key=userId 的set集合中删除，把视频id添加到RdbUserDoNotLike key=userId 的set集合
	if valid == true {
		if _, err := redis.RdbUserLike.SAdd(redis.Ctx, userIdStr, videoId).Result(); err != nil {
			log.Printf("添加赞操作1到redis失败：%v", err)
			return err
		}
		if _, err := redis.RdbUserDoNotLike.SRem(redis.Ctx, userIdStr, videoId).Result(); err != nil {
			log.Printf("添加赞操作2到redis失败：%v", err)
			return err
		}
	} else {
		if valid == false {
			if _, err := redis.RdbUserDoNotLike.SAdd(redis.Ctx, userIdStr, videoId).Result(); err != nil {
				log.Printf("添加赞操作3到redis失败：%v", err)
				return err
			}
			if _, err := redis.RdbUserLike.SRem(redis.Ctx, userIdStr, videoId).Result(); err != nil {
				log.Printf("添加赞操作4到redis失败：%v", err)
				return err
			}
		} else {
			return errors.New("action_type数据错误")
		}
	}

	//用来打印查错的
	//userLike, err1 := redis.RdbUserLike.SMembers(ctx, userIdStr).Result()
	//if err1 != nil {
	//	return err1
	//}
	//userDoNotLike, err2 := redis.RdbUserDoNotLike.SMembers(ctx, userIdStr).Result()
	//if err2 != nil {
	//	return err2
	//}
	//
	//fmt.Println("userLike:", userLike)
	//fmt.Println("userDoNotLike:", userDoNotLike)

	if err != nil {
		return err
	}
	return nil

}

func FavoriteList(userId int64) ([]vo.VideoInfo, error) {
	videoList := make([]entity.Video, 0)
	videos := make([]vo.VideoInfo, 0)
	err := dao.GetVideoListByUserId(userId, &videoList)
	if err != nil {
		return videos, err
	}
	videos, err = GetVideoArrayByVideoListForFavoriteList(videoList, userId)
	return videos, err
}

// RdbToDb 执行定时任务
func RdbToDb(second int, f utils.Fn) {
	t := utils.NewTick(second, f)
	err := t.Start()
	panic(err)

}

// KeyUserRdbToDb 从redis中取出来的一个string的value数组存入到数据库中
//未用，需要改
func KeyUserRdbToDb() error {
	if countVar >= 1 {
		var userIdStr = getUserId()
		//println(userIdStr)
		err := UserLikeSetRdbToDb(userIdStr)
		if err != nil {
			return err
		}
		err = UserDoNotLikeSetRdbToDb(userIdStr)
		if err != nil {
			return err
		}
		return nil
	}
	return nil
	//var userIdStr = getUserId()
	//println(userIdStr)
	//err := UserLikeSetRdbToDb(userIdStr)
	//if err != nil {
	//	return err
	//}
	//err = UserDoNotLikeSetRdbToDb(userIdStr)
	//if err != nil {
	//	return err
	//}
	//return nil

}
func UserLikeSetRdbToDb(userIdStr string) error {
	getUserLikeSet := redis.RdbUserLike.SMembers(redis.Ctx, userIdStr)

	res, err := getUserLikeSet.Result()

	if err != nil {
		return err
	}

	videoListStr := res

	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		return err
	}

	for i := 0; i < len(videoListStr); i++ {

		videoId, err := strconv.ParseInt(videoListStr[i], 10, 64)
		if videoId == 0 {
			continue
		}
		if err != nil {
			return err
		}

		//写进model里
		fmt.Println(res, videoId)
		videoUserID, err := dao.GetVideoUserIdByVideoId(videoId)
		if videoUserID == 0 {
			continue
		}
		if err != nil {
			return err
		}

		var n int64
		dao.DB.Model(&entity.Favorite{}).Where("user_id=? and video_id=?", userId, videoId).Count(&n)
		if n != 0 {
			continue
		}
		dao.DB.Create(&entity.Favorite{
			UserID:      userId,
			VideoID:     videoId,
			VideoUserID: videoUserID,
		})
	}
	return nil
}
func UserDoNotLikeSetRdbToDb(userIdStr string) error {
	getUserDoNotLikeSet := redis.RdbUserDoNotLike.SMembers(redis.Ctx, userIdStr)

	res, err := getUserDoNotLikeSet.Result()
	if err != nil {
		return err
	}

	videoListStr := res

	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		return err
	}
	for i := 0; i < len(videoListStr); i++ {
		videoId, err := strconv.ParseInt(videoListStr[i], 10, 64)
		if err != nil {
			return err
		}

		//删除对应数据
		dao.DB.Where("user_id=? and video_id=?", userId, videoId).Delete(entity.Favorite{})
	}
	return nil
}

func IsRdbToDb() {

	RdbToDb(5, KeyUserRdbToDb)
}

func getUserId() string {
	return userIdVar
}

func setUserId(userId string) {
	userIdVar = userId
}
