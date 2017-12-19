package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"

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
	var masterOnly = false
	var help = false
	var verbose = false

	// parse args
	flags := flag.NewFlagSet(App.Name, flag.ContinueOnError)

	flags.BoolVar(&masterOnly, "master", masterOnly, "master")
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
	var arg string
	if len(args) == 0 {
		arg = "127.0.0.1:6379"
	} else {
		arg = args[len(args)-1]
	}

	client := redis.NewClient(&redis.Options{
		Addr: arg,
	})
	cluster, err := rcc.ClusterNodes(client)
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("%v-%v failed: ", App.Name, App.Version))
		fmt.Printf("%v-%v failed: %v\n", App.Name, App.Version, err)
		os.Exit(1)
	}

	nmaster := 0
	for _, master := range cluster {
		if !master.Slave {
			nmaster++
		}
	}
	for _, master := range cluster {
		if !master.Slave {
			nmaster--
			if nmaster > 0 {
				fmt.Print("├─ ")
			} else {
				fmt.Print("└─ ")
			}
			fmt.Printf("%s %s:%d ", master.ID, master.Host, master.Port)
			fmt.Print("[")
			for i, flag := range master.Flags {
				if len(master.Flags)-1 != i {
					fmt.Printf("%s,", flag)
				} else {
					fmt.Printf("%s", flag)
				}
			}
			fmt.Print("] ")
			fmt.Printf("%d %d %d %s %v", master.PingSent, master.PongRecv, master.ConfigEpoch, master.LinkState, master.Slots)
			fmt.Print("\n")

			if !masterOnly {
				nslave := 0
				for _, slave := range cluster {
					if slave.Slave {
						if slave.SlaveOf == master.ID {
							nslave++
						}
					}
				}
				for _, slave := range cluster {
					if slave.Slave {
						if slave.SlaveOf == master.ID {
							nslave--
							if nmaster > 0 {
								fmt.Print("│  ")
							} else {
								fmt.Print("    ")
							}
							if nslave > 0 {
								fmt.Print("├── ")
							} else {
								fmt.Print("└── ")
							}
							fmt.Printf("%s %s:%d ", slave.ID, slave.Host, slave.Port)
							fmt.Print("[")
							for i, flag := range slave.Flags {
								if len(slave.Flags)-1 != i {
									fmt.Printf("%s,", flag)
								} else {
									fmt.Printf("%s", flag)
								}
							}
							fmt.Print("] ")
							fmt.Printf("%d %d %d %s", slave.PingSent, slave.PongRecv, slave.ConfigEpoch, slave.LinkState)
							fmt.Print("\n")
						}
					}
				}
			}
		}
	}

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
