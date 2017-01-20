package main

import "fmt"
import (
	"flag"
	lg "github.com/advantageous/go-logback/logging"
	"github.com/hashicorp/hcl"
	"io/ioutil"
	"os"
	"runtime"
	"strconv"
	"strings"
	"text/template"
	"reflect"
)

const defaultConfig = `

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

type Config struct {
	Verbose bool `hcl:"verbose"`
	// Addresses of hosts that are deemed contact points.
	// Cassandra nodes use this list of hosts to find each other and learn
	// the topology of the ring.  You must change this if you are running  multiple nodes!
	ClusterSeeds string `hcl:"cluster_seeds"`

	CassandraHome string `hcl:"home_dir"`

	MultiDataCenter bool `hcl:"multi_dc"`

	DataDirs []string `hcl:"data_dirs"`

	CommitLogDir string `hcl:"commit_log_dir"`

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
	// Cassandra snitch type.
	Snitch string `hcl:"snitch"`
	//Name of the cassandra cluster.
	ClusterName string `hcl:"cluster_name"`
	//Number of tokens that this node wants/has. Used for Cassandra VNODES.
	NumTokens int `hcl:"num_tokens"`

	//Location of template file for cassandra conf.
	CassandraConfigTemplate string `hcl:"conf_yaml_template"`

	//Location of cassandra yaml config file.
	CassandraConfigFileName string `hcl:"conf_yaml_file"`

	ClusterPort    int `hcl:"cluster_port"`
	ClusterSslPort int `hcl:"cluster_ssl_port"`

	ClientPort int `hcl:"client_port"`
}

func main() {

	debug, configFilename, logger := initialCommandLineParse()
	initConfigFile(configFilename, logger)

	config, err := LoadConfig(configFilename, debug, logger)

	if err != nil {
		logger.Errorf("Unable to load config filename %s  \n", configFilename)
		logger.ErrorError("Error was", err)
		os.Exit(1)
	}

	if config.Verbose {
		displayConfig(config)
	}

	ProcessTemplate(config.CassandraConfigTemplate, config.CassandraConfigFileName, config, logger)

}

func printHelp(config *Config) {


	reflected:=reflect.ValueOf(*config)
	reflectedType:=reflected.Type()


	fmt.Printf("|%-25s |%-15s |%-20s |%-20s |%-30s |%-32s|\n", "Template Var Name", "Type", "Config Name", "Command line", "Environment Variable", "Default Value")
	fmt.Printf("|%-25s |%-15s |%-20s |%-20s |%-30s |%-32s|\n", "---", "---", "---", "---", "---", "---")
	// loop through the struct's fields and set the map
	for i := 0; i < reflected.NumField(); i++ {
		value := reflected.Field(i)
		field := reflectedType.Field(i)


		typeName:=field.Type.Name()

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

func displayConfig(config *Config) {

	reflected:=reflect.ValueOf(*config)
	reflectedType:=reflected.Type()


	fmt.Printf("%-30s %-20s %v\n", "Field Name", "Type", "Value")

	// loop through the struct's fields and set the map
	for i := 0; i < reflected.NumField(); i++ {
		value := reflected.Field(i)
		field := reflectedType.Field(i)

		fmt.Printf("%-30s %-20s %v\n", field.Name, field.Type.Name(), value.Interface())

	}

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

func ProcessTemplate(inputFileName string, outputFileName string, any interface{}, logger lg.Logger) error {
	bytes, err := ioutil.ReadFile(inputFileName)
	if err != nil {
		logger.Errorf("Unable to load template %s  \n", inputFileName)
		logger.ErrorError("Error was", err)
		return err
	}

	template, err := template.New("test").Parse(string(bytes))
	if err != nil {
		logger.Errorf("Unable to parse template %s  \n", inputFileName)
		logger.ErrorError("Error was", err)
		return err
	}

	outputFile, err := os.Create(outputFileName)
	if err != nil {
		logger.ErrorError(fmt.Sprintf("Unable to open output file %s", outputFileName), err)
		return err
	}
	defer outputFile.Close()
	template.Execute(outputFile, any)
	return nil
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

func initDefaults(config *Config, logger lg.Logger) {

	overrideWithEnvOrDefault("CASSANDRA_CLUSTER_NAME", &config.ClusterName, "mycluster", logger)
	overrideWithEnvOrDefault("CASSANDRA_HOME_DIR", &config.CassandraHome, "/opt/cassandra", logger)
	overrideWithEnvOrDefault("CASSANDRA_YAML_TEMPLATE", &config.CassandraConfigTemplate,
		config.CassandraHome+"/conf/cassandra-yaml.template", logger)
	overrideWithEnvOrDefault("CASSANDRA_YAML_FILE", &config.CassandraConfigFileName,
		config.CassandraHome+"/conf/cassandra.yaml", logger)
	overrideWithEnvOrDefault("CASSANDRA_SNITCH", &config.Snitch, "SimpleSnitch", logger)
	overrideWithEnvOrDefault("CASSANDRA_CLUSTER_SEEDS", &config.ClusterSeeds, "127.0.0.1", logger)
	overrideWithEnvOrDefault("CASSANDRA_CLIENT_INTERFACE", &config.ClientListenInterface, "", logger)
	overrideWithEnvOrDefault("CASSANDRA_CLIENT_ADDRESS", &config.ClientListenAddress, "", logger)
	overrideWithEnvOrDefault("CASSANDRA_CLUSTER_INTERFACE", &config.ClusterListenInterface, "", logger)
	overrideWithEnvOrDefault("CASSANDRA_CLUSTER_ADDRESS", &config.ClusterListenAddress, "", logger)
	overrideWithEnvOrDefault("CASSANDRA_COMMIT_LOG_DIR", &config.CommitLogDir, config.CassandraHome+"/commitlog", logger)

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


	initYamlTemplate(config.CassandraConfigTemplate, logger)


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

func overrideWithEnvOrDefault(envName string, value *string, defaultValue string, logger lg.Logger) {
	envValue := os.Getenv(envName)
	if envValue != "" {
		logger.Debug("Using", envName, "to override", "value=", envValue)
		*value = envValue
	}
	if *value == "" {
		if defaultValue!= "" {
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
		if defaultValue!= 0 {
			logger.Debug("Using", "default value for", envName, "of", defaultValue)
		}
		*value = defaultValue
	}
}

func initYamlTemplate(templateFileName string, logger lg.Logger) {
	if _, err := os.Stat(templateFileName); os.IsNotExist(err) {
		logger.Debug("Cassandra YAML template does not exist so we are creating it", templateFileName)
		err = ioutil.WriteFile(templateFileName, []byte(yamlTemplate), 0644)
		if err != nil {
			logger.ErrorError("Unable to write tempalte file "+templateFileName, err)
		}
	}
}

func initConfigFile(configFileName string, logger lg.Logger) {
	if _, err := os.Stat(configFileName); os.IsNotExist(err) {
		logger.Debug("Cloud Cassandra config file does not exist so we are creating it", configFileName)
		err = ioutil.WriteFile(configFileName, []byte(defaultConfig), 0644)
		if err != nil {
			logger.ErrorError("Unable to write config file "+configFileName, err)
		}
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

	flag.StringVar(&config.ClientListenInterface, "client-interface", config.ClientListenInterface,
		"Client address for client driver communication. Example: eth0, eth1, etc.")

	flag.StringVar(&config.Snitch, "snitch", config.Snitch,
		"Snitch type. Example: GossipingPropertyFileSnitch, PropertyFileSnitch, Ec2Snitch, etc.")

	flag.StringVar(&config.CassandraConfigTemplate, "conf-yaml-template", config.CassandraConfigTemplate,
		"Location of cassandra configuration template")

	dataDir := flag.String("data-dirs", "", "Location of Cassandra Data directories")
	help:=flag.Bool("help-info", false, "Prints out help information")


	flag.Parse()
	initDataDirectories(config, logger, *dataDir)
	if *help {
		printHelp(config)
	}





}


const yamlTemplate = `

