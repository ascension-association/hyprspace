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
)

var name = ""
var id = "test"

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

func loadJSON[T any](filename string) (T, error) {
    var data T
    fileData, err := os.ReadFile(filename)
    if err != nil {
        return data, err
    }
    return data, json.Unmarshal(fileData, &data)
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
	    config, _ := loadJSON("/perm/hyprspace.json")
	    log.Println(config.listenAddresses)
	}

	log.Println("Running hyprspace...")
	run(true, "/usr/local/bin/hyprspace", "up", "--config", "/perm/hyprspace.json", "--interface", "hs0")

}

