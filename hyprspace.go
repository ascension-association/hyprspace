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
	"strings"

	execute "github.com/alexellis/go-execute/v2"
	"github.com/gokrazy/gokrazy"
)

var ip = "10.1.1.1"
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

func main() {
	log.Println("Initializing network...")

	// wait for network
	gokrazy.WaitFor("net-online")

	// initialize hyprspace
	if _, err := os.Stat("/perm/hyprspace-config.yaml"); errors.Is(err, os.ErrNotExist) {
		log.Println("Initializing hyprspace...")
		run(false, "/usr/local/bin/busybox", "touch", "/perm/hyprspace-config.yaml")
		run(false, "/usr/local/bin/busybox", "chmod", "600", "/perm/hyprspace-config.yaml")
		run(false, "/usr/local/bin/hyprspace", "init", "utun0", "--config", "/perm/hyprspace-config.yaml")
		run(false, "/usr/local/bin/busybox", "sed", "-i", "s/address: .*/address: 10.1.1.255\\/24/", "/perm/hyprspace-config.yaml")
	}

	// run hyprspace
	if len(id) > 0 {
		log.Println("Checking peer...")

		var found bool = false
		content, _ := os.ReadFile("/perm/hyprspace-config.yaml")
        words := strings.Fields(string(content))
        for _, word := range words {
            if word == id {
                found = true
            }
        }
        if !found {
            run(false, "/usr/local/bin/busybox", "sed", "-i", "s/peers: .*/peers:/", "/perm/hyprspace-config.yaml")
            run(false, "/usr/local/bin/busybox", "echo", "  " + ip + ":", ">>", "/perm/hyprspace-config.yaml")
            run(false, "/usr/local/bin/busybox", "echo", "    id: " + id, ">>", "/perm/hyprspace-config.yaml")
        }

        log.Println("Running hyprspace...")
    	run(true, "/usr/local/bin/hyprspace", "up", "utun0", "--config", "/perm/hyprspace-config.yaml")
	} else {
	    log.Println("No id provided. Exiting...")
	}
}

