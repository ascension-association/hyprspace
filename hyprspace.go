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
	"io"

	execute "github.com/alexellis/go-execute/v2"
	"github.com/gokrazy/gokrazy"
	
	"github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"
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
		run(false, "/usr/local/bin/busybox", "sed", "-i", "s/address: .*/address: 10.1.1.222\\/24/", "/perm/hyprspace-config.yaml")
	}

	if len(id) > 0 {
		// add peer
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
		log.Println("Adding peer...")
			run(false, "/usr/local/bin/busybox", "sed", "-i", "s/peers: .*/peers:/", "/perm/hyprspace-config.yaml")
			file, _ := os.OpenFile("/perm/hyprspace-config.yaml", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
			file.WriteString("  " + ip + ":\n" + "    id: " + id + "\n")
			file.Close()
		}

		// run hyprspace
		log.Println("Running hyprspace...")
		run(false, "/usr/local/bin/busybox", "sysctl", "-w", "net.core.rmem_max=2048000")
		run(false, "/usr/local/bin/busybox", "sysctl", "-w", "net.core.wmem_max=2048000")
		run(true, "/usr/local/bin/busybox", "grep", "^  id:", "/perm/hyprspace-config.yaml")
		run(true, "/usr/local/bin/hyprspace", "up", "utun0", "--config", "/perm/hyprspace-config.yaml")
		
		// ssh
		ssh.Handle(func(s ssh.Session) {
		    authorizedKey := gossh.MarshalAuthorizedKey(s.PublicKey())
		    io.WriteString(s, fmt.Sprintf("public key used by %s:\n", s.User()))
		    s.Write(authorizedKey)
	    })

	    publicKeyOption := ssh.PublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
		    return true // allow all keys, or use ssh.KeysEqual() to compare against known keys
	    })

	    log.Println("starting ssh server on port 2222...")
	    log.Fatal(ssh.ListenAndServe(":2222", nil, publicKeyOption))
	
	} else {
		log.Println("No id provided. Exiting...")
	}
}
