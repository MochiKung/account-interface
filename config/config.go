package config

import (
	"encoding/xml"
	"log"
)

const (
	DefaultFile = "conf/main.xml"
)

var (
	Default *Root
)

func init() {
	var err error

	// load default configurations
	Default, err = Load(DefaultFile)
	if err != nil {
		log.Fatalln(err)
	}
}

type Root struct {
	XMLName  xml.Name `xml:"itemcode-db"`
	Server   *Server  `xml:"server"`
	Database Database `xml:"database"`
}

type Server struct {
	Address string `xml:"address"`
	Tls     *Tls   `xml:"tls"`
}

type Tls struct {
	Enable             bool   `xml:"enable,attr"`
	CertificateFile    string `xml:"certificate-file"`
	CertificateKeyFile string `xml:"certificate-key-file"`
}

type Database struct {
	BoltDB BoltDB `xml:"bolt-db"`
}

type BoltDB struct {
	ClientDB       string `xml:"client-db"`
	UserDB         string `xml:"user-db"`
	AccessTokenDB  string `xml:"access-token-db"`
	RefreshTokenDB string `xml:"refresh-token-db"`
}
