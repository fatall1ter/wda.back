package main

import (
	"git.countmax.ru/countmax/wda.back/infra"
	"github.com/sethvargo/go-signalcontext"
)

var (
	build   string
	githash string
	version = "0.0.7"
)

// @title WDA
// @version 1.0/0.0.4
// @Description This is a simple service for authorization control and users management
// @Description сервис для авторизации пользователей и управления пользователями

// @contact.name API Support
// @contact.url https://1020.watcom.ru
// @contact.email 1020@watcom.ru

func main() {
	ctx, cancel := signalcontext.OnInterrupt()
	defer cancel()
	serv := infra.NewServer(version, build, githash)
	serv.Run()
	<-ctx.Done()
	serv.Stop()
}
