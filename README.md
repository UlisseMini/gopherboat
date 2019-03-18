# gopherboat
> Don't ask

## What the hell is this
you asked?!?!?, actally this is a system for spawninng "boats" that are
actually just OS processes, you can communicate with the boat with a simple send/recv api,
you can send a signal to a boat and the boat can send a message back to you.

all messages are blocking.

## Example
```go
package main

import (
	"fmt"
	"log"
	"os"
	"time"

	gb "github.com/UlisseMini/gopherboat"
)

// Number of pings to send to the boat.
const N = 3

// Boats must be initalized and passed to init for this to work.
var boats = []gb.Boat{
	{
		Name: "PingBoat",
		Start: func(b *gb.BoatAPI) {
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

func main() {
	gb.Init(boats)
	log.Print("launching boat")

	b, err := gb.Start("PingBoat")
	if err != nil {
		log.Fatal(err)
	}

	// Send N messages to the boat, they will be returned
	// after.
	for i := 0; i < N; i++ {
		log.Printf("sending ping #%d", i)
		err := b.Send(fmt.Sprintf("ping #%d", i))
		if err != nil {
			log.Fatal(err)
		}
	}

	// Recv the messages back from the boat.
	for i := 0; i < N; i++ {
		log.Printf("recv ping #%d", i)
		msg := new(string)
		err := b.Recv(msg)
		if err != nil {
			log.Fatal(err)
		}

		// check that the message is expected.
		want := fmt.Sprintf("got '%s'", fmt.Sprintf("ping #%d", i))
		if *msg != want {
			log.Fatalf("got %q, want %q", *msg, want)
		}
	}

	log.Print("tipping the boat")
	if err := b.Tip(); err != nil {
		log.Fatal(err)
	}
}
```
