/* NameCheapDynDNS
 *
 * MIT License Copyright 2014 <Kristopher Watts>
*/

package main

import (
	"errors"
	"flag"
	"fmt"
	"time"
	"os"
	"io/ioutil"
	"strings"
	"net/http"

	"code.google.com/p/gcfg"
)
type ipConf struct {
	Host string
	Domain string
	Key string
}

type cfgType struct {
	Global struct {
		UpdateInterval int
	}
	DynDomain map[string]*ipConf
}

const (
	whatismyipurl string = "http://bot.whatismyipaddress.com"
	baseURL string = "http://dynamicdns.park-your-domain.com/update?"
)

var (
	confFile = flag.String("c", "/opt/NameCheapDynDNS/settings.conf", "Path to configuration file")
)

func init() {
	flag.Parse()
	if *confFile == "" {
		fmt.Printf("No settings file specified")
		os.Exit(-1)
	}

}

func main() {
	currIP := ""

	//get config file content
	confContent, err := ioutil.ReadFile(*confFile)
	if err != nil {
		fmt.Printf("Failed to load config file: %v\n", err)
		os.Exit(-1)
	}

	//load the configuration into our conf struct
	conf, err := loadConfigs(string(confContent))
	if err != nil {
		fmt.Printf("Invalid configuration file: %v\n", err)
		os.Exit(-1)
	}

	for {
		localIP, err := getIP()
		if err == nil {
			if currIP != localIP {
				err = UpdateDynIPs(localIP, conf)
				if err != nil {
					fmt.Printf("Failed to update host: %v\n", err)
				}
			}
		} else {
			fmt.Printf("Failed to hit google and get remote IP: %v", err)
		}
		time.Sleep(time.Minute*time.Duration(conf.Global.UpdateInterval))
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
	if err != nil {
		return "", err
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
			fmt.Printf("Failed to update %v: %v\n", key, err)
		}
	}
	return totalError
}

func UpdateIPs(address string, cfg *ipConf) error {
	url := fmt.Sprintf("%s?ip=%s&host=%s&domain=%s&password=%s",
		baseURL, address, cfg.Host, cfg.Domain, cfg.Key)
	body, err := hitURL(url)
	if err != nil {
		return errors.New("Failed to issue update request")
	}
	if !strings.Contains(string(body), "<ErrCount>0</ErrCount>") {
		return fmt.Errorf("Failed to update %v.%v", cfg.Host, cfg.Domain)
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
	if err != nil {
		return nil, err
	}
	return body[0:n], nil
}

func loadConfigs(cnt string) (cfgType, error) {
	var cfg cfgType
	err := gcfg.ReadStringInto(&cfg, cnt)
	return cfg, err
}
