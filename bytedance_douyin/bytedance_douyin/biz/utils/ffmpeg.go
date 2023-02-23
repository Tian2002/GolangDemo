package utils

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type Bind struct {
	FFMpeg         string        `json:"ffmpeg"`          //yaml:"ffmpeg"
	CommandTimeout time.Duration `json:"command_timeout"` //yaml:"command_timeout
}

// Transcoding h264 mp4 format file, overwrite is true, overwrite an existing file
func (ff *Bind) Transcoding(src string, dst string, overwrite bool) (output []byte, err error) {
	args := []string{"-i", src, "-c:v", "libx264", "-strict", "-2", dst}
	if overwrite {
		args = append([]string{"-y"}, args...)
	}
	ctx, _ := context.WithTimeout(context.Background(), ff.CommandTimeout)
	cmd := exec.CommandContext(ctx, ff.FFMpeg, args...)
	output, err = cmd.Output()
	if ctx.Err() != nil {
		return output, ctx.Err()
	} else if err != nil {
		return output, err
	}
	return output, err
}

// Thumbnail a thumbnail taken from a moment in the video, overwrite is true, overwrite an existing file
func (ff *Bind) Thumbnail(src string, dst string, duration time.Duration, overwrite bool) error {
	//args := []string{"-i", src, "-ss", fmt.Sprintf("%f", duration.Seconds()), "-vframes", "1", dst}
	args := []string{"-i", src, "-ss", fmt.Sprintf("%f", duration.Seconds()), "-vframes", "1", dst}
	if overwrite {
		args = append([]string{"-y"}, args...)
	}
	cmd := exec.Command(ff.FFMpeg, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func DoFfmpeg(src, dst string) error {
	bind := Bind{
		//FFMpeg: `D:/Program Files/ffmpeg-2023-02-16-git-aeceefa622-essentials_build/bin/ffmpeg.exe`, //配置
		FFMpeg:         `./biz/utils/ffmpeg-2023-02-16-git-aeceefa622-essentials_build/bin/ffmpeg.exe`, //配置
		CommandTimeout: 60 * time.Second,
	}
	//转码
	_, err := bind.Transcoding(src, dst, true)
	if err != nil {
		fmt.Println(err)
		return errors.New("视频转码失败")
	}
	//截取封面
	index := strings.LastIndex(dst, ".")
	dst = dst[:index] + ".jpg"
	err = bind.Thumbnail(src, dst, 1, true)
	if err != nil {
		return errors.New("截取封面帧失败")
	}
	return nil
}
