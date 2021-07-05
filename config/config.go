package config

import (
	"time"

	. "github.com/gogrlx/grlx/types"
	nats_server "github.com/nats-io/nats-server/v2/server"
)

var DefaultTestOptions nats_server.Options
var BuildInfo Version

var ConfigRoot = "./etc/grlx/"
var FarmerPKI = ConfigRoot + "pki/farmer/"
var SproutPKI = ConfigRoot + "pki/sprout/"
var CertFile = FarmerPKI + "tls-cert.pem"
var KeyFile = FarmerPKI + "tls-key.pem"
var NKeyPubFile = FarmerPKI + "farmer.nkey.pub"
var NKeyPrivFile = FarmerPKI + "farmer.nkey"

var CertificateValidTime = 365 * 24 * time.Hour

func init() {
	DefaultTestOptions = nats_server.Options{
		Host:                  "0.0.0.0",
		Port:                  4443,
		NoSigs:                true,
		MaxControlLine:        4096,
		DisableShortFirstPing: true,
		Trace:                 true,
		Debug:                 true,
	}
}
