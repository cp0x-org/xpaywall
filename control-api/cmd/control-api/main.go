package main

import (
	"context"
	"log"
	"os"

	controlapiapp "github.com/cp0x-org/xpaywall/control-api/internal/app"
)

func main() {
	if err := controlapiapp.Run(context.Background(), os.Args); err != nil {
		log.Fatalf("control-api failed: %v", err)
	}
}
