package main

import (
	"flag"
	"fmt"
	"github.com/mesment/sparrow/pkg/xlog"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func init() {
	flag.String("f", "", "config file")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)
}

func main() {

	log := xlog.DefaultConfig().Build()
	fmt.Printf("%v\n", log)
	log.Info("hello", xlog.Any("a", "b"))
	log.Info("msg", xlog.String("key", "value"))
}
