package handler

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"testing"
	"time"

	"github.com/ChainSafe/gossamer/lib/common"
	"github.com/ChainSafe/gossamer/lib/crypto/sr25519"
	"github.com/gorilla/websocket"

	"github.com/stretchr/testify/require"
)

func TestVerifySign(t *testing.T) {
	err := VerifySign("", "", "")
	if err != nil {
		panic(err)
	}

	fmt.Println("verified signature")
}

func TestPubkeyToAddress(t *testing.T) {
	// randomly generated from subkey
	pub, _ := common.HexToBytes("")
	addr := ""

	pk, err := sr25519.NewPublicKey(pub)
	require.NoError(t, err)
	a := pk.Address()
	require.Equal(t, addr, string(a))
}

func TestWebsocket(t *testing.T) {

	addr := "ws://127.0.0.1:8000/ws/getevents"

	u, err := url.Parse(addr)
	if err != nil {
		log.Fatal(err)
	}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("read msg error:", err)
				return
			}
			fmt.Printf("receive msg: %s\n", message)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	<-c

	err = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Println("close coon:", err)
		return
	}
	select {
	case <-done:
	case <-time.After(time.Second):
	}

}
