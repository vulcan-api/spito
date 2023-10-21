package main

import (
	"fmt"
	"github.com/nasz-elektryk/spito-rules/lua"
	"os"
)

func main() {
	fmt.Println("Wrong way, run instead \"go test ./...\"")
	data, _ := os.ReadFile("./lua/test.lua")
	lua.DoesRulePasses(string(data))
}
