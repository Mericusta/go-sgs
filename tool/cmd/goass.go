package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/Mericusta/go-sgs/tool/src/goass"
)

func main() {
	workMode := strings.ToLower(os.Getenv("MODE"))
	switch workMode {
	case "goass":
		goass.GoAssistant()
	default:
		panic(fmt.Sprintf("unknown work mode %v", workMode))
	}
}
