package main

import (
	"github.com/fallmo/obc-watcher/cmd/obc-watcher/api"
	"github.com/fallmo/obc-watcher/cmd/obc-watcher/utils"
)

func main() {
	utils.StartupTasks()
	go api.StartServer()
	// utils.StartWatchingOBCs()
	utils.StartOBCInformer()
}