# This file was generated with the template {{.CassandraConfigTemplate}} by cassandra-cloud.
# You can find cassandra-cloud at https://github.com/cloudurable/cassandra-cloud.
# Cassandra Cloud is used to automate deployment to EC2 and similar cloud environments.

cluster_name: "{{.ClusterName}}"
num_tokens: {{.NumTokens}}
storage_port: {{.ClusterPort}}
ssl_storage_port: {{.ClusterSslPort}}
native_transport_port: {{.ClientPort}}
endpoint_snitch: {{.Snitch}}

{{if .ClientListenAddress}}# Listen address for client communication
rpc_address: {{.ClientListenAddress}}{{end}}
{{if .ClientListenInterface}}# Listen network interface for client communication
rpc_interface: {{.ClientListenInterface}}{{end}}


{{if .ClusterListenAddress}}# Listen address for storage cluster communication
listen_address: {{.ClusterListenAddress}}{{end}}
{{if .ClusterListenInterface}}# Listen network interface for storage cluster communication
listen_interface: {{.ClusterListenInterface}}{{end}}

{{if .MultiDataCenter}}max_hints_delivery_threads: 16{{else}}max_hints_delivery_threads: 2{{end}}


data_file_directories:
{{range .DataDirs}}     - {{.}}{{end}}

