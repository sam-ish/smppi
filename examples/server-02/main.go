package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	smpp "github.com/sam-ish/smppi"
	"github.com/sam-ish/smppi/pdu"
)

var (
	serverAddr string
	systemID   string
	msgID      int
)

func main() {
	flag.StringVar(&serverAddr, "addr", "localhost:2775", "server will listen on this address.")
	flag.StringVar(&systemID, "systemid", "ExampleServer", "descriptive server identification.")
	flag.Parse()

	sessConf := smpp.SessionConf{
		Handler: smpp.HandlerFunc(func(ctx *smpp.Context) {
			switch ctx.CommandID() {
			case pdu.BindTransceiverID:
				btrx, err := ctx.BindTRx()
				if err != nil {
					fail("Invalid PDU in context error: %+v", err)
				}
				resp := btrx.Response(systemID)
				// Validate the credentials
				if !isValidBindAccount(btrx.SystemID, btrx.Password) {
					if err := ctx.Respond(resp, pdu.StatusInvPaswd); err != nil {
						fail("Server can't respond to the Binding request: %+v", err)
					}
				}
				if err := ctx.Respond(resp, pdu.StatusOK); err != nil {
					fail("Server can't respond to the Binding request: %+v", err)
				}
			case pdu.SubmitSmID:
				sm, err := ctx.SubmitSm()
				if err != nil {
					fail("Invalid PDU in context error: %+v", err)
				}
				fmt.Fprintf(os.Stdout, "UPPER: %s\n", strings.ToUpper(sm.ShortMessage))
				msgID++
				resp := sm.Response(fmt.Sprintf("msgID_%d", msgID))
				if err := ctx.Respond(resp, pdu.StatusOK); err != nil {
					fail("Server can't respond to the submit_sm request: %+v", err)
				}
			case pdu.UnbindID:
				unb, err := ctx.Unbind()
				if err != nil {
					fail("Invalid PDU in context error: %+v", err)
				}
				resp := unb.Response()
				if err := ctx.Respond(resp, pdu.StatusOK); err != nil {
					fail("Server can't respond to the submit_sm request: %+v", err)
				}
				ctx.CloseSession()
			}
		}),
	}
	srv := smpp.NewServer(serverAddr, sessConf)

	fmt.Println("'%s' is listening on '%s'", systemID, serverAddr)
	err := srv.ListenAndServe()
	if err != nil {
		fail("Serving exited with error: %+v", err)
	}
	fmt.Println("Server closed")
}

func fail(msg string, params ...interface{}) {
	fmt.Println(msg+"\n", params)
	os.Exit(1)
}

func isValidBindAccount(systemID string, password string) bool {
	// ...
	return true
}
