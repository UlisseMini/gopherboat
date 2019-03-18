// Package gopherboat provides a gopher and a boat, duh
package gopherboat

import (
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
	"time"
)

// BoatAPI is the api given to started boats.
type BoatAPI struct {
	stdin  *gob.Decoder
	stdout *gob.Encoder
}

// NewAPI Initializes a new boat api
func NewAPI() *BoatAPI {
	return &BoatAPI{
		stdin:  gob.NewDecoder(os.Stdin),
		stdout: gob.NewEncoder(os.Stdout),
	}
}

// Recv reads the next value from stdin and stores it in the data
// represented by the empty interface value. If e is nil, the value will be
// discarded. Otherwise, the value underlying e must be a pointer to the
// correct type for the next data item received. If the input is at EOF, Decode
// returns io.EOF and does not modify e.
func (b BoatAPI) Recv(e interface{}) error {
	return b.stdin.Decode(e)
}

// Send sends the data item represented by the empty interface value to the supervisor,
// guaranteeing that all necessary type information has been transmitted first.
// Passing a nil pointer to Encoder will panic, as they cannot be transmitted
// by gob.
func (b BoatAPI) Send(e interface{}) error {
	return b.stdout.Encode(e)
}

// the envoirment variable that is present when spawned inside a boat
const envKey = "GOPHERBOAT_NAME"

// Boat represents a boat that can be started.
type Boat struct {
	Name  string
	Start func(api *BoatAPI)
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

		stdinGob:  gob.NewEncoder(stdin),
		stdoutGob: gob.NewDecoder(stdout),
	}

	return handle, nil
}

// BoatHandle is a handle on a running boat
type BoatHandle struct {
	// the running process
	cmd *exec.Cmd

	// the processes file descriptors, not used
	// but handy to have.
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser

	// gob handles on stdin and stdout
	stdinGob  *gob.Encoder
	stdoutGob *gob.Decoder
}

// Send sends the data item represented by the empty interface value to the boat,
// guaranteeing that all necessary type information has been transmitted first.
// Passing a nil pointer to Encoder will panic, as they cannot be transmitted
// by gob.
func (b *BoatHandle) Send(m interface{}) error {
	return b.stdinGob.Encode(m)
}

// Recv reads the next value from stdin and stores it in the data
// represented by the empty interface value. If e is nil, the value will be
// discarded. Otherwise, the value underlying e must be a pointer to the
// correct type for the next data item received. If the input is at EOF, Decode
// returns io.EOF and does not modify e.
func (b *BoatHandle) Recv(e interface{}) error {
	return b.stdoutGob.Decode(e)
}

// Signal sends a signal to a boat.
func (b *BoatHandle) Signal(sig os.Signal) error { return b.cmd.Process.Signal(sig) }

// Wait waits for the boat process to exit.
func (b *BoatHandle) Wait() (*os.ProcessState, error) { return b.cmd.Process.Wait() }

// Tip tips over the boat drowning its gophers, a delay can
// be provided to give them time to grap something that floats before the
// boat is tipped over.
// TODO: Graceful shutdown
func (b *BoatHandle) Tip() error {
	// Give it 100 milliseconds to exit before sending SIGKILL
	if err := b.Signal(os.Interrupt); err != nil {
		return err
	}
	time.Sleep(time.Millisecond * 100)

	// Send SIGKILL
	if err := b.Signal(syscall.SIGKILL); err != nil {
		return err
	}

	// Wait for the process to exit, discard its process state
	_, err := b.Wait()
	return err
}

// The boat runtime.
func runBoat(boat Boat) {
	// We want to terminate after this function since the main function will be ran
	// otherwise.
	defer func() {
		e := recover()
		if e == nil {
			os.Exit(0)
		}
		panic(e)
	}()

	// Start in a new goroutine
	boat.Start(NewAPI())
}