commitlog_directory: {{.CommitLogDir}}


seed_provider:
    - class_name: org.apache.cassandra.locator.SimpleSeedProvider
      parameters:
          - seeds: "{{.ClusterSeeds}}"



# Use ergonomics with these

# For workloads with more data than can fit in memory, Cassandra's
# bottleneck will be reads that need to fetch data from
# disk. "concurrent_reads" should be set to (16 * number_of_drives) in
# order to allow the operations to enqueue low enough in the stack
# that the OS and drives can reorder them. Same applies to
# "concurrent_counter_writes", since counter writes read the current
# values before incrementing and writing them back.
#
# On the other hand, since writes are almost never IO bound, the ideal
# number of "concurrent_writes" is dependent on the number of cores in
# your system; (8 * number_of_cores) is a good rule of thumb.
concurrent_reads: 32
concurrent_writes: 32
concurrent_counter_writes: 32

# If your data directories are backed by SSD, you should increase this
# to the number of cores.
#concurrent_compactors: 1

# stream_throughput_outbound_megabits_per_sec: 200



trickle_fsync: true
trickle_fsync_interval_in_kb: 10240

hinted_handoff_enabled: true
max_hint_window_in_ms: 10800000 # 3 hours
hinted_handoff_throttle_in_kb: 1024







hints_flush_period_in_ms: 10000
max_hints_file_size_in_mb: 128
batchlog_replay_throttle_in_kb: 1024
prepared_statements_cache_size_mb:
thrift_prepared_statements_cache_size_mb:
partitioner: org.apache.cassandra.dht.Murmur3Partitioner
key_cache_size_in_mb:
key_cache_save_period: 14400
# key_cache_keys_to_save: 100
# row_cache_class_name: org.apache.cassandra.cache.OHCProvider
row_cache_size_in_mb: 0
row_cache_save_period: 0
# row_cache_keys_to_save: 100
counter_cache_size_in_mb:
counter_cache_save_period: 7200
# counter_cache_keys_to_save: 100
# saved_caches_directory: /var/lib/cassandra/saved_caches
commitlog_sync: periodic
commitlog_sync_period_in_ms: 10000
commitlog_segment_size_in_mb: 16
disk_optimization_strategy: ssd


cdc_enabled: false
# cdc_raw_directory: /var/lib/cassandra/cdc_raw

## Security
authenticator: AllowAllAuthenticator
authorizer: AllowAllAuthorizer
role_manager: CassandraRoleManager
roles_validity_in_ms: 2000
# roles_update_interval_in_ms: 2000
permissions_validity_in_ms: 2000
# permissions_update_interval_in_ms: 2000
credentials_validity_in_ms: 2000
# credentials_update_interval_in_ms: 2000

## Failure modes
disk_failure_policy: stop
commit_failure_policy: stop


concurrent_materialized_view_writes: 32
memtable_allocation_type: offheap_objects
index_summary_capacity_in_mb:
index_summary_resize_interval_in_minutes: 60

# listen_on_broadcast_address: false
# internode_authenticator: org.apache.cassandra.auth.AllowAllInternodeAuthenticator

start_native_transport: true


# native_transport_port_ssl: 9142
# native_transport_max_threads: 128
# native_transport_max_frame_size_in_mb: 256
# native_transport_max_concurrent_connections: -1
# native_transport_max_concurrent_connections_per_ip: -1
start_rpc: false
rpc_port: 9160
# broadcast_rpc_address: 1.2.3.4
# enable or disable keepalive on rpc/native connections
rpc_keepalive: true
rpc_server_type: sync

