package config

import (
	"os"

	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
)

var SnapClient snap.Client

func InitMidtrans() {
    serverKey := os.Getenv("MIDTRANS_SERVER_KEY")
    env := midtrans.Sandbox // Ganti ke midtrans.Production jika sudah live
    SnapClient.New(serverKey, env)
}