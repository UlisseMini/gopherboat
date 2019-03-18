// Note: This is one big integration test since everything is so intertwined.
// Things that still need testing include,
// - Sending signals
// - Waiting for the process to exit
// Bad things with the integration tests
// - If it fails i will not get information from the boat.
//   maybe have a goroutine record stderr from the boat and use it for failure messages.

package gopherboat

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"
)

var boats = []Boat{
	{
		Name: "Integration",
		Start: func(b *BoatAPI) {
			var messages []string
			// Recv the 3 echo pings
			for i := 0; i < N; i++ {
				buf := new(string)
				if err := b.Recv(buf); err != nil {
					b.Send(fmt.Sprintf("error: %#v", err))
					log.Fatal(err)
				}

				messages = append(messages, *buf)
			}

			// Send them back
			for i := 0; i < N; i++ {
				msg := fmt.Sprintf("got '%s'", messages[i])
				if err := b.Send(msg); err != nil {
					log.Fatal(err)
				}
			}

			// Wait for a long time, the supervisor should kill us.
			time.Sleep(time.Minute)

			// Make sure we fail if the supervisor did not kill us.
			os.Exit(1)
		},
	},
}

// Number of pings to send to the integration boat.
const N = 3

// Main integration test, tests a whole bunch of stuff at once
// since its so interconnected.
func TestBoat(t *testing.T) {
	Init(boats)
	t.Log("launching boat")

	b, err := Start("Integration")
	if err != nil {
		t.Fatal(err)
	}

	// Send N messages to the boat, they will be returned
	// after.
	for i := 0; i < N; i++ {
		t.Logf("sending ping #%d", i)
		err := b.Send(fmt.Sprintf("ping #%d", i))
		if err != nil {
			t.Fatal(err)
		}
	}

	// Recv the messages back from the boat.
	for i := 0; i < N; i++ {
		t.Logf("recv ping #%d", i)
		msg := new(string)
		err := b.Recv(msg)
		if err != nil {
			t.Fatal(err)
		}

		// check that the message is expected.
		want := fmt.Sprintf("got '%s'", fmt.Sprintf("ping #%d", i))
		if *msg != want {
			t.Fatalf("got %q, want %q", *msg, want)
		}
	}

	t.Log("tipping the boat")
	if err := b.Tip(); err != nil {
		t.Fatal(err)
	}
}
