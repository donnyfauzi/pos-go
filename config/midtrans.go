package config

import (
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
	"os"
)

var MidtransClient snap.Client

func InitMidtrans() {
	serverKey := os.Getenv("server_key_mid")

	MidtransClient.New(serverKey, midtrans.Sandbox)
	// Untuk production: midtrans.Production
}
