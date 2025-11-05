package main

import (
	"github.com/roydevashish/queuectl/cmd"
	"github.com/roydevashish/queuectl/internal/storage"
)

func main() {
	storage.InitDB()
	cmd.Execute()
}
