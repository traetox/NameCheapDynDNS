/* NameCheapDynDNS
 *
 * MIT License Copyright 2014 <Kristopher Watts>
 */

package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	gcfg "gopkg.in/gcfg.v1"
)

type ipConf struct {
	Host   string
	Domain string
	Key    string
}

type cfgType struct {
	Global struct {
		UpdateInterval int
	}
	DynDomain map[string]*ipConf
}

const (
	whatismyipurl string = "https://ipinfo.io/ip"
	baseURL       string = "http://dynamicdns.park-your-domain.com/update?"
)

var (
	confFile = flag.String("c", "/opt/NameCheapDynDNS/settings.conf", "Path to configuration file")
)

func init() {
	flag.Parse()
	if *confFile == "" {
		log.Fatal("No settings file specified")
	}

}

func main() {
	currIP := ""

	//get config file content
	confContent, err := ioutil.ReadFile(*confFile)
	if err != nil {
		log.Fatalf("Failed to load config file: %v\n", err)
	}

	//load the configuration into our conf struct
	conf, err := loadConfigs(string(confContent))
	if err != nil {
		log.Fatalf("Invalid configuration file: %v\n", err)
	}

	for {
		localIP, err := getIP()
		if err == nil {
			if currIP != localIP {
				err = UpdateDynIPs(localIP, conf)
				if err != nil {
					log.Printf("Failed to update host: %v\n", err)
				}
			}
		} else {
			log.Printf("Failed to hit google and get remote IP: %v", err)
		}
		time.Sleep(time.Minute * time.Duration(conf.Global.UpdateInterval))
	}
}

func getIP() (string, error) {
	resp, err := http.Get(whatismyipurl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body := make([]byte, 4096)
	n, err := resp.Body.Read(body)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("Failed to read body %s: %v %d", resp.Status, err, n)
	}
	if net.ParseIP(string(body[0:n])) == nil {
		return "", fmt.Errorf("Bad IP response: \"%v\"", string(body[0:n]))
	}
	ipaddy := string(body[0:n])
	return ipaddy, nil
}

func UpdateDynIPs(address string, cfg cfgType) error {
	var totalError error
	for key, val := range cfg.DynDomain {
		err := UpdateIPs(address, val)
		if err != nil {
			totalError = err
			log.Printf("Failed to update %v: %v\n", key, err)
		}
	}
	return totalError
}

func UpdateIPs(address string, cfg *ipConf) error {
	url := fmt.Sprintf("%s?ip=%s&host=%s&domain=%s&password=%s",
		baseURL, address, cfg.Host, cfg.Domain, cfg.Key)
	body, err := hitURL(url)
	if err != nil {
		return fmt.Errorf("Failed to update IP: %v", err)
	}
	if !strings.Contains(string(body), "<ErrCount>0</ErrCount>") {
		return fmt.Errorf("Failed to update %v.%v: %v", cfg.Host, cfg.Domain, string(body))
	}
	return nil
}

func hitURL(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body := make([]byte, 4096)
	n, err := resp.Body.Read(body)
	if err != nil && err != io.EOF {
		return nil, err
	}
	return body[0:n], nil
}

func loadConfigs(cnt string) (cfgType, error) {
	var cfg cfgType
	err := gcfg.ReadStringInto(&cfg, cnt)
	return cfg, err
}
