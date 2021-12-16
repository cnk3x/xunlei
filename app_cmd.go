package main

import (
	"log"
	"os"
	"os/signal"
	"time"
)

func run(args []string) {
	options := Options{
		Name:         "U-NAS-迅雷",
		Port:         2345,
		Internal:     false,
		DownloadPATH: "/downloads",
	}
	Flag(os.Args[0]+" run", &options, args)

	log.Printf("设备名称(不一定有用): %s", options.Name)
	log.Printf("面板端口: %d", options.Port)
	log.Printf("下载目录: %s", options.DownloadPATH)

	defer NewApp(options).Start().Stop()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, os.Kill)
	<-quit
	go func() {
		select {
		case <-time.After(time.Second * 5):
			os.Exit(1)
		}
	}()
}
