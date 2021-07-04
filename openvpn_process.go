package openvpn

import (
	"bufio"
	"fmt"
	"github.com/golang/glog"
	"github.com/mungaij83/go-openvpn/utils"
	"os/exec"
	"sync"

)

type Process struct {
	StdOut     chan string `json:"-"`
	StdErr     chan string `json:"-"`
	Stopped    chan bool   `json:"-"`
	parameters []string
	socket     string
	config     *Config
	Env        map[string]string
	Clients    map[string]*utils.Client
	shutdown   chan bool
	waitGroup  sync.WaitGroup
}

func NewProcess(socket string, config*Config) *Process {
	p := &Process{
		Env:      make(map[string]string, 0),
		Clients:  make(map[string]*utils.Client, 0),
		socket:   socket,
		config: config,
		shutdown: make(chan bool),
	}
	return p
}

func (p *Process) Start() (err error) {
	// Check if the process is already running
	if p.Stopped != nil {
		select {
		case <-p.Stopped:
			// Everything is good, no process running
		default:
			return fmt.Errorf("openvpn is already started, aborting")
		}
	}
	// Add the management interface path to the config
	return p.Restart()
}

func (p *Process) Stop() (err error) {
	close(p.shutdown)
	p.waitGroup.Wait()

	return
}

func (p *Process) Shutdown() error {
	return p.Stop()
}

func (p *Process) Restart() (err error) {
	// Fetch the current config
	config, err := p.config.Validate()
	if err != nil {
		return err
	}
	glog.V(1).Infof("OPENVPN: Parameters: %+v", config)
	// Create the command
	cmd := exec.Command("openvpn", config...)

	// Attatch monitors for stdout, stderr and exit
	release := make(chan bool)
	defer close(release)
	p.ProcessMonitor(cmd, release)

	// Try to start the process
	err = cmd.Start()
	if err != nil {
		glog.Errorf("RESTART FAILED: %v", err)
		return err
	}

	return
}

func (p *Process) ProcessMonitor(cmd *exec.Cmd, release chan bool) {

	p.stdoutMonitor(cmd)
	p.stderrMonitor(cmd)

	p.Stopped = make(chan bool)

	go func() {
		p.waitGroup.Add(1)
		defer p.waitGroup.Done()

		defer close(p.Stopped)

		// Watch if the process exits
		done := make(chan error)
		go func() {
			<-release // Wait for the process to start
			done <- cmd.Wait()
		}()

		// Wait for shutdown or exit
		select {
		case <-p.shutdown:
			// Kill the server
			if err := cmd.Process.Kill(); err != nil {
				return
			}
			err := <-done // allow goroutine to exit
			glog.Errorf("process killed with error = %v", err)
		case err := <-done:
			glog.Errorf("process done with error = %v", err)
			return
		}

	}()
}

func (p *Process) stdoutMonitor(cmd *exec.Cmd) {
	stdout, _ := cmd.StdoutPipe()
	go func() {
		p.waitGroup.Add(1)
		defer p.waitGroup.Done()

		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			select {
			case p.StdOut <- scanner.Text():
			default:
				glog.V(2).Infof("OPENVPN stdout: %v", scanner.Text())
			}

		}
		if err := scanner.Err(); err != nil {
			glog.Warningf("OPENVPN stdout: (failed to read: %v)", err)
			return
		}
	}()
}

func (p *Process) stderrMonitor(cmd *exec.Cmd) {
	stderr, _ := cmd.StderrPipe()
	go func() {
		p.waitGroup.Add(1)
		defer p.waitGroup.Done()

		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			select {
			case p.StdErr <- scanner.Text():
			default:
				glog.Warningf("OPENVPN stderr: %v", scanner.Text())
			}
		}
		if err := scanner.Err(); err != nil {
			glog.Warningf("OPENVPN stderr: (failed to read %v)", err)
			return
		}
	}()
}
