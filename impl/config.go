package impl

import (
	"io/ioutil"
	"github.com/hashicorp/hcl"
	lg "github.com/advantageous/go-logback/logging"
	"runtime"
	"os"
	"strconv"
	"flag"
	"fmt"
	"reflect"
	"strings"
	"C"
	"os/exec"
)

type Config struct {
	DataDirs []string `hcl:"data_dirs"`

	CassandraHome string `hcl:"home_dir"`
	// Addresses of hosts that are deemed contact points.
	// Cassandra nodes use this list of hosts to find each other and learn
	// the topology of the ring.  You must change this if you are running  multiple nodes!
	ClusterSeeds string `hcl:"cluster_seeds"`
	// Address or interface to bind to and tell other Cassandra nodes to connect to.
	// You _must_ change this if you want multiple nodes to be able to communicate!
	// Set listen_address OR listen_interface, not both.
	// Leaving it blank leaves it up to InetAddress.getLocalHost(). This
	// will always do the Right Thing _if_ the node is properly configured
	// (hostname, name resolution, etc), and the Right Thing is to use the
	// address associated with the hostname (it might not be).
	ClusterListenAddress string `hcl:"cluster_address"`
	// Set listen_address OR listen_interface, not both. Interfaces must correspond
	// to a single address, IP aliasing is not supported.
	ClusterListenInterface string `hcl:"cluster_interface"`
	// Address to listen for client connections.
	ClientListenAddress string `hcl:"client_address"`
	// Interface to listen for client connections. Address and interface can't both be set.
	ClientListenInterface string `hcl:"client_interface"`
	ClientPort int `hcl:"client_port"`
	//Name of the cassandra cluster.
	ClusterName string `hcl:"cluster_name"`
	ClusterPort    int `hcl:"cluster_port"`
	ClusterSslPort int `hcl:"cluster_ssl_port"`
	//AUTO, or a number string, i.e., 100MB
	CmsYoungGenSize string `hcl:"cms_young_gen_size"`
	CommitLogDir string `hcl:"commit_log_dir"`

	ReplaceAddress string `hcl:"replace_address"`

	//GC stats
	GCStatsEnabled bool `hcl:"gc_stats_enabled"`
	// CMS, G1, AUTO - Auto uses G1 if heap is over 8GB (default) but CMS if under.
	GC string `hcl:"gc"`
	//Threshold in GB of when to use G1 vs CMS
	G1ThresholdGBs int `hcl:"gc_g1_threshold_gbs"`
	//AUTO, or a number
	G1ParallelGCThreads string `hcl:"g1_parallel_threads"`
	//AUTO or the number or threads
	G1ConcGCThreads string `hcl:"g1_concurrent_threads"`


	//Location of cassandra jvm options file.
	JvmOptionsFileName string `hcl:"conf_jvm_options_file"`
	//Location of jvm options template.
	JvmOptionsTemplate string `hcl:"conf_jvm_options_template"`

	//AUTO, or a number string, i.e., 5GB
	MinHeapSize string `hcl:"min_heap_size"`
	//AUTO, or a number string, i.e., 5GB
	MaxHeapSize string `hcl:"max_heap_size"`
	MultiDataCenter bool `hcl:"multi_dc"`

	//Number of tokens that this node wants/has. Used for Cassandra VNODES.
	NumTokens int `hcl:"num_tokens"`

	// Cassandra snitch type.
	Snitch string `hcl:"snitch"`

	Verbose bool `hcl:"verbose"`


	//Location of template file for cassandra conf.
	YamlConfigTemplate string `hcl:"conf_yaml_template"`
	//Location of cassandra yaml config file.
	YamlConfigFileName string `hcl:"conf_yaml_file"`


}

func initConfigFile(configFileName string, logger lg.Logger) {
	if _, err := os.Stat(configFileName); os.IsNotExist(err) {
		logger.Debug("Cloud Cassandra config file does not exist so we are creating it", configFileName)
		err = ioutil.WriteFile(configFileName, []byte(CassandraCloudConfig), 0644)
		if err != nil {
			logger.ErrorError("Unable to write config file "+configFileName, err)
		}
	}
}

func LoadConfig(filename string, debug bool, logger lg.Logger) (*Config, error) {

	initConfigFile(filename, logger)

	if debug {
		logger.Printf("Loading config %s", filename)
	}

	configBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return LoadConfigFromString(string(configBytes), logger)
}

