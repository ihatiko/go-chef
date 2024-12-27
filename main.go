package main

import (
	"fmt"
	gochefcodegenutils "github.com/ihatiko/go-chef-code-gen-utils"
	"log/slog"
	"os"
	"strings"
)

const corePathPackage = "github.com/ihatiko/go-chef-proxy"
const coreNamePackage = "go-chef-proxy"

func main() {
	params := strings.Join(os.Args[1:], " ")
	gochefcodegenutils.AutoUpdate(corePathPackage)
	composer := gochefcodegenutils.NewExecutor()
	proxyCommand := fmt.Sprintf("%s %s", coreNamePackage, params)
	result, err := composer.ExecDefaultCommand(proxyCommand)
	if err != nil {
		slog.Error("Error executing command: ", slog.Any("error", err), slog.String("command", params))
	}
	slog.Info(result.String())
}
