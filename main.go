package main

import (
	"flag"
	"os"
	"strings"
	cassieConf "github.com/cloudurable/cassandra-cloud/impl"
	lg "github.com/advantageous/go-logback/logging"
)



func main() {



	debug, configFilename, logger := initialCommandLineParse()

	config, err := cassieConf.LoadConfig(configFilename, debug, logger)

	if err != nil {
		logger.Errorf("Unable to load config filename %s  \n", configFilename)
		logger.ErrorError("Error was", err)
		os.Exit(1)
	}


	cassieConf.ProcessTemplate(config.YamlConfigTemplate, config.YamlConfigFileName, config, logger)
	cassieConf.ProcessTemplate(config.JvmOptionsTemplate, config.JvmOptionsFileName, config, logger)
}

func initialCommandLineParse() (bool, string, lg.Logger) {
	flag.Bool("debug", false, "Turn on debugging")
	flag.String("config", "", "Location of config file")
	var debug bool
	var configFilename string
	foundIndex := -1
	for index, arg := range os.Args {
		if arg == "-debug" {
			debug = true
		} else if strings.HasPrefix(arg, "-debug=") {
			debug = true
		} else if arg == "-config" {
			foundIndex = index
		} else if strings.HasPrefix(arg, "-config=") {
			split := strings.Split(arg, "=")
			configFilename = split[1]
		}
	}
	if foundIndex != -1 {
		configFilename = os.Args[foundIndex+1]
	}
	var logger lg.Logger
	if debug {
		logger = lg.NewSimpleDebugLogger("cassandra-cloud")
	} else {
		logger = lg.NewSimpleLogger("cassandra-cloud")
	}
	if configFilename == "" {
		configFilename = os.Getenv("CASSANDRA_CLOUD_CONFIG")
	}
	if configFilename == "" {
		cassandraHome := os.Getenv("CASSANDRA_HOME")
		if cassandraHome == "" {
			configFilename = "/opt/cassandra/conf/cloud.conf"
		} else {
			configFilename = cassandraHome + "/conf/cloud.conf"
		}
	}
	return debug, configFilename, logger
}