func displayConfig(config *Config) {

	reflected := reflect.ValueOf(*config)
	reflectedType := reflected.Type()

	fmt.Printf("%-30s %-20s %v\n", "Field Name", "Type", "Value")

	// loop through the struct's fields and set the map
	for i := 0; i < reflected.NumField(); i++ {
		value := reflected.Field(i)
		field := reflectedType.Field(i)

		fmt.Printf("%-30s %-20s %v\n", field.Name, field.Type.Name(), value.Interface())

	}

}

func LoadConfigFromString(data string, logger lg.Logger) (*Config, error) {
	config := &Config{}

	logger.Debug("Loading log...")
	err := hcl.Decode(&config, data)
	if err != nil {
		return nil, err
	}
	initDefaults(config, logger)
	bindCommandlineArgs(config, logger)

	if config.Verbose {
		displayConfig(config)
	}

	return config, nil

}

const CassandraCloudConfig = `

cluster_name = "My Cluster"

# Defaults to false. Used for verbose mode.
# verbose = false

# Comma delimited list of seed hosts. Defaults to 127.0.0.1. This is used for bootstrap only.
# Production clusters should have at least two seed servers.
# cluster_seeds = 127.0.0.1

# Cassandra home directory. Defaults to /opt/cassandra.
# home_dir = /opt/cassandra

# Are we supporting multiple DCs? Defaults to false.
# multi_dc = false

# Only one of these two can be set (defaults to address set to localhost Linux)
# cluster_address = localhost
# cluster_interface=eth1

# Only one of these two can be set (defaults to address set to localhost for OSX and interface set to eth0 for Linux)
# client_address=localhost
# client_interface=eth0

# Sets the snitch type for Cassandra. Defaults to simple snitch (for now).
# snitch=SimpleSnitch

# Sets up VNODE weight for servder. Defaults to 32 tokens per node
# num_tokens=32

# Location of template file. Defaults to {{home_dir}}/conf/cassandra-yaml.template
# conf_yaml_template = /opt/cassandra/conf/cassandra-yaml.template

# Location of cassandra config yaml file that we want to replace.
# Defaults to {{home_dir}}/conf/cassandra.yaml
# conf_yaml_file = /opt/cassandra/conf/cassandra.yaml

# Cluster port. Defaults to 7000.
# cluster_port = 7000
# Cluster SSL port. Defaults to 7001.
# cluster_ssl_port = 7001

# Client port. Defaults to 9042.
# client_port = 9042

# Data directories for Cassandra SSTables. Defaults to ["/opt/cassandra/data"]
# data_dirs = ["/opt/cassandra/data"]
`

