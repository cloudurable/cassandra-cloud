package main

import "fmt"
import (
	"flag"
	"github.com/hashicorp/hcl"
	"io/ioutil"
	lg "github.com/advantageous/go-logback/logging"
	"os"
	"text/template"
)

type Config struct {
	Verbose bool `hcl:"verbose"`
	// Addresses of hosts that are deemed contact points.
	// Cassandra nodes use this list of hosts to find each other and learn
	// the topology of the ring.  You must change this if you are running  multiple nodes!
	ContactPointSeeds string  `hcl:"cluster_seeds"`

	CassandraHome string  `hcl:"cassandra_home"`

	// Address or interface to bind to and tell other Cassandra nodes to connect to.
	// You _must_ change this if you want multiple nodes to be able to communicate!
	// Set listen_address OR listen_interface, not both.
	// Leaving it blank leaves it up to InetAddress.getLocalHost(). This
	// will always do the Right Thing _if_ the node is properly configured
	// (hostname, name resolution, etc), and the Right Thing is to use the
	// address associated with the hostname (it might not be).
	ClusterListenAddress string  `hcl:"cluster_address"`

	// Set listen_address OR listen_interface, not both. Interfaces must correspond
	// to a single address, IP aliasing is not supported.
	ClusterListenInterface string `hcl:"cluster_interface"`

	// Address to listen for client connections.
	ClientListenAddress string  `hcl:"client_address"`
	// Interface to listen for client connections. Address and interface can't both be set.
	ClientListenInterface string `hcl:"client_interface"`
	// Cassandra snitch type.
	Snitch string `hcl:"snitch"`
	//Name of the cassandra cluster.
	ClusterName string `hcl:"cluster_name"`
	//Number of tokens that this node wants/has. Used for Cassandra VNODES.
	NumTokens int `hcl:"num_tokens"`

	//Location of template file for cassandra conf.
	CassandraConfigTemplate string `hcl:"conf_yaml_template"`
}

func bindVars(config *Config) {

	flag.BoolVar(&config.Verbose, "v", false, "Turns on verbose mode")

	flag.StringVar(&config.ClusterName, "cluster-name", config.ClusterName,
		"Name of the cluster")

	flag.StringVar(&config.ContactPointSeeds, "cluster-seeds", config.ContactPointSeeds,
		"Comma delimited list of initial clustrer contact points for bootstrapping")

	flag.StringVar(&config.ClusterListenAddress, "cluster-address", config.ClusterListenAddress,
		"Cluster address for inter-node communication. Example: 192.43.32.10, localhost, etc.")

	flag.StringVar(&config.ClusterListenInterface, "cluster-interface", config.ClusterListenInterface,
		"Cluster interface for inter-node communication.  Example: eth0, eth1, etc.")

	flag.StringVar(&config.ClientListenAddress, "client-address", config.ClientListenAddress,
		"Client address for client driver communication. Example: 192.43.32.10, localhost, etc.")

	flag.StringVar(&config.ClientListenInterface, "client-interface", config.ClientListenInterface,
		"Client address for client driver communication. Example: eth0, eth1, etc.")

	flag.StringVar(&config.Snitch, "snitch", config.Snitch,
		"Snitch type. Example: GossipingPropertyFileSnitch, PropertyFileSnitch, Ec2Snitch, etc.")

	flag.StringVar(&config.CassandraConfigTemplate, "conf-yaml-template", config.CassandraConfigTemplate,
		"Location of cassandra configuration template")

	flag.Parse()

	if config.ClusterListenAddress == "" && config.ClusterListenInterface == "" {
		config.ClusterListenAddress = "localhost"
	}

	if config.ClientListenAddress == "" && config.ClientListenInterface == "" {
		config.ClientListenInterface = "eth0"
	}


}
func main() {

	debug := flag.Bool("debug", false, "Turn on Debugging")

	var logger lg.Logger

	if *debug {
		logger = lg.NewSimpleDebugLogger("cassandra-cloud")
	} else {
		logger = lg.NewSimpleLogger("cassandra-cloud")
	}

	configFilename := flag.String("config", "/opt/cassandra/conf/cloud.conf", "Location of config file")

	config, err := LoadConfig(*configFilename, *debug, logger)

	if err != nil {
		logger.Errorf("Unable to load config filename %s  \n", *configFilename)
		logger.ErrorError("Error was", err)
		os.Exit(1)
	}

	bindVars(config)

	if config.Verbose {
		fmt.Printf("Verbose on.\n")
	}


	yamlConfigTempalteBytes, err := ioutil.ReadFile(config.CassandraConfigTemplate)
	if err != nil {
		logger.Errorf("Unable to load cassandra yaml template %s  \n", config.CassandraConfigTemplate)
		logger.ErrorError("Error was", err)
		os.Exit(2)
	}



	template, err := template.New("test").Parse(string(yamlConfigTempalteBytes))
	if err != nil {
		logger.Errorf("Unable to parse template %s  \n", config.CassandraConfigTemplate)
		logger.ErrorError("Error was", err)
		os.Exit(3)
	}


	yamlConfFile, err := os.Create("/opt/cassandra/conf/cassandra.yaml")
	if err != nil {
		logger.ErrorError("Unable to open yaml conf", err)
		os.Exit(4)
	}

	if config.Verbose {
		logger.Printf("cluster name %s ", config.ClusterName)
	}
	template.Execute(yamlConfFile, config)
	yamlConfFile.Close()

}

func LoadConfigFromString(data string, logger lg.Logger) (*Config, error) {

	if logger == nil {
		logger = lg.NewSimpleLogger("SYSTEMD_CONFIG_DEBUG")
	}
	config := &Config{}

	logger.Debug("Loading log...")
	err := hcl.Decode(&config, data)
	if err != nil {
		return nil, err
	}

	if config.CassandraConfigTemplate == "" {
		config.CassandraConfigTemplate = "/opt/cassandra/conf/cassandra-conf.template"
	}

	return config, nil

}

func LoadConfig(filename string, debug bool, logger lg.Logger) (*Config, error) {

	if debug {
		logger.Printf("Loading config %s", filename)
	}

	configBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return LoadConfigFromString(string(configBytes), logger)
}