# Uncomment to set socket buffer size for internode communication
# Note that when setting this, the buffer size is limited by net.core.wmem_max
# and when not setting it it is defined by net.ipv4.tcp_wmem
# See also:
# /proc/sys/net/core/wmem_max
# /proc/sys/net/core/rmem_max
# /proc/sys/net/ipv4/tcp_wmem
# /proc/sys/net/ipv4/tcp_wmem
# and 'man tcp'
# internode_send_buff_size_in_bytes:

# Uncomment to set socket buffer size for internode communication
# Note that when setting this, the buffer size is limited by net.core.wmem_max
# and when not setting it it is defined by net.ipv4.tcp_wmem
# internode_recv_buff_size_in_bytes:

# Frame size for thrift (maximum message length).
thrift_framed_transport_size_in_mb: 15

incremental_backups: false
snapshot_before_compaction: false
auto_snapshot: true

column_index_size_in_kb: 64
column_index_cache_size_in_kb: 2


compaction_throughput_mb_per_sec: 16
sstable_preemptive_open_interval_in_mb: 50

# inter_dc_stream_throughput_outbound_megabits_per_sec: 200

read_request_timeout_in_ms: 5000
range_request_timeout_in_ms: 10000
write_request_timeout_in_ms: 2000
counter_write_request_timeout_in_ms: 5000
cas_contention_timeout_in_ms: 1000
truncate_request_timeout_in_ms: 60000
request_timeout_in_ms: 10000
cross_node_timeout: true

# Set socket timeout for streaming operation.
# The stream session is failed if no data/ack is received by any of the participants
# within that period, which means this should also be sufficient to stream a large
# sstable or rebuild table indexes.
# Default value is 86400000ms, which means stale streams timeout after 24 hours.
# A value of zero means stream sockets should never time out.
# streaming_socket_timeout_in_ms: 86400000

# phi value that must be reached for a host to be marked down.
# most users should never need to adjust this.
# phi_convict_threshold: 8



dynamic_snitch_update_interval_in_ms: 100
dynamic_snitch_reset_interval_in_ms: 60000
dynamic_snitch_badness_threshold: 0.15
request_scheduler: org.apache.cassandra.scheduler.NoScheduler

# request_scheduler_id: keyspace

# Enable or disable inter-node encryption
# JVM defaults for supported SSL socket protocols and cipher suites can
# be replaced using custom encryption options. This is not recommended
# unless you have policies in place that dictate certain settings, or
# need to disable vulnerable ciphers or protocols in case the JVM cannot
# be updated.
# FIPS compliant settings can be configured at JVM level and should not
# involve changing encryption settings here:
# https://docs.oracle.com/javase/8/docs/technotes/guides/security/jsse/FIPS.html
# *NOTE* No custom encryption options are enabled at the moment
# The available internode options are : all, none, dc, rack
#
# If set to dc cassandra will encrypt the traffic between the DCs
# If set to rack cassandra will encrypt the traffic between the racks
#
# The passwords used in these options must match the passwords used when generating
# the keystore and truststore.  For instructions on generating these files, see:
# http://download.oracle.com/javase/6/docs/technotes/guides/security/jsse/JSSERefGuide.html#CreateKeystore
#
server_encryption_options:
    internode_encryption: none
    keystore: conf/.keystore
    keystore_password: cassandra
    truststore: conf/.truststore
    truststore_password: cassandra
    # More advanced defaults below:
    # protocol: TLS
    # algorithm: SunX509
    # store_type: JKS
    # cipher_suites: [TLS_RSA_WITH_AES_128_CBC_SHA,TLS_RSA_WITH_AES_256_CBC_SHA,TLS_DHE_RSA_WITH_AES_128_CBC_SHA,TLS_DHE_RSA_WITH_AES_256_CBC_SHA,TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA]
    # require_client_auth: false
    # require_endpoint_verification: false