func initDefaults(config *Config, logger lg.Logger) {

	overrideWithEnvOrDefault("CASSANDRA_GC", &config.GC, "AUTO", logger)
	overrideWithEnvOrDefault("CASSANDRA_G1_PARALLEL_THREADS", &config.G1ParallelGCThreads, "AUTO", logger)
	overrideWithEnvOrDefault("CASSANDRA_G1_CONCURRENT_THREADS", &config.G1ConcGCThreads, "AUTO", logger)
	overrideNumberWithEnvOrDefault("CASSANDRA_GC_G1_THRESHOLD_GB", &config.G1ThresholdGBs, 5, logger)
	overrideWithEnvOrDefault("CASSANDRA_CMS_YOUNG_GEN_SIZE", &config.CmsYoungGenSize, "AUTO", logger)
	overrideWithEnvOrDefault("CASSANDRA_MAX_HEAP_SIZE", &config.MaxHeapSize, "AUTO", logger)
	overrideWithEnvOrDefault("CASSANDRA_MIN_HEAP_SIZE", &config.MinHeapSize, "AUTO", logger)
	overrideWithEnvOrDefault("CASSANDRA_CMS_YOUNG_GEN_SIZE", &config.CmsYoungGenSize, "AUTO", logger)

	config.GC = strings.ToUpper(config.GC)
	config.G1ParallelGCThreads = strings.ToUpper(config.G1ParallelGCThreads)
	config.G1ConcGCThreads = strings.ToUpper(config.G1ConcGCThreads)
	config.CmsYoungGenSize = strings.ToUpper(config.CmsYoungGenSize)
	config.MinHeapSize = strings.ToUpper(config.MinHeapSize)
	config.MaxHeapSize = strings.ToUpper(config.MaxHeapSize)

	config = gcErgonomics(config, logger)

	overrideWithEnvOrDefault("CASSANDRA_CLUSTER_NAME", &config.ClusterName, "mycluster", logger)
	overrideWithEnvOrDefault("CASSANDRA_HOME_DIR", &config.CassandraHome, "/opt/cassandra", logger)
	overrideWithEnvOrDefault("CASSANDRA_CONF_YAML_TEMPLATE", &config.YamlConfigTemplate,
		config.CassandraHome+"/conf/cassandra-yaml.template", logger)
	overrideWithEnvOrDefault("CASSANDRA_CONF_YAML_FILE", &config.YamlConfigFileName,
		config.CassandraHome+"/conf/cassandra.yaml", logger)
	overrideWithEnvOrDefault("CASSANDRA_CONF_JVM_OPTIONS_TEMPLATE", &config.JvmOptionsTemplate,
		config.CassandraHome+"/conf/jvm-options.template", logger)
	overrideWithEnvOrDefault("CASSANDRA_CONF_JVM_OPTIONS_FILE", &config.JvmOptionsFileName,
		config.CassandraHome+"/conf/jvm.options", logger)

	overrideWithEnvOrDefault("CASSANDRA_SNITCH", &config.Snitch, "SimpleSnitch", logger)
	overrideWithEnvOrDefault("CASSANDRA_CLUSTER_SEEDS", &config.ClusterSeeds, "127.0.0.1", logger)
	overrideWithEnvOrDefault("CASSANDRA_CLIENT_INTERFACE", &config.ClientListenInterface, "", logger)
	overrideWithEnvOrDefault("CASSANDRA_CLIENT_ADDRESS", &config.ClientListenAddress, "", logger)
	overrideWithEnvOrDefault("CASSANDRA_CLUSTER_INTERFACE", &config.ClusterListenInterface, "", logger)
	overrideWithEnvOrDefault("CASSANDRA_CLUSTER_ADDRESS", &config.ClusterListenAddress, "", logger)
	overrideWithEnvOrDefault("CASSANDRA_COMMIT_LOG_DIR", &config.CommitLogDir, config.CassandraHome+"/commitlog", logger)

	overrideWithEnvOrDefault("CASSANDRA_REPLACE_ADDRESS", &config.ReplaceAddress, "", logger)

	overrideNumberWithEnvOrDefault("CASSANDRA_NUM_TOKENS", &config.NumTokens, 32, logger)
	overrideNumberWithEnvOrDefault("CASSANDRA_CLUSTER_PORT", &config.ClusterPort, 7000, logger)
	overrideNumberWithEnvOrDefault("CASSANDRA_CLUSTER_SSL_PORT", &config.ClusterSslPort, 7001, logger)
	overrideNumberWithEnvOrDefault("CASSANDRA_CLIENT_PORT", &config.ClientPort, 9042, logger)

	if config.ClientListenAddress != "" && config.ClientListenInterface != "" {
		logger.Error("The client listen address and the client listen interface can't both be set")
	} else if config.ClientListenAddress == "" && config.ClientListenInterface == "" {
		if runtime.GOOS == "linux" {
			logger.Debug("ClientListenAddress and ClientListenInterface were not set, setting to eth0 as the OS is Linux")
			config.ClientListenInterface = "eth0"
		} else {
			logger.Debug("ClientListenAddress and ClientListenInterface were not set, setting to localhost as the OS is NOT Linux")
			config.ClientListenAddress = "localhost"
		}
	}
	if config.ClusterListenAddress != "" && config.ClusterListenInterface != "" {
		logger.Error("The cluster listen address and the cluster listen interface can't both be set")
	} else if config.ClusterListenAddress == "" && config.ClusterListenInterface == "" {
		logger.Debug("ClusterListenAddress and ClusterListenInterface were not set, setting to localhost")
		config.ClusterListenAddress = "localhost"
	}

	initYamlTemplate(config.YamlConfigTemplate, logger)

	initJvmOptionsTemplate(config.JvmOptionsTemplate, logger)

}
func gcErgonomics(config *Config, logger lg.Logger) *Config {
	if config.G1ParallelGCThreads == "AUTO" {
		if runtime.NumCPU() > 10 {
			config.G1ParallelGCThreads = strconv.Itoa(runtime.NumCPU() - 1)
		} else {
			config.G1ParallelGCThreads = strconv.Itoa(runtime.NumCPU())
		}
	}
	if config.G1ConcGCThreads == "AUTO" {
		config.G1ConcGCThreads = config.G1ParallelGCThreads
	}
	if config.CmsYoungGenSize == "AUTO" {
		config.CmsYoungGenSize = strconv.Itoa(runtime.NumCPU()) + "00MB"
	}

	memSize := getMemory(logger)

	if config.MaxHeapSize == "AUTO" {
		maxHeapSize := memSize * 7 / 10
		config.MaxHeapSize = strconv.FormatUint( maxHeapSize / 1000000, 10 ) + "m"
	}
	if config.MinHeapSize == "AUTO" {
		config.MinHeapSize = config.MaxHeapSize
	}
	if config.GC == "AUTO" {
		actualGB := int(memSize / 1000000000)

		if actualGB  > config.G1ThresholdGBs {
			config.GC = "G1"
		} else {
			config.GC = "CMS"
		}
	}
	return config
}

