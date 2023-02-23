package server

import (
	"bytedance_douyin/biz/dao"
	"bytedance_douyin/biz/model/entity"
	"bytedance_douyin/biz/model/vo"
	"bytedance_douyin/biz/redis"
	"bytedance_douyin/biz/utils"
	"errors"
	"fmt"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/spf13/viper"
	"mime/multipart"
	"os"
	"strconv"
	"strings"
	"time"
)

const MaxVideos = 30 //配置
//var MaxVideos = config.Config.Feed.MaxVideos

func Feed(latestTime string, userId int64) ([]vo.VideoInfo, error) {
	//若已登录，得到用户的点赞视频列表 userLikeVideoList

	//从数据库中获得数据的模型
	videoList := make([]entity.Video, 0, MaxVideos)

	//返回给server且前端要的模型
	var videoInfoList []vo.VideoInfo

	//dao层获得最多entity.MaxVideos个视频
	if strings.Compare(latestTime, "") == 0 {
		err := dao.GetVideosLimitMaxVideos(&videoList)
		if err != nil {
			return nil, err
		}
	} else {
		time, err := strconv.ParseInt(latestTime, 10, 64)
		if err != nil {
			return videoInfoList, err
		}
		err = dao.GetVideosLimitMaxVideosByLatestTime(&videoList, time)
		if err != nil {
			return nil, err
		}
	}

	//写一个函数，将通过拿到的videoList,得到需要的可以返回的VideoArray
	var err error
	videoInfoList, err = GetVideoArrayByVideoList(videoList, userId)

	return videoInfoList, err

}

func GetVideoArrayByVideoList(videoList []entity.Video, userId int64) ([]vo.VideoInfo, error) {
	var videoInfoList = make([]vo.VideoInfo, len(videoList))
	if len(videoList) == 0 {
		return videoInfoList, errors.New("已经没有视频啦")
	}
	if userId == 0 {
		//未登录状态查看feed
		for i := 0; i < MaxVideos; i++ {
			videoInfoList[i].ID = videoList[i].ID
			videoInfoList[i].Title = videoList[i].Title
			videoInfoList[i].CoverURL = videoList[i].CoverURL
			videoInfoList[i].PlayURL = videoList[i].PlayURL
			//下面要获得author,isFavorite,favoriteCount,commentCount的信息
			author, _ := GetUserInfoById(videoList[i].UserId)
			videoInfoList[i].Author = author
			videoInfoList[i].IsFavorite = false
			videoInfoList[i].FavoriteCount, _ = GetVideoFavoriteCount(videoInfoList[i].ID)
			videoInfoList[i].CommentCount, _ = GetVideoCommentCount(videoInfoList[i].ID)

			//需要改

			if i+1 == len(videoList) {
				videoInfoList = videoInfoList[:i+1]
				break
			}
		}
	} else {
		//登录状态查看feed
		for i := 0; i < MaxVideos; i++ {
			videoInfoList[i].ID = videoList[i].ID
			videoInfoList[i].Title = videoList[i].Title
			videoInfoList[i].CoverURL = videoList[i].CoverURL
			videoInfoList[i].PlayURL = videoList[i].PlayURL
			//下面要获得author,isFavorite,favoriteCount,commentCount的信息
			author, _ := GetUserInfoById(videoList[i].UserId)
			isFollow1Db, err := IsFollow(userId, author.ID)
			if err != nil {
				return nil, err
			}

			author.IsFollow = isFollow1Db

			videoInfoList[i].Author = author

			isFavorite, _ := IsFavorite(userId, videoInfoList[i].ID)
			isFavorite1Rdb := isFavorite1Rdb(userId, videoInfoList[i].ID)
			if isFavorite1Rdb == true && isFavorite == true || isFavorite1Rdb == true && isFavorite == false {
				videoInfoList[i].IsFavorite = true
			}
			if isFavorite1Rdb == false && isFavorite == false || isFavorite1Rdb == false && isFavorite == true {
				videoInfoList[i].IsFavorite = false
			}

			videoInfoList[i].FavoriteCount, _ = GetVideoFavoriteCount(videoInfoList[i].ID)
			if isFavorite1Rdb == false && isFavorite == true {
				videoInfoList[i].FavoriteCount--
			}
			if isFavorite1Rdb == true && isFavorite == false {
				videoInfoList[i].FavoriteCount++
			}

			videoInfoList[i].CommentCount, _ = GetVideoCommentCount(videoInfoList[i].ID)

			if i+1 == len(videoList) {
				videoInfoList = videoInfoList[:i+1]
				break
			}
		}

	}

	return videoInfoList, nil
}

