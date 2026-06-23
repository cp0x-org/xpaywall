// Command mpp-client exercises an xgateway MPP-protected endpoint end to end.
//
// It first issues an unpaid request to surface the HTTP 402 Payment Required
// challenge, then pays it with a Tempo "charge" credential and prints the
// settled response. The result proves a paid MPP call routed through xgateway.
//
// Configuration (from a .env file in the working directory or the environment):
//
//	RPC_URL      Tempo JSON-RPC endpoint (e.g. https://rpc.moderato.tempo.xyz)
//	PRIVATE_KEY  payer wallet private key (funded on the same Tempo chain)
//
// Usage:
//
//	go run . [target-url]
//
// target-url defaults to http://localhost:8081/http-endpoint when omitted.
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/tempoxyz/mpp-go/pkg/client"
	"github.com/tempoxyz/mpp-go/pkg/mpp"
	charge "github.com/tempoxyz/mpp-go/pkg/tempo/client"
)

const defaultTarget = "http://localhost:8081/http-endpoint"

func main() {
	// .env is optional — environment variables take precedence when both exist.
	_ = godotenv.Load()

	rpcURL := firstNonEmpty(os.Getenv("RPC_URL"), os.Getenv("MPP_RPC_URL"))
	privateKey := firstNonEmpty(os.Getenv("PRIVATE_KEY"), os.Getenv("MPP_PRIVATE_KEY"))
	if rpcURL == "" {
		log.Fatal("RPC_URL is required (set it in .env or the environment)")
	}
	if privateKey == "" {
		log.Fatal("PRIVATE_KEY is required (set it in .env or the environment)")
	}

	target := defaultTarget
	if len(os.Args) > 1 && strings.TrimSpace(os.Args[1]) != "" {
		target = strings.TrimSpace(os.Args[1])
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	fmt.Printf("target : %s\n", target)
	fmt.Printf("rpc    : %s\n", rpcURL)

	// --- Step 1: unpaid request — expect an MPP 402 challenge ----------------
	fmt.Println("\n[1] sending unpaid request ...")
	unpaid, err := http.Get(target) //nolint:gosec // target is operator-supplied
	if err != nil {
		log.Fatalf("unpaid request failed: %v", err)
	}
	unpaidBody, _ := io.ReadAll(unpaid.Body)
	_ = unpaid.Body.Close()

	fmt.Printf("    status           : %s\n", unpaid.Status)
	fmt.Printf("    WWW-Authenticate : %s\n", unpaid.Header.Get("WWW-Authenticate"))
	fmt.Printf("    body             : %s\n", strings.TrimSpace(string(unpaidBody)))
	if unpaid.StatusCode != http.StatusPaymentRequired {
		log.Fatalf("expected 402 Payment Required, got %s", unpaid.Status)
	}

	// --- Step 2: paid request — MPP client settles the charge and retries ----
	fmt.Println("\n[2] paying via MPP (Tempo charge) and retrying ...")
	method, err := charge.New(charge.Config{PrivateKey: privateKey, RPCURL: rpcURL})
	if err != nil {
		log.Fatalf("build tempo charge method: %v", err)
	}
	mppClient := client.New([]client.Method{method})

	paid, err := mppClient.Get(ctx, target)
	if err != nil {
		log.Fatalf("paid request failed: %v", err)
	}
	paidBody, _ := io.ReadAll(paid.Body)
	_ = paid.Body.Close()

	fmt.Printf("    status : %s\n", paid.Status)
	if header := paid.Header.Get("Payment-Receipt"); header != "" {
		if receipt, perr := mpp.ParseReceipt(header); perr == nil {
			fmt.Printf("    receipt: status=%s method=%s reference=%s\n",
				receipt.Status, receipt.Method, receipt.Reference)
		}
	}
	fmt.Printf("    body   : %s\n", strings.TrimSpace(string(paidBody)))

	if paid.StatusCode != http.StatusOK {
		log.Fatalf("\nFAILED: paid request returned %s (is the upstream behind xgateway running?)", paid.Status)
	}

	fmt.Println("\nSUCCESS: paid MPP call through xgateway completed.")
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
