// Binary hyprspace is a gokrazy wrapper program that runs the bundled hyprspace
// executable in /usr/local/bin/hyprspace after doing any necessary runtime system
// setup.
package main

import (
	"errors"
	"os/exec"
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"syscall"
	"unsafe"

	"github.com/gokrazy/gokapi"
	"github.com/gokrazy/gokapi/ondeviceapi"
	"github.com/gokrazy/gokrazy"

	"golang.org/x/crypto/ssh"
	execute "github.com/alexellis/go-execute/v2"

)

var ip = "10.1.1.1"
var id = ""
var login = "ssh"

var (
	authorizedKeysPath = flag.String("authorized_keys",
		"/perm/breakglass.authorized_keys",
		"path to an OpenSSH authorized_keys file; if the value is 'ec2', fetch the SSH key(s) from the AWS IMDSv2 metadata")

	hostKeyPath = flag.String("host_key",
		"/perm/breakglass.host_key",
		"path to a PEM-encoded RSA, DSA or ECDSA private key (create using e.g. ssh-keygen -f /perm/breakglass.host_key -N '' -t rsa)")

	port = flag.String("port",
		"22",
		"port for breakglass to listen on")

	enableBanner = flag.Bool("enable_banner",
		true,
		"Adds a banner to greet the user on login")

	forwarding = flag.String("forward",
		"",
		"allow port forwarding. Use `loopback` for loopback interfaces and `private-network` for private networks")
)

func loadAuthorizedKeys(path string) (map[string]bool, error) {
	var b []byte
	var err error
	switch path {
	default:
		b, err = ioutil.ReadFile(path)
	}
	if err != nil {
		return nil, err
	}

	result := make(map[string]bool)

	s := bufio.NewScanner(bytes.NewReader(b))
	for lineNum := 1; s.Scan(); lineNum++ {
		if tr := strings.TrimSpace(s.Text()); tr == "" || strings.HasPrefix(tr, "#") {
			continue
		}
		pubKey, _, _, _, err := ssh.ParseAuthorizedKey(s.Bytes())
		if err != nil {
			return nil, err
		}
		result[string(pubKey.Marshal())] = true
	}
	if err := s.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func loadHostKey(path string) (ssh.Signer, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return ssh.ParsePrivateKey(b)
}

func createHostKey(path string) (ssh.Signer, error) {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return nil, err
	}

	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0400)
	if err == nil {
		defer file.Close()

		var pkcs8 []byte
		if pkcs8, err = x509.MarshalPKCS8PrivateKey(key); err == nil {
			err = pem.Encode(file, &pem.Block{
				Type:  "PRIVATE KEY",
				Bytes: pkcs8,
			})
		}
	}
	if err != nil {
		log.Printf("could not save generated host key: %v", err)
	}

	return ssh.NewSignerFromKey(key)
}

func buildTimestamp() (string, error) {
	cfg, err := gokapi.ConnectOnDevice()
	if err != nil {
		return "", err
	}
	cl := ondeviceapi.NewAPIClient(cfg)
	res, _, err := cl.SuperviseApi.Index(context.Background())
	if err != nil {
		return "", err
	}
	return res.BuildTimestamp, nil
}

var motd string

func initMOTD() error {
	if !*enableBanner {
		return nil
	}

	hostname, err := os.Hostname()
	if err != nil {
		log.Printf("os.Hostname(): %v", err)
		hostname = "gokrazy"
	}
	const maxSpace = "                 "
	if len(hostname) > len(maxSpace) {
		hostname = hostname[:len(maxSpace)]
	}
	hostname += `"`
	if padding := len(maxSpace) - len(hostname); padding > 0 {
		hostname += strings.Repeat(" ", padding)
	}

	buildTimestamp, err := buildTimestamp()
	if err != nil {
		return err
	}

	motd = fmt.Sprintf(`              __                           
 .-----.-----|  |--.----.---.-.-----.--.--.
 |  _  |  _  |    <|   _|  _  |-- __|  |  |
 |___  |_____|__|__|__| |___._|_____|___  |
 |_____|  host:  "%s |_____|
          model: %s
          build: %s
`,
		hostname,
		gokrazy.Model(),
		buildTimestamp)
	return nil
}

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

func setWinsize(f *os.File, w, h int) {
	syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), uintptr(syscall.TIOCSWINSZ),
		uintptr(unsafe.Pointer(&struct{ h, w, x, y uint16 }{uint16(h), uint16(w), 0, 0})))
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
		if login == "ssh" {
			if _, err := os.Stat("/user/breakglass"); errors.Is(err, os.ErrNotExist) {
				log.Println("Cannot enable SSH: breakglass not found")
			} else {
				log.Println("Running SSH...")
                // unable to use breakglass directly, by design:
                //  https://github.com/gokrazy/breakglass/blob/02513c1dabef87398006421b82e48be9cf712382/README.md?plain=1#L12-L14
                
                ssh := exec.Command("ssh", "192.168.31.192")
	            //if args := flag.Args()[1:]; len(args) > 0 {
		        //    ssh.Args = append(ssh.Args, args...)
	            //}
	            //log.Printf("%v", ssh.Args)
	            ssh.Stdin = os.Stdin
	            ssh.Stdout = os.Stdout
	            ssh.Stderr = os.Stderr
	            if err := ssh.Run(); err != nil {
		            //return fmt.Errorf("%v: %v", ssh.Args, err)
	            }
                
			}
		}
	} else {
		log.Println("No id provided. Exiting...")
	}
}

