package impl

import (
	"os"
	"io/ioutil"
	lg "github.com/advantageous/go-logback/logging"
)

func initJvmOptionsTemplate(templateFileName string, logger lg.Logger) {
	if _, err := os.Stat(templateFileName); os.IsNotExist(err) {
		logger.Debug("Cassandra JVM Option template does not exist so we are creating it", templateFileName)
		err = ioutil.WriteFile(templateFileName, []byte(JvmOptionsTemplate), 0644)
		if err != nil {
			logger.ErrorError("Unable to write tempalte file "+templateFileName, err)
		}
	}
}

const JvmOptionsTemplate = `

# To replace a node that has died, restart a new node in its place specifying the address of the
# dead node. The new node must not have any data in its data directory, that is, it must be in the
# same state as before bootstrapping.

{{if .ReplaceAddress}}# Replacing address
-Dcassandra.replace_address={{.ReplaceAddress}}
{{end}}

#-Dcassandra.join_ring=true|false
# Set to false to clear all gossip state for the node on restart. Use when you have changed node
# information in cassandra.yaml (such as listen_address).
#-Dcassandra.load_ring_state=true|false

## All by default
#-Dcassandra.available_processors=number_of_processors

# The directory location of the cassandra.yaml file.
#-Dcassandra.config=directory

# Enable pluggable metrics reporter. See Pluggable metrics reporting in Cassandra 2.0.2.
#-Dcassandra.metricsReporterConfigFile=file

# Allow restoring specific tables from an archived commit log.
#-Dcassandra.replayList=table

# Set the default location for the trigger JARs. (Default: conf/triggers)
#-Dcassandra.triggers_dir=directory

# For testing new compaction and compression strategies. It allows you to experiment with different
# strategies and benchmark write performance differences without affecting the production workload.
#-Dcassandra.write_survey=true

########################
# GENERAL JVM SETTINGS #
########################

# enable assertions.  disabling this in production will give a modest
# performance benefit (around 5%).
# -ea

# enable thread priorities, primarily so we can give periodic tasks
# a lower priority to avoid interfering with client workload
-XX:+UseThreadPriorities

# allows lowering thread priority without being root on linux - probably
# not necessary on Windows but doesn't harm anything.
# see http://tech.stolsvik.com/2010/01/linux-java-thread-priorities-workar
-XX:ThreadPriorityPolicy=42

# Enable heap-dump if there's an OOM
-XX:+HeapDumpOnOutOfMemoryError

# Per-thread stack size.
-Xss256k

# Larger interned string table, for gossip's benefit (CASSANDRA-6410)
-XX:StringTableSize=1000003

# Make sure all memory is faulted and zeroed on startup.
# This helps prevent soft faults in containers and makes
# transparent hugepage allocation more effective.
-XX:+AlwaysPreTouch

# Disable biased locking as it does not benefit Cassandra.
-XX:-UseBiasedLocking

# Enable thread-local allocation blocks and allow the JVM to automatically
# resize them at runtime.
-XX:+UseTLAB
-XX:+ResizeTLAB

# http://www.evanjones.ca/jvm-mmap-pause.html
-XX:+PerfDisableSharedMem

# Prefer binding to IPv4 network intefaces (when net.ipv6.bindv6only=1). See
# http://bugs.sun.com/bugdatabase/view_bug.do?bug_id=6342561 (short version:
# comment out this entry to enable IPv6 support).
-Djava.net.preferIPv4Stack=true

### Debug options

# uncomment to enable flight recorder
#-XX:+UnlockCommercialFeatures
#-XX:+FlightRecorder

# uncomment to have Cassandra JVM listen for remote debuggers/profilers on port 1414
#-agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=1414

# uncomment to have Cassandra JVM log internal method compilation (developers only)
#-XX:+UnlockDiagnosticVMOptions
#-XX:+LogCompilation

#################
# HEAP SETTINGS #
#################

# Heap size is automatically calculated by cassandra-env based on this
# formula: max(min(1/2 ram, 1024MB), min(1/4 ram, 8GB))
# That is:
# - calculate 1/2 ram and cap to 1024MB
# - calculate 1/4 ram and cap to 8192MB
# - pick the max
#
# For production use you may wish to adjust this for your environment.
# If that's the case, uncomment the -Xmx and Xms options below to override the
# automatic calculation of JVM heap memory.
#
# It is recommended to set min (-Xms) and max (-Xmx) heap sizes to
# the same value to avoid stop-the-world GC pauses during resize, and
# so that we can lock the heap in memory on startup to prevent any
# of it from being swapped out.
-Xms{{.MinHeapSize}}
-Xmx{{.MaxHeapSize}}


#################
#  GC SETTINGS  #
#################

## Setting  up {{.GC}}


{{if .GC eq "CMS"}}
### CMS Settings
# Young generation size is automatically calculated by cassandra-env
# based on this formula: min(100 * num_cores, 1/4 * heap size)
#
# The main trade-off for the young generation is that the larger it
# is, the longer GC pause times will be. The shorter it is, the more
# expensive GC will be (usually).
#
# It is not recommended to set the young generation size if using the
# G1 GC, since that will override the target pause-time goal.
# More info: http://www.oracle.com/technetwork/articles/java/g1gc-1984535.html
#
# The example below assumes a modern 8-core+ machine for decent
# times. If in doubt, and if you do not particularly want to tweak, go
# 100 MB per physical CPU core.
-Xmn{{.CmsYoungGenSize}}
-XX:+UseParNewGC
-XX:+UseConcMarkSweepGC
-XX:+CMSParallelRemarkEnabled
-XX:SurvivorRatio=8
-XX:MaxTenuringThreshold=1
-XX:CMSInitiatingOccupancyFraction=75
-XX:+UseCMSInitiatingOccupancyOnly
-XX:CMSWaitDuration=10000
-XX:+CMSParallelInitialMarkEnabled
-XX:+CMSEdenChunksRecordAlways
# some JVMs will fill up their heap when accessed via JMX, see CASSANDRA-6541
-XX:+CMSClassUnloadingEnabled
{{end}}

{{if eq .GC "G1" }}
## Use the Hotspot garbage-first collector.
-XX:+UseG1GC
#
## Have the JVM do less remembered set work during STW, instead
## preferring concurrent GC. Reduces p99.9 latency.
-XX:G1RSetUpdatingPauseTimePercent=5
#
## Main G1GC tunable: lowering the pause target will lower throughput and vise versa.
## 200ms is the JVM default and lowest viable setting
## 1000ms increases throughput. Keep it smaller than the timeouts in cassandra.yaml.
-XX:MaxGCPauseMillis=500
# Save CPU time on large (>= 16GB) heaps by delaying region scanning
# until the heap is 70% full. The default in Hotspot 8u40 is 40%.
-XX:InitiatingHeapOccupancyPercent=70

# For systems with > 8 cores, the default ParallelGCThreads is 5/8 the number of logical cores.
# Otherwise equal to the number of cores when 8 or less.
# Machines with > 10 cores should try setting these to <= full cores.
-XX:ParallelGCThreads={{.G1ParallelGCThreads}}
# By default, ConcGCThreads is 1/4 of ParallelGCThreads.
# Setting both to the same value can reduce STW durations.
-XX:ConcGCThreads={{.G1ConcGCThreads}}
{{end}}



### GC logging options -- uncomment to enable
{{if .GCStatsEnabled}}# Turn on GC stats
-XX:+PrintGCDetails
-XX:+PrintGCDateStamps
-XX:+PrintHeapAtGC
-XX:+PrintTenuringDistribution
-XX:+PrintGCApplicationStoppedTime
-XX:+PrintPromotionFailure
#-XX:PrintFLSStatistics=1
-Xloggc:{{.CassandraHome}}/log/gc.log
-XX:+UseGCLogFileRotation
-XX:NumberOfGCLogFiles=10
-XX:GCLogFileSize=10M
{{end}}

`
