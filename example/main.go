package main

import (
	"github.com/mesment/sparrow/pkg/xlog"
)

func main() {

	config := xlog.Config{
		EnableConsole:true,
		ConsoleJSONFormat:true,
		ConsoleLevel: "debug",
		EnableFile:true,
		Name: "default.log",
		Dir: "./logfiles",
		Level: "info",
		AddCaller:true,
		CallerSkip:1,
		Debug:false,
	}
	logger := config.Build()
	//logger.SetLevel(xlog.DebugLevel)
	logger.Info("debug", xlog.String("a", "b"))
	logger.Infof("info %s", "a")
	logger.Debugw("debug", "a", "b")

}
