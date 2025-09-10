package config

import (
	"os"

	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
)

var SnapClient snap.Client

func InitMidtrans() {
	midtrans.ServerKey = os.Getenv("MIDTRANS_SERVER_KEY")
	midtrans.ClientKey = os.Getenv("MIDTRANS_CLIENT_KEY")
	
	environment := os.Getenv("MIDTRANS_ENVIRONMENT")
	if environment == "production" {
		midtrans.Environment = midtrans.Production
	} else {
		midtrans.Environment = midtrans.Sandbox
	}

	SnapClient = snap.Client{}
	SnapClient.New(midtrans.ServerKey, midtrans.Environment)
}