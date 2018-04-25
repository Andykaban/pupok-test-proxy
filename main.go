package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/armon/go-socks5"
	"io/ioutil"
	"log"
	"os"
)

type ProxyConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

func main() {
	var proxyConfig ProxyConfig
	config := flag.String("config", "config.json", "")
	flag.Parse()

	raw, err := os.Open(*config)
	defer raw.Close()
	if err != nil {
		log.Fatal(err)
	}
	byteValue, _ := ioutil.ReadAll(raw)
	json.Unmarshal(byteValue, &proxyConfig)

	credentials := socks5.StaticCredentials{
		proxyConfig.Login: proxyConfig.Password,
	}
	conf := &socks5.Config{
		Logger: log.New(os.Stdout, "[pupok-test-proxy]", log.Ldate|log.Ltime|log.Lshortfile),
		AuthMethods: []socks5.Authenticator{
			socks5.UserPassAuthenticator{credentials},
		},
	}

	server, err := socks5.New(conf)
	if err != nil {
		log.Fatal(err)
	}
	listenAddr := fmt.Sprintf("%s:%s", proxyConfig.Host, proxyConfig.Port)
	log.Printf("Listen %s...", listenAddr)
	if err := server.ListenAndServe("tcp", listenAddr); err != nil {
		log.Fatal(err)
	}
}
