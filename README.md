# Cassandra Cloud

***CassandraCloud*** is a tool that helps you configure Cassandra for clustered environments. It works well in 
*Docker*, *EC2*, and *VirtualBox* environments (and similar environments). It allows you to configure Cassandra easily. 
For example, it could be kicked off as a `USER_DATA` script in Amazon EC2 (AWS EC2). ***CassandraCloud*** usually runs 
once when an instance is first launched and then never again. 


## Config Overrides and Templates

***CassandraCloud*** allows you to override values via ***OS ENVIRONMENT*** variables. 
There is an HCL config file, and there are command line arguments. 

The HCL config file can be overridden with **ENVIRONMENT** which can be overridden with command line arguments. 

***CassandraCloud*** will generate `${CASSANDRA_HOME}/conf/cassandra.yaml` file. 
You can specify a custom template (usually found in `${CASSANDRA_HOME}/conf/cassandra-yaml.template`). 

#### Example ${CASSANDRA_HOME}/conf/cassandra-yaml.template
```
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
```

To learn the complete syntax of the template this [template syntax guide](https://golang.org/pkg/text/template/).


The above could generate a cassandra.yaml file as follows:

#### cassandra.yaml
```yaml
cluster_name: "My Cluster"
num_tokens: 32
storage_port: 7000
ssl_storage_port: 7001
native_transport_port: 9042
endpoint_snitch: SimpleSnitch


# Listen network interface for client communication
rpc_interface: localhost



# Listen network interface for storage cluster communication
listen_interface: localhost

max_hints_delivery_threads: 2


data_file_directories:
     - /opt/cassandra/data

commitlog_directory: /opt/cassandra/commitlog


seed_provider:
    - class_name: org.apache.cassandra.locator.SimpleSeedProvider
      parameters:
          - seeds: "127.0.0.1"
```

There are also templates for  `jvm.options`, `cassandra-env.sh`, and `logback.xml`. 

## Usage


```sh
./cassandra-cloud  -h
  -client-address string
        Client address for client driver communication. Example: 192.43.32.10, localhost, etc. (default "localhost")
  -client-interface string
        Client address for client driver communication. Example: eth0, eth1, etc.
  -cluster-address string
        Cluster address for inter-node communication. Example: 192.43.32.10, localhost, etc. (default "localhost")
  -cluster-interface string
        Cluster interface for inter-node communication.  Example: eth0, eth1, etc.
  -cluster-name string
        Name of the cluster (default "My Cluster")
  -cluster-seeds string
        Comma delimited list of initial clustrer contact points for bootstrapping (default "127.0.0.1")
  -cms-young-gen-size string
        If using CMS as GC, selects the proper size for the CMS YoungGen. Set this to a specific size of AUTO for environment ergonomics (default "800MB")
  -conf-jvm-options-file string
        JVM Option location which will be overwritten with template. (default "/opt/cassandra/conf/jvm.options")
  -conf-jvm-options-template string
        JVM Option template location. Used to generate the jvm.options file using system ergonomics. (default "/opt/cassandra/conf/jvm-options.template")
  -conf-yaml-template string
        Location of cassandra configuration template (default "/opt/cassandra/conf/cassandra-yaml.template")
  -config string
        Location of config file
  -data-dirs string
        Location of Cassandra Data directories
  -debug
        Turn on debugging
  -g1-concurrent-threads string
        The count of G1 Parallel threads. Values: AUTO, or some number. Uses ergonomics to pick a number (default "8")
  -g1-parallel-threads string
        The count of G1 Parallel threads. Values: AUTO, or some number. Uses ergonomics to pick a thread count (default "8")
  -gc string
        GC type. Values: CMS, G1, or AUTO. If you set to AUTO, if heap is bigger than 5 GB (gc-g1-threshold-gbs), G1 is used, otherwise CMS. (default "GC1")
  -gc-g1-threshold-gbs int
        GC threshold switch. Defaults to 5 GB. If gc set to AUTO, if heap is bigger than gc-g1-threshold-gbs, G1 is used, otherwise CMS. (default 5)
  -gc_stats_enabled
        Enable logging GC stats from JVM.
  -help-info
        Prints out help information
  -max-heap-size string
        Sets the MaxHeapSize using a size string, i.e., 10GB or uses AUTO to enable system environment ergonomics. (70% of free heap) (default "4859MB")
  -min-heap-size string
        Sets the MaxHeapSize using a size string, i.e., 10GB or uses AUTO to enable system environment ergonomics. (Set to MaxHeapSize) (default "4859MB")
  -snitch string
        Snitch type. Example: GossipingPropertyFileSnitch, PropertyFileSnitch, Ec2Snitch, etc. (default "SimpleSnitch")
  -v    Turns on verbose mode

```

Command line flag syntax:
```
-flag
-flag=x
-flag x  // non-boolean flags only
```

## Configuration

#### Cloud conf usually found in ${CASSANDRA_HOME}/conf/cloud.conf
```conf

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
```

#### Template variable (types, and how to override them) 

The table below lists the variable names. 
A setting in the config file (`Config Name`) overrides the default value(`Default Value`). 
`Environment Variable`s override config file settings  (`Config Name`). 
Lastly values passed on the command line (`Command line Args`) override the config. 

The `Template Var Name`s are used by the templates. 

For example the `CommitLogDir` template var can be used in a template by referring to it as `{{.CommitLogDir}}`.


|Template Var Name         |Type            |Config Name          |Command line         |Environment Variable           |Default Value                   |
|---                       |---             |---                  |---                  |---                            |---                             |
|DataDirs                  |[]string        |data_dirs            |-data-dirs           |CASSANDRA_DATA_DIRS            |[/opt/cassandra/data                     ]|
|CassandraHome             |string          |home_dir             |-home-dir            |CASSANDRA_HOME_DIR             |/opt/cassandra                          |
|ClusterSeeds              |string          |cluster_seeds        |-cluster-seeds       |CASSANDRA_CLUSTER_SEEDS        |127.0.0.1                               |
|ClusterListenAddress      |string          |cluster_address      |-cluster-address     |CASSANDRA_CLUSTER_ADDRESS      |localhost                               |
|ClusterListenInterface    |string          |cluster_interface    |-cluster-interface   |CASSANDRA_CLUSTER_INTERFACE    |                                        |
|ClientListenAddress       |string          |client_address       |-client-address      |CASSANDRA_CLIENT_ADDRESS       |localhost                               |
|ClientListenInterface     |string          |client_interface     |-client-interface    |CASSANDRA_CLIENT_INTERFACE     |                                        |
|ClientPort                |int             |client_port          |-client-port         |CASSANDRA_CLIENT_PORT          |9042                                    |
|ClusterName               |string          |cluster_name         |-cluster-name        |CASSANDRA_CLUSTER_NAME         |My Cluster                              |
|ClusterPort               |int             |cluster_port         |-cluster-port        |CASSANDRA_CLUSTER_PORT         |7000                                    |
|ClusterSslPort            |int             |cluster_ssl_port     |-cluster-ssl-port    |CASSANDRA_CLUSTER_SSL_PORT     |7001                                    |
|CmsYoungGenSize           |string          |cms_young_gen_size   |-cms-young-gen-size  |CASSANDRA_CMS_YOUNG_GEN_SIZE   |800MB                                   |
|CommitLogDir              |string          |commit_log_dir       |-commit-log-dir      |CASSANDRA_COMMIT_LOG_DIR       |/opt/cassandra/commitlog                |
|GCStatsEnabled            |bool            |gc_stats_enabled     |-gc-stats-enabled    |CASSANDRA_GC_STATS_ENABLED     |false                                   |
|GC                        |string          |gc                   |-gc                  |CASSANDRA_GC                   |GC1  if over 5GB heap free                                   |
|G1ThresholdGBs            |int             |gc_g1_threshold_gbs  |-gc-g1-threshold-gbs |CASSANDRA_GC_G1_THRESHOLD_GBS  |5                                       |
|G1ParallelGCThreads       |string          |g1_parallel_threads  |-g1-parallel-threads |CASSANDRA_G1_PARALLEL_THREADS  |8                                       |
|G1ConcGCThreads           |string          |g1_concurrent_threads |-g1-concurrent-threads |CASSANDRA_G1_CONCURRENT_THREADS |8                                       |
|JvmOptionsFileName        |string          |conf_jvm_options_file |-conf-jvm-options-file |CASSANDRA_CONF_JVM_OPTIONS_FILE |/opt/cassandra/conf/jvm.options         |
|JvmOptionsTemplate        |string          |conf_jvm_options_template |-conf-jvm-options-template |CASSANDRA_CONF_JVM_OPTIONS_TEMPLATE |/opt/cassandra/conf/jvm-options.template|
|MinHeapSize               |string          |min_heap_size        |-min-heap-size       |CASSANDRA_MIN_HEAP_SIZE        |4859MB                                  |
|MaxHeapSize               |string          |max_heap_size        |-max-heap-size       |CASSANDRA_MAX_HEAP_SIZE        |4859MB                                  |
|MultiDataCenter           |bool            |multi_dc             |-multi-dc            |CASSANDRA_MULTI_DC             |false                                   |
|NumTokens                 |int             |num_tokens           |-num-tokens          |CASSANDRA_NUM_TOKENS           |32                                      |
|Snitch                    |string          |snitch               |-snitch              |CASSANDRA_SNITCH               |SimpleSnitch                            |
|Verbose                   |bool            |verbose              |-verbose             |CASSANDRA_VERBOSE              |false                                   |
|YamlConfigTemplate        |string          |conf_yaml_template   |-conf-yaml-template  |CASSANDRA_CONF_YAML_TEMPLATE   |/opt/cassandra/conf/cassandra-yaml.template|
|YamlConfigFileName        |string          |conf_yaml_file       |-conf-yaml-file      |CASSANDRA_CONF_YAML_FILE       |/opt/cassandra/conf/cassandra.yaml      |

## About us
[Cloudurable](http://cloudurable.com/) provides AMIs, cloudformation templates and monitoring tools 
to support [Cassandra in production running in EC2](http://cloudurable.com/services/index.html). 
We also teach advanced [Cassandra courses which teaches how one could develop, support and deploy Cassandra to production in AWS EC2](http://cloudurable.com/services/index.html). 

## More details to follow