func getMemory(logger lg.Logger) uint64 {
	memSize,err := GetMemory()
	if err != nil {
		logger.ErrorError("Unable to get memory size defaulting to 5GB heap", err)
		memSize = 5000000000
	} else {
		memSize = memSize * 1000
	}
	return memSize
}



func  GetMemory() (uint64, error) {

	var output string

	var command string
	if runtime.GOOS == "linux" {
		command = "/usr/bin/free"
	} else if runtime.GOOS == "darwin" {
		command = "/usr/local/bin/free"
	}
	if out, err := exec.Command(command).Output(); err != nil {
		return 0, err
	} else {
		output = string(out)
	}

	lines := strings.Split(output, "\n")
	line1 := lines[1]

	var total uint64
	var free uint64
	var used uint64
	var shared uint64
	var buffer uint64
	var available uint64
	var mem string

	fmt.Sscanf(line1, "%s %d %d %d %d %d %d", &mem, &total, &used, &free, &shared, &buffer, &available)

	return free, nil

}

func overrideWithEnvOrDefault(envName string, value *string, defaultValue string, logger lg.Logger) {
	envValue := os.Getenv(envName)
	if envValue != "" {
		logger.Debug("Using", envName, "to override", "value=", envValue)
		*value = envValue
	}
	if *value == "" {
		if defaultValue != "" {
			logger.Debug("Using", "default value for", envName, "of", defaultValue)
		}
		*value = defaultValue
	}
}

func overrideNumberWithEnvOrDefault(envName string, value *int, defaultValue int, logger lg.Logger) {
	envValue := os.Getenv(envName)
	if envValue != "" {
		logger.Debug("Using", envName, "to override", "value=", envValue)
		*value, _ = strconv.Atoi(envValue)
	}
	if *value == 0 {
		if defaultValue != 0 {
			logger.Debug("Using", "default value for", envName, "of", defaultValue)
		}
		*value = defaultValue
	}
}