func PublishVideo(userId int64, videoTitle string, videoData *multipart.FileHeader, c *app.RequestContext) error {
	//创建文件夹
	videoPath := fmt.Sprintf("./static/videos/userId%d", userId)
	exit, err := HasDir(videoPath)
	if err != nil {
		return err
	}
	err = CreatDir(exit, videoPath)
	if err != nil {
		return err
	}

	//准备数据
	//videoName := fmt.Sprintf("%d_%s_%v.mp4", userId, videoTitle, time.Now().Format("2006-01-02 150405"))
	//videoPathAndName := fmt.Sprintf("%v/%d_%s_%v.mp4", videoPath, userId, videoTitle, time.Now().Format("2006-01-02 150405")) //可优化
	videoName := fmt.Sprintf("%d%s%v.mp4", userId, videoTitle, time.Now().Unix())
	pictureName := fmt.Sprintf("%d%s%v.jpg", userId, videoTitle, time.Now().Unix())
	videoPathAndName := fmt.Sprintf("%v/%d%s%v.mp4", videoPath, userId, videoTitle, time.Now().Unix()) //可优化
	videoPathAndNameOld := fmt.Sprintf("%v/%d%v.mp4", videoPath, userId, time.Now().Unix())
	if err = c.SaveUploadedFile(videoData, videoPathAndNameOld); err != nil {
		return err
	}
	err = utils.DoFfmpeg(videoPathAndNameOld, videoPathAndName)

	if err = c.SaveUploadedFile(videoData, videoPathAndName); err != nil {
		return err
	}
	host := "172.29.219.224" //配置
	port := "8888"           //配置
	fmt.Println(viper.GetString("thisServer.host"))
	//host := config.Config.App.Host //配置
	//port := config.Config.App.Port //配置
	videoUrl := fmt.Sprintf("http://%s:%s/static/videos/userId%d/%v", host, port, userId, videoName)
	coverUrl := fmt.Sprintf("http://%s:%s/static/videos/userId%d/%v", host, port, userId, pictureName)

	if err != nil {
		return err
	}
	//存储到数据库
	newVideo := &entity.Video{
		UserId:   userId,
		Title:    videoTitle,
		CoverURL: coverUrl,
		PlayURL:  videoUrl,
	}
	return dao.CreateVideoByEntity(newVideo)

	//视频名称格式：userId_videoTitle_time.Now
	//videoName := fmt.Sprintf("%d_%s_%v", userId, videoTitle, time.Now())

	//把视频存到本机本项目中的videos目录中
	//saveFile := filepath.Join("./static/videos/", videoName)
}

func PublishList(u, m string) ([]vo.VideoInfo, error) {

	myId, _ := strconv.ParseInt(m, 10, 64)
	userId, _ := strconv.ParseInt(u, 10, 64)
	videoList := make([]vo.VideoInfo, 0)
	userInfo := &vo.UserInfo{}
	err := dao.GetUserInfoById(userId, userInfo)
	if err != nil {
		return videoList, err
	}
	videos := make([]entity.Video, 0)
	err = dao.GetVideoListByUserId(userId, &videos)
	if err != nil {
		return videoList, err
	}
	for _, v := range videos {
		commentCount, err1 := dao.GetVideoCommentCount(v.ID)
		if err1 != nil {
			return videoList, err1
		}
		favoriteCount, err2 := dao.GetVideoFavoriteCount(v.ID)
		if err2 != nil {
			return videoList, err1
		}
		isFavorite, err3 := dao.IsFavorite(myId, v.ID)
		if err3 != nil {
			return videoList, err1
		}
		video := vo.VideoInfo{
			Author:        *userInfo,
			CommentCount:  commentCount,
			CoverURL:      v.CoverURL,
			FavoriteCount: favoriteCount,
			ID:            v.ID,
			IsFavorite:    isFavorite,
			PlayURL:       v.PlayURL,
			Title:         v.Title,
		}
		videoList = append(videoList, video)
	}
	return videoList, nil
}

