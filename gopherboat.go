// Package gopherboat provides a gopher and a boat, duh
package gopherboat

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"
)

// BoatMessage represents a boat message, they are sent on stdout and received
// from stdin, encoded in gob.
type BoatMessage struct {
	Type string      // type of the message
	Data interface{} // data with the message
}

func runBoat(boat Boat) {
	defer os.Exit(0)

	// Start in a new goroutine
	boat.Start()
}

// the envoirment variable that is present when spawned inside a boat
const envKey = "GOPHERBOAT_NAME"

type Boat struct {
	Name  string
	Start func()
}

// Init must be called with all the boats in dock at the start of main()
func Init(boats []Boat) {
	name := os.Getenv(envKey)
	if name == "" {
		return
	}

	for _, boat := range boats {
		if boat.Name != "" {
			runBoat(boat)
		}
	}

	panic(fmt.Sprintf("Boat with name %q is not in the boats array.", name))
}

// Start a boat, the boat must be in the boats passed to Init and have the same name..
func Start(boat string) (*BoatHandle, error) {
	myPath, err := os.Executable()
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(myPath)
	cmd.Env = append(os.Environ(), "GOPHERBOAT_NAME="+boat)
	// TODO: Fork process in a cross platform way

	// handle statement when?
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	handle := &BoatHandle{
		cmd:    cmd,
		stdin:  stdin,
		stdout: stdout,
		stderr: stderr,
	}

	return handle, nil
}

// BoatHandle is a handle on a running boat
type BoatHandle struct {
	// the running process
	cmd *exec.Cmd

	// the processes file descriptors
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser
}

// Tip tips over the boat drowning its gophers, a delay can
// be provided to give them time to grap something that floats before the
// boat is tipped over.
// TODO: Graceful shutdown
func (b *BoatHandle) Tip() error {
	// Give it 100 milliseconds to exit before sending SIGKILL
	if err := b.cmd.Process.Signal(os.Interrupt); err != nil {
		return err
	}
	time.Sleep(time.Millisecond * 100)

	if err := b.cmd.Process.Kill(); err != nil {
		return err
	}

	_, err := b.cmd.Process.Wait()
	return err
}
