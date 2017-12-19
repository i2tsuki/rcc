package main

import (
	"flag"
	"fmt"
	"log"
	// "net"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/go-redis/redis"
	"github.com/kizkoh/rcc/rcc"
	"github.com/pkg/errors"
)

// debug is extended bool and output debug message
type debug bool

func (debug debug) Printf(f string, v ...interface{}) {
	if debug {
		log.Printf(f, v...)
	}
}

// DEBUG is global debug type
var DEBUG debug

func main() {
	var help = false
	var verbose = false

	// parse args
	flags := flag.NewFlagSet(App.Name, flag.ContinueOnError)

	flags.BoolVar(&verbose, "verbose", verbose, "verbose")
	flags.BoolVar(&help, "h", help, "help")
	flags.BoolVar(&help, "help", help, "help")
	flags.BoolVar(&help, "version", help, "version")

	flags.Usage = func() { usage() }
	if err := flags.Parse(os.Args[1:]); err != nil {
		err = errors.Wrap(err, fmt.Sprintf("%v-%v failed: ", App.Name, App.Version))
		fmt.Printf("%v-%v failed: %v\n", App.Name, App.Version, err)
		os.Exit(1)
	}

	if help {
		usage()
		os.Exit(0)
	}

	DEBUG = debug(verbose)

	args := flags.Args()
	var master string
	var slave string
	if len(args) != 2 {
		usage()
		os.Exit(0)
	} else {
		master = args[len(args)-1]
	}
	slave = args[len(args)-2]

	masterClient := redis.NewClient(&redis.Options{
		Addr: master,
	})
	cluster, err := rcc.ClusterNodes(masterClient)
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("%v-%v failed: ", App.Name, App.Version))
		fmt.Fprintf(os.Stderr, "%+v", err)
		os.Exit(1)
	}

	var myself rcc.ClusterNode
	for _, node := range cluster {
		for _, flag := range node.Flags {
			if flag == "myself" {
				myself = node
			}
		}
	}
	masterID := myself.ID
	masterIP := myself.IP
	masterPort := fmt.Sprintf("%d", myself.Port)

	slaveClient := redis.NewClient(&redis.Options{
		Addr: slave,
	})
	// ToDo: Assert new slave node is cluster
	// Assert new slave node is empty
	if err := rcc.AssertEmptyNode(slaveClient); err != nil {
		err = errors.Wrap(err, fmt.Sprintf("%v-%v failed: ", App.Name, App.Version))
		fmt.Fprintf(os.Stderr, "%+v", err)
		os.Exit(1)
	}
	if _, err := slaveClient.ClusterMeet(masterIP, masterPort).Result(); err != nil {
		err = errors.Wrap(err, fmt.Sprintf("%v-%v failed: ", App.Name, App.Version))
		fmt.Fprintf(os.Stderr, "%+v", err)
		os.Exit(1)
	}
	// getConfigSignature := func() {

	// }
	// WaitClusterJoin()
	time.Sleep(5 * time.Second)

	fmt.Printf("configure node as replica of %s\n", master)
	if _, err := slaveClient.ClusterReplicate(masterID).Result(); err != nil {
		err = errors.Wrap(err, fmt.Sprintf("%v-%v failed: ", App.Name, App.Version))
		fmt.Fprintf(os.Stderr, "%+v", err)
		os.Exit(1)
	}
	fmt.Print("new nodes added correctly\n")
}

func usage() {
	helpText := `
usage:
   {{.Name}} [command options]

version:
   {{.Version}}

author:
   kizkoh<GitHub: https://github.com/kizkoh>

options:
   --verbose                                    Print verbose messages
   --help, -h                                   Show help
   --version                                    Print the version
`
	t := template.New("usage")
	t, _ = t.Parse(strings.TrimSpace(helpText))
	t.Execute(os.Stdout, App)
	fmt.Println()
}
