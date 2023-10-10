package main

import (
	"fmt"
	"reflect_parse_config/config"
)

func main() {
	config.SetUpConfig("./config.cfg")

	fmt.Println(config.Properties)
}