// HasDir 判断该目录是否纯在
func HasDir(path string) (bool, error) {
	_, _err := os.Stat(path)
	if _err == nil {
		return true, nil
	}
	if os.IsNotExist(_err) {
		return false, nil
	}
	return false, _err
}

// CreatDir 判断是否需要创建目录
func CreatDir(exit bool, videoPath string) error {
	if !exit {
		err := os.MkdirAll(videoPath, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

func GetUserInfoById(userId int64) (vo.UserInfo, error) {
	var userInfo vo.UserInfo
	err := dao.GetUserInfoById(userId, &userInfo)
	if err != nil {
		return vo.UserInfo{}, err
	}
	userInfo.TotalFavorited, err = dao.GetTotalFavorited(userId)
	if err != nil {
		return vo.UserInfo{}, err
	}
	userInfo.WorkCount, err = dao.GetWorkCount(userId)
	if err != nil {
		return vo.UserInfo{}, err
	}
	userInfo.FavoriteCount, err = dao.GetFavoriteCount(userId)
	if err != nil {
		return vo.UserInfo{}, err
	}
	userInfo.FollowCount, err = dao.GetFollowCount(userId)
	if err != nil {
		return vo.UserInfo{}, err
	}
	userInfo.FollowerCount, err = dao.GetFollowerCount(userId)
	if err != nil {
		return vo.UserInfo{}, err
	}
	return userInfo, err

}

func GetVideoFavoriteCount(videoId int64) (int64, error) {
	count, err := dao.GetVideoFavoriteCount(videoId)
	if err != nil {
		return 0, err
	}
	return count, nil
}
func GetVideoCommentCount(videoId int64) (int64, error) {
	count, err := dao.GetVideoCommentCount(videoId)
	if err != nil {
		return 0, err
	}
	return count, nil
}
func IsFavorite(userId int64, videoId int64) (bool, error) {
	isFavorite, err := dao.IsFavorite(userId, videoId)
	if err != nil {
		return false, err
	}
	return isFavorite, nil
}

func IsFollow(userId int64, toUserId int64) (bool, error) {
	isFavorite, err := dao.IsFollow(userId, toUserId)
	if err != nil {
		return false, err
	}
	return isFavorite, nil
}
func isFavorite1Rdb(userId int64, videoId int64) bool {
	userIdStr := strconv.FormatInt(userId, 10)
	isFavorite := redis.RdbUserLike.SIsMember(redis.Ctx, userIdStr, videoId)
	//println("userid:", userIdStr)
	//println("toUserId:", videoId)
	//println(isFavorite.Val())

	return isFavorite.Val()
}

func GetVideoArrayByVideoListForFavoriteList(videoList []entity.Video, userId int64) ([]vo.VideoInfo, error) {
	var videoInfoList = make([]vo.VideoInfo, len(videoList))
	if len(videoList) == 0 {
		return videoInfoList, errors.New("已经没有视频啦")
	}

	for i := 0; i < len(videoList); i++ {
		videoInfoList[i].ID = videoList[i].ID
		videoInfoList[i].Title = videoList[i].Title
		videoInfoList[i].CoverURL = videoList[i].CoverURL
		videoInfoList[i].PlayURL = videoList[i].PlayURL
		//下面要获得author,isFavorite,favoriteCount,commentCount的信息
		author, _ := GetUserInfoById(videoList[i].UserId)
		videoInfoList[i].Author = author
		videoInfoList[i].IsFavorite, _ = IsFavorite(userId, videoInfoList[i].ID)
		videoInfoList[i].FavoriteCount, _ = GetVideoFavoriteCount(videoInfoList[i].ID)
		videoInfoList[i].CommentCount, _ = GetVideoCommentCount(videoInfoList[i].ID)

	}

	return videoInfoList, nil
}
