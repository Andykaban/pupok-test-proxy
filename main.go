package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/armon/go-socks5"
)

var proxyConfig ProxyConfig
var statData *ConnStat

//ProxyConfig structure for main configuration
type ProxyConfig struct {
	Host            string `json:"host"`
	Port            string `json:"port"`
	Login           string `json:"login"`
	Password        string `json:"password"`
	DeadlineTimeOut int    `json:"deadline_timeout,omitempty"`
}

//ConnStat structure for aggregate clients connections
type ConnStat struct {
	mutex *sync.RWMutex
	stat  map[string]bool
}

//UpdateConnStat function update client connection info
func (c *ConnStat) UpdateConnStat(addr string, status bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.stat[addr] = status
}

//GetActiveConnections function return active connections slice
func (c *ConnStat) GetActiveConnections() []string {
	activeConnections := make([]string, 0)
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	//for debugging
	//log.Println(c.stat)
	for ipAddr, status := range c.stat {
		if status == true {
			activeConnections = append(activeConnections, ipAddr)
		}
	}
	return activeConnections
}

//DeleteClosedConnections funtion wich delete closed connections data
func (c *ConnStat) DeleteClosedConnections() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for ipAddr, status := range c.stat {
		if status == false {
			delete(c.stat, ipAddr)
		}
	}
}

//NewConnStat function return ConnStat struct
func NewConnStat() *ConnStat {
	return &ConnStat{mutex: &sync.RWMutex{}, stat: make(map[string]bool)}
}

//HTTPConnStatHandler main http handler
func HTTPConnStatHandler(w http.ResponseWriter, r *http.Request) {
	activeConnections := statData.GetActiveConnections()
	activeConnectionsStr := strings.Join(activeConnections, "\n")
	fmt.Fprintf(w, activeConnectionsStr)
}

//HandleSocks5Connect SOCK5 client connection handler
func HandleSocks5Connect(server *socks5.Server, connection net.Conn, deadlineInSec int) {
	ipAddr := connection.RemoteAddr().String()
	statData.UpdateConnStat(ipAddr, true)
	if deadlineInSec != 0 {
		deadLineDuration := time.Duration(deadlineInSec) * time.Second
		connection.SetDeadline(time.Now().Add(deadLineDuration))
	}
	err := server.ServeConn(connection)
	if err != nil {
		log.Println("---------------------------------")
		log.Println(err)
		log.Println("---------------------------------")
	}
	statData.UpdateConnStat(ipAddr, false)
}

//RunHTTPServer function which setup and run HTTP server
func RunHTTPServer() {
	http.HandleFunc("/", HTTPConnStatHandler)
	log.Println("Start web server on 9000 port")
	http.ListenAndServe(":9000", nil)
}

//ClosedConnectionsRemover function wich remove closed conections by timer
func ClosedConnectionsRemover() {
	for {
		statData.DeleteClosedConnections()
		time.Sleep(10 * time.Second)
	}
}

//Usage print usage function
func Usage() {
	log.Println("!!! Pupok proxy server !!!")
	log.Println("Usage: pupok-test-proxy -config 'path to config proxy file'")
	flag.PrintDefaults()
}

func init() {
	config := flag.String("config", "config.json", "Path to configuration file")
	flag.Usage = Usage
	flag.Parse()

	raw, err := os.Open(*config)
	defer raw.Close()
	if err != nil {
		log.Fatal(err)
	}
	byteValue, _ := ioutil.ReadAll(raw)
	json.Unmarshal(byteValue, &proxyConfig)

	statData = NewConnStat()
}

func main() {
	credentials := socks5.StaticCredentials{
		proxyConfig.Login: proxyConfig.Password,
	}
	deadlineInSeconds := proxyConfig.DeadlineTimeOut

	conf := &socks5.Config{
		Logger: log.New(os.Stdout, "[pupok-test-proxy]", log.Ldate|log.Ltime|log.Lshortfile),
		AuthMethods: []socks5.Authenticator{
			socks5.UserPassAuthenticator{Credentials: credentials},
		},
	}

	socks5Server, err := socks5.New(conf)
	go RunHTTPServer()
	go ClosedConnectionsRemover()
	if err != nil {
		log.Fatal(err)
	}
	listenAddr := fmt.Sprintf("%s:%s", proxyConfig.Host, proxyConfig.Port)
	log.Printf("Listen %s...", listenAddr)
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatal(err)
	}
	for {
		connection, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go HandleSocks5Connect(socks5Server, connection, deadlineInSeconds)
	}
}
