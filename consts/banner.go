package consts

import (
	"fmt"
	"strings"

	"github.com/deluan/navidrome/static"
)

func getBanner() string {
	data, _ := static.Asset("banner.txt")
	return strings.TrimSuffix(string(data), "\n")
}

func Banner() string {
	version := "Version: " + Version()
	padding := strings.Repeat(" ", 52-len(version))
	return fmt.Sprintf("%s%s%s\n", getBanner(), padding, version)
}
