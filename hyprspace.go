// Binary hyprspace is a gokrazy wrapper program that runs the bundled hyprspace
// executable in /usr/local/bin/hyprspace after doing any necessary runtime system
// setup.
package main

import (
	"fmt"
	"log"
	"context"
	"os"
	"errors"
	"encoding/json"

	execute "github.com/alexellis/go-execute/v2"
	"github.com/gokrazy/gokrazy"
	"github.com/hyprspace/hyprspace"
)

var name = ""
var id = ""

func run(logging bool, exe string, args ...string) {
	var cmd execute.ExecTask

	if logging {
		cmd = execute.ExecTask{
			Command:     exe,
			Args:        args,
			StreamStdio: true,
		}
	} else {
		cmd = execute.ExecTask{
			Command:     exe,
			Args:        args,
			StreamStdio: false,
			DisableStdioBuffer: true,
		}
	}

	res, err := cmd.Execute(context.Background())

	if err != nil {
		fmt.Errorf("Error: %v", err)
	}

	if res.ExitCode != 0 {
		fmt.Errorf("Error: %v", res.Stderr)
	}
}

type Peer struct {
    Name string `json:"name"`
    ID   string `json:"id"`
}

type Config struct {
	ListenAddresses		[]string	`json:"listenAddresses"`
	PrivateKey			string		`json:"privateKey"`
	Peers				[]Peer  	`json:"peers"`
}

func main() {
	log.Println("Initializing network...")

	// wait for network
	gokrazy.WaitFor("net-online")

	if _, err := os.Stat("/perm/hyprspace.json"); errors.Is(err, os.ErrNotExist) {
		log.Println("Initializing hyprspace...")
		run(true, "/usr/local/bin/hyprspace", "init", "--config", "/perm/hyprspace.json")
	}

	if len(id) > 0 {
		log.Println("Checking peer...")
	    configData, err := os.ReadFile("/perm/hyprspace.json")
	    if err != nil {
            log.Fatalf("Error reading JSON file: %v", err)
        }
        config := Config{}
        err = json.Unmarshal(configData, &config)
        if err != nil {
	        log.Fatalf("Error unmarshaling JSON: %v", err)
        }

        found := false
        for _, peer := range config.Peers {
            if peer.ID == id {
                found = true
            }
        }

        if !found {
            log.Println("Adding peer...")
            peers := config.Peers
            if len(peers) > 0 {
                // append peer
                newPeer := Peer{ Name: name, ID: id }
                peers = append(peers, newPeer)
            } else {
                // add peer
                peers = []Peer{
                    {Name: name, ID: id},
                }
            }

            config.Peers = peers

            configBytes, err := json.MarshalIndent(config, "", "  ")
	        if err != nil {
		        log.Fatalf("Error marshaling updated data: %v", err)
	        }

	        err = os.WriteFile("/perm/hyprspace.json", configBytes, 0644)
	        if err != nil {
		        log.Fatalf("Error writing updated JSON file: %v", err)
	        }

        }

        log.Println("Running hyprspace...")
    	//run(true, "/usr/local/bin/hyprspace", "up", "--config", "/perm/hyprspace.json", "--interface", "hs0")
	} else {
	    log.Println("No peer name/id provided. Exiting...")
	}
}