# enable or disable client/server encryption.
client_encryption_options:
    enabled: false
    # If enabled and optional is set to true encrypted and unencrypted connections are handled.
    optional: false
    keystore: conf/.keystore
    keystore_password: cassandra
    # require_client_auth: false
    # Set trustore and truststore_password if require_client_auth is true
    # truststore: conf/.truststore
    # truststore_password: cassandra
    # More advanced defaults below:
    # protocol: TLS
    # algorithm: SunX509
    # store_type: JKS
    # cipher_suites: [TLS_RSA_WITH_AES_128_CBC_SHA,TLS_RSA_WITH_AES_256_CBC_SHA,TLS_DHE_RSA_WITH_AES_128_CBC_SHA,TLS_DHE_RSA_WITH_AES_256_CBC_SHA,TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA]

internode_compression: dc

# Enable or disable tcp_nodelay for inter-dc communication.
# Disabling it will result in larger (but fewer) network packets being sent,
# reducing overhead from the TCP protocol itself, at the cost of increasing
# latency if you block for cross-datacenter responses.
inter_dc_tcp_nodelay: false

# TTL for different trace types used during logging of the repair process.
tracetype_query_ttl: 86400
tracetype_repair_ttl: 604800

# By default, Cassandra logs GC Pauses greater than 200 ms at INFO level
# This threshold can be adjusted to minimize logging if necessary
# gc_log_threshold_in_ms: 200

enable_user_defined_functions: false
enable_scripted_user_defined_functions: false
windows_timer_interval: 1


# Enables encrypting data at-rest (on disk). Different key providers can be plugged in, but the default reads from
# a JCE-style keystore. A single keystore can hold multiple keys, but the one referenced by
# the "key_alias" is the only key that will be used for encrypt opertaions; previously used keys
# can still (and should!) be in the keystore and will be used on decrypt operations
# (to handle the case of key rotation).
#
# It is strongly recommended to download and install Java Cryptography Extension (JCE)
# Unlimited Strength Jurisdiction Policy Files for your version of the JDK.
# (current link: http://www.oracle.com/technetwork/java/javase/downloads/jce8-download-2133166.html)
#
# Currently, only the following file types are supported for transparent data encryption, although
# more are coming in future cassandra releases: commitlog, hints
transparent_data_encryption_options:
    enabled: false
    chunk_length_kb: 64
    cipher: AES/CBC/PKCS5Padding
    key_alias: testing:1
    # CBC IV length for AES needs to be 16 bytes (which is also the default size)
    # iv_length: 16
    key_provider:
      - class_name: org.apache.cassandra.security.JKSKeyProvider
        parameters:
          - keystore: conf/.keystore
            keystore_password: cassandra
            store_type: JCEKS
            key_password: cassandra


#####################
# SAFETY THRESHOLDS #
#####################

# When executing a scan, within or across a partition, we need to keep the
# tombstones seen in memory so we can return them to the coordinator, which
# will use them to make sure other replicas also know about the deleted rows.
# With workloads that generate a lot of tombstones, this can cause performance
# problems and even exaust the server heap.
# (http://www.datastax.com/dev/blog/cassandra-anti-patterns-queues-and-queue-like-datasets)
# Adjust the thresholds here if you understand the dangers and want to
# scan more tombstones anyway.  These thresholds may also be adjusted at runtime
# using the StorageService mbean.
tombstone_warn_threshold: 1000
tombstone_failure_threshold: 100000

# Log WARN on any batch size exceeding this value. 5kb per batch by default.
# Caution should be taken on increasing the size of this threshold as it can lead to node instability.
batch_size_warn_threshold_in_kb: 5

# Fail any batch exceeding this value. 50kb (10x warn threshold) by default.
batch_size_fail_threshold_in_kb: 50

# Log WARN on any batches not of type LOGGED than span across more partitions than this limit
unlogged_batch_across_partitions_warn_threshold: 10

# Log a warning when compacting partitions larger than this value
compaction_large_partition_warning_threshold_mb: 100

# GC Pauses greater than gc_warn_threshold_in_ms will be logged at WARN level
# Adjust the threshold based on your application throughput requirement
# By default, Cassandra logs GC Pauses greater than 200 ms at INFO level
gc_warn_threshold_in_ms: 1000

# Maximum size of any value in SSTables. Safety measure to detect SSTable corruption
# early. Any value size larger than this threshold will result into marking an SSTable
# as corrupted.
# max_value_size_in_mb: 256


`
