package main

import (
	"log"

	"github.com/jessevdk/go-flags"
	"github.com/mariusdieckmann/igvmultibrowser/server"
	"github.com/spf13/viper"
)

var opts struct {
	ConfigFile string `short:"c" long:"configfile" description:"File of the config file" default:"config/default-conf.yaml"`
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	_, err := flags.Parse(&opts)
	if err != nil {
		log.Fatalln(err.Error())
	}

	viper.SetConfigFile(opts.ConfigFile)
	err = viper.ReadInConfig()
	if err != nil {
		log.Fatalln(err.Error())
	}

	server.Run()
}
