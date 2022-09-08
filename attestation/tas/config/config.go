// Description: Store TAS configurations

package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"log"

	"github.com/spf13/viper"
)

const (
	// config path
	strLocalConf = "."
	// config file
	confName = "config"
	confExt  = "yaml"

	confServPort       = "tasconfig.port"
	confRestPort       = "tasconfig.rest"
	confTASKeyCertFile = "tasconfig.akskeycertfile"
	confTASPrivKeyFile = "tasconfig.aksprivkeyfile"
	confHWITCACertFile = "tasconfig.huaweiitcafile"
)

type (
	tasConfig struct {
		servPort       string
		restPort       string
		tasKeyCertFile string
		tasPrivKeyFile string
		hwItCACertFile string
	}
)

var (
	defaultPaths = []string{
		strLocalConf,
	}
	tasCfg    *tasConfig
	ascert    *x509.Certificate
	asprivkey *rsa.PrivateKey
	hwcert    *x509.Certificate
)

func LoadConfigs() {
	log.Print("Load TAS configs...")
	if tasCfg != nil {
		return
	}
	tasCfg = &tasConfig{}
	viper.SetConfigName(confName)
	viper.SetConfigType(confExt)
	for _, s := range defaultPaths {
		viper.AddConfigPath(s)
	}
	err := viper.ReadInConfig()
	if err != nil {
		return
	}

	tasCfg.servPort = viper.GetString(confServPort)
	tasCfg.restPort = viper.GetString(confRestPort)
	tasCfg.tasKeyCertFile = viper.GetString(confTASKeyCertFile)
	tasCfg.tasPrivKeyFile = viper.GetString(confTASPrivKeyFile)
	tasCfg.hwItCACertFile = viper.GetString(confHWITCACertFile)
}

func InitializeAS() error {
	// parse as cert
	certfile := GetASCertFile()
	ascertbyte, err := ioutil.ReadFile(certfile)
	if err != nil {
		return err
	}
	ascertBlock, _ := pem.Decode(ascertbyte)
	ascert, err = x509.ParseCertificate(ascertBlock.Bytes)
	if err != nil {
		return err
	}
	// parse as private key
	askeyfile := GetASKeyFile()
	askeybyte, err := ioutil.ReadFile(askeyfile)
	if err != nil {
		return err
	}
	askeyBlock, _ := pem.Decode(askeybyte)
	asprivkey, err = x509.ParsePKCS1PrivateKey(askeyBlock.Bytes)
	if err != nil {
		return err
	}
	// parse hw it ca cert
	hwcafile := GetHWCertFile()
	hwcabyte, err := ioutil.ReadFile(hwcafile)
	if err != nil {
		return err
	}
	hwcaBlock, _ := pem.Decode(hwcabyte)
	hwcert, err = x509.ParseCertificate(hwcaBlock.Bytes)
	if err != nil {
		return err
	}
	return nil
}

func GetConfigs() *tasConfig {
	return tasCfg
}

func GetServerPort() string {
	if tasCfg == nil {
		return ""
	}
	return tasCfg.servPort
}

func GetASCertFile() string {
	if tasCfg == nil {
		return ""
	}
	return tasCfg.tasKeyCertFile
}

func GetASKeyFile() string {
	if tasCfg == nil {
		return ""
	}
	return tasCfg.tasPrivKeyFile
}

func GetHWCertFile() string {
	if tasCfg == nil {
		return ""
	}
	return tasCfg.hwItCACertFile
}

func GetASCert() *x509.Certificate {
	return ascert
}

func GetASPrivKey() *rsa.PrivateKey {
	return asprivkey
}

func GetHWCert() *x509.Certificate {
	return hwcert
}
