package impl

import (
	"os"
	"io/ioutil"
	lg "github.com/advantageous/go-logback/logging"
)

func initYamlTemplate(templateFileName string, logger lg.Logger) {
	if _, err := os.Stat(templateFileName); os.IsNotExist(err) {
		logger.Debug("Cassandra YAML template does not exist so we are creating it", templateFileName)
		err = ioutil.WriteFile(templateFileName, []byte(YamlTemplate), 0644)
		if err != nil {
			logger.ErrorError("Unable to write tempalte file "+templateFileName, err)
		}
	}
}

const YamlTemplate = `

# This file was generated with the template {{.YamlConfigTemplate}} by cassandra-cloud.
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




{{if .ClusterBroadcastAddress}}# Broadcast address for storage cluster communication
broadcast_address: {{.ClusterBroadcastAddress}}{{end}}
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