func bindCommandlineArgs(config *Config, logger lg.Logger) {

	flag.BoolVar(&config.Verbose, "v", false, "Turns on verbose mode")

	flag.StringVar(&config.ClusterName, "cluster-name", config.ClusterName,
		"Name of the cluster")

	flag.StringVar(&config.ClusterSeeds, "cluster-seeds", config.ClusterSeeds,
		"Comma delimited list of initial clustrer contact points for bootstrapping")

	flag.StringVar(&config.ClusterListenAddress, "cluster-address", config.ClusterListenAddress,
		"Cluster address for inter-node communication. Example: 192.43.32.10, localhost, etc.")

	flag.StringVar(&config.ClusterListenInterface, "cluster-interface", config.ClusterListenInterface,
		"Cluster interface for inter-node communication.  Example: eth0, eth1, etc.")

	flag.StringVar(&config.ClientListenAddress, "client-address", config.ClientListenAddress,
		"Client address for client driver communication. Example: 192.43.32.10, localhost, etc.")


	flag.StringVar(&config.ReplaceAddress, "-replace-address", config.ReplaceAddress,
		"Replace address used to replace a Cassandra node that has failed or is being replaced.")

	flag.StringVar(&config.ClientListenInterface, "client-interface", config.ClientListenInterface,
		"Client address for client driver communication. Example: eth0, eth1, etc.")

	flag.StringVar(&config.Snitch, "snitch", config.Snitch,
		"Snitch type. Example: GossipingPropertyFileSnitch, PropertyFileSnitch, Ec2Snitch, etc.")

	flag.StringVar(&config.GC, "gc", config.GC,
		"GC type. Values: CMS, G1, or AUTO. If you set to AUTO, if heap is bigger than 5 GB (gc-g1-threshold-gbs), G1 is used, otherwise CMS.")

	flag.IntVar(&config.G1ThresholdGBs, "gc-g1-threshold-gbs", config.G1ThresholdGBs,
		"GC threshold switch. Defaults to 5 GB. If gc set to AUTO, if heap is bigger than gc-g1-threshold-gbs, G1 is used, otherwise CMS.")

	flag.StringVar(&config.JvmOptionsTemplate, "conf-jvm-options-template", config.JvmOptionsTemplate,
		"JVM Option template location. Used to generate the jvm.options file using system ergonomics.")

	flag.StringVar(&config.JvmOptionsFileName, "conf-jvm-options-file", config.JvmOptionsFileName,
		"JVM Option location which will be overwritten with template.")


	flag.StringVar(&config.G1ParallelGCThreads, "g1-parallel-threads", config.G1ParallelGCThreads,
		"The count of G1 Parallel threads. Values: AUTO, or some number. Uses ergonomics to pick a thread count")

	flag.StringVar(&config.G1ConcGCThreads, "g1-concurrent-threads", config.G1ConcGCThreads,
		"The count of G1 Parallel threads. Values: AUTO, or some number. Uses ergonomics to pick a number")

	flag.BoolVar(&config.GCStatsEnabled, "gc_stats_enabled", config.GCStatsEnabled,
		"Enable logging GC stats from JVM.")

	flag.StringVar(&config.CmsYoungGenSize, "cms-young-gen-size", config.CmsYoungGenSize,
		"If using CMS as GC, selects the proper size for the CMS YoungGen. Set this to a specific size of AUTO for environment ergonomics")

	flag.StringVar(&config.MaxHeapSize, "max-heap-size", config.MaxHeapSize,
		"Sets the MaxHeapSize using a size string, i.e., 10GB or uses AUTO to enable system environment ergonomics. (70% of free heap)")

	flag.StringVar(&config.MinHeapSize, "min-heap-size", config.MinHeapSize,
		"Sets the MaxHeapSize using a size string, i.e., 10GB or uses AUTO to enable system environment ergonomics. (Set to MaxHeapSize)")

	flag.StringVar(&config.YamlConfigTemplate, "conf-yaml-template", config.YamlConfigTemplate,
		"Location of cassandra configuration template")

	dataDir := flag.String("data-dirs", "", "Location of Cassandra Data directories")
	help := flag.Bool("help-info", false, "Prints out help information")

	flag.Parse()
	initDataDirectories(config, logger, *dataDir)
	if *help {
		printHelp(config)
	}

}

func printHelp(config *Config) {

	reflected := reflect.ValueOf(*config)
	reflectedType := reflected.Type()

	fmt.Printf("|%-25s |%-15s |%-20s |%-20s |%-30s |%-32s|\n", "Template Var Name", "Type", "Config Name", "Command line", "Environment Variable", "Default Value")
	fmt.Printf("|%-25s |%-15s |%-20s |%-20s |%-30s |%-32s|\n", "---", "---", "---", "---", "---", "---")
	// loop through the struct's fields and set the map
	for i := 0; i < reflected.NumField(); i++ {
		value := reflected.Field(i)
		field := reflectedType.Field(i)

		typeName := field.Type.Name()

		if field.Type.Kind() == reflect.Slice {
			typeName = "[]" + field.Type.Elem().Name()
		}

		tag := field.Tag
		configName := tag.Get("hcl")
		cmdName := "-" + strings.Replace(configName, "_", "-", -1)
		envName := "CASSANDRA_" + strings.ToUpper(configName)
		fmt.Printf("|%-25s |%-15s |%-20s |%-20s |%-30s |%-40v|\n", field.Name, typeName, configName,
			cmdName, envName, value.Interface())

	}
}

func initDataDirectories(config *Config, logger lg.Logger, commandLineOverride string) {
	envValue := os.Getenv("CASSANDRA_DATA_DIRS")
	if envValue != "" {
		logger.Debug("CASSANDRA_DATA_DIRS was set, using it to initialize data dirs", envValue)
		config.DataDirs = strings.Split(envValue, ",")
	}

	if commandLineOverride != "" {
		logger.Debug("Command line argument -data-dirs was set, using it to initialize data dirs", commandLineOverride)
		config.DataDirs = strings.Split(commandLineOverride, ",")
	}

	if len(config.DataDirs) == 0 {
		config.DataDirs = append(config.DataDirs, config.CassandraHome+"/data")
	}

	logger.Debug("Data Directories set to", config.DataDirs)

}
