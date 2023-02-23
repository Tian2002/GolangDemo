package utils

import (
	"errors"
	"strings"
)

// ActionValid 在error为nil时，返回true则是添加操作，返回false则是删除操作
func ActionValid(action string) (bool, error) {
	if strings.Compare("1", action) == 0 {
		return true, nil
	}
	if strings.Compare("2", action) == 0 {
		return false, nil
	}

	return false, errors.New("parameter parsing error") //参数解析错误
}
