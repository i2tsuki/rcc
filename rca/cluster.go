package rca

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/go-redis/redis"
	"github.com/pkg/errors"
)

// ClusterNode is redis cluster node struct
type ClusterNode struct {
	ID string
	// FIXME: ip addr
	IP          string
	Host        string
	Port        uint64
	Flags       []string
	Slave       bool
	Master      bool
	SlaveOf     string
	PingSent    uint64
	PongRecv    uint64
	ConfigEpoch uint64
	LinkState   string
}

// ClusterNodes provide 'CLUSTER NODES' command result
func ClusterNodes(client *redis.Client) (cluster []ClusterNode, err error) {

	// err := client.Set("key", "value", 0).Err()
	nodes := client.ClusterNodes()

	val, err := nodes.Result()
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("%v-%v failed: ", App.Name, App.Version))
		return nil, err
	}
	val = strings.TrimSpace(val)

	for _, line := range strings.Split(val, "\n") {
		var node ClusterNode
		rows := strings.Split(line, " ")
		// Cluster Node ID
		node.ID = rows[0]
		// Cluster Node IP address
		node.IP = strings.Split(rows[1], ":")[0]
		// Cluster Node host
		hosts, err := net.LookupAddr(node.IP)
		if err != nil || len(hosts) == 0 {
			node.Host = node.IP
		} else {
			node.Host = hosts[0]
		}
		// Cluster Node port number
		port, err := strconv.ParseUint(strings.Split(rows[1], ":")[1], 10, 64)
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("%v-%v failed: ", App.Name, App.Version))
			return nil, err
		}
		node.Port = port
		// Cluster Node state
		flags := strings.Split(rows[2], ",")
		node.Flags = flags
		node.Master = false
		node.Slave = false
		for _, flag := range flags {
			switch flag {
			case "master":
				node.Master = true
			case "slave":
				node.Slave = true
			}
		}
		// Cluster Node slaveof
		node.SlaveOf = rows[3]
		// Append node into cluster
		cluster = append(cluster, node)
	}
	return cluster, nil
}

// DescribeIP return IP, when string is hostname it resolve hostname, or ip return ip, nor returns nil
func DescribeIP(s string) (ip net.IP, err error) {
	ip = net.ParseIP(s)
	if ip != nil {
		addrs, err := net.LookupAddr(s)
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("%v-%v failed: ", App.Name, App.Version))
			return nil, err
		}
		ip = net.ParseIP(addrs[0])
		return ip, nil
	}
	return ip, nil
}

// AssertEmptyNode check node empty and returns nil if node is empty
func AssertEmptyNode(client *redis.Client) (err error) {
	resp, err := client.ClusterInfo().Result()
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("%v-%v failed: ", App.Name, App.Version))
		return err
	}
	resp = strings.TrimSpace(resp)
	for _, line := range strings.Split(resp, "\n") {
		row := strings.Split(line, ":")
		if row[0] == "cluster_known_nodes" {
			value, err := strconv.ParseUint(strings.TrimSpace(row[1]), 10, 64)
			if err != nil {
				err = errors.Wrap(err, fmt.Sprintf("%v-%v failed: ", App.Name, App.Version))
				return err
			}
			if value == 1 {
				resp, err = client.Info("db0").Result()
				if err != nil {
					err = errors.Wrap(err, fmt.Sprintf("%v-%v failed: ", App.Name, App.Version))
					return err
				}
				if resp != "" {
					return errors.New("node is not empty, either the node already knows other nodes (check with CLUSTER NODES) or contains some key in database 0")
				}
			}
			return nil
		}
	}
	return nil
}
