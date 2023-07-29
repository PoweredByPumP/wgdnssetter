package wgdnssetter

import (
	"fmt"
	"github.com/joshbetz/config"
	"github.com/xiaost/jsonport"
	"io/ioutil"
	"os"
	"strings"
)

var appConfig *config.Config
var validMailTld string
var dbDir string
var dnsTld string
var registeredClients map[string]string

func init() {
	appConfig = config.New("settings.json")

	err := appConfig.Get("valid_mail_tld", &validMailTld)
	if err != nil || validMailTld == "" {
		fmt.Println("Setting valid_mail_tld is not set or file unreadable.")
		os.Exit(1)
	}

	err = appConfig.Get("db_dir", &dbDir)
	if err != nil || dbDir == "" {
		fmt.Println("Setting db_dir is not set or file unreadable.")
		os.Exit(1)
	}

	err = appConfig.Get("dns_tld", &dnsTld)
	if err != nil || dnsTld == "" {
		fmt.Println("Setting dns_tld is not set or file unreadable.")
		os.Exit(1)
	}

	registeredClients = make(map[string]string)
}

func SetDnsClientEntry() bool {
	clients, _ := ioutil.ReadDir(dbDir)

OUTER:
	for _, client := range clients {
		if !client.IsDir() {
			filePath := dbDir + "/" + client.Name()
			fmt.Println("Reading " + filePath)

			data, err := os.ReadFile(filePath)
			if err != nil {
				fmt.Println("Can't read file located in " + filePath)
				continue
			}

			c, err := jsonport.Unmarshal(data)
			if err != nil {
				fmt.Println("File " + filePath + " is not a JSON file")
				continue
			}

			email, err := c.GetString("email")
			if err != nil {
				fmt.Println("Can't read file located in " + filePath)
				fmt.Println()
				continue
			}

			if email == "" ||
				!strings.Contains(email, "@"+validMailTld) {
				fmt.Println("Client email doesn't meet requirements")
				continue
			}

			fmt.Println(email)

			ips, err := c.Get("allocated_ips").StringArray()
			if err != nil {
				fmt.Println("Can't read file located in " + filePath)
				continue
			}

			for _, ip := range ips {
				fmt.Println(ip)
				ipArray := strings.Split(ip, "/")

				if ipArray[1] != "32" {
					fmt.Println("Ip " + ip + " for client " + email + " doesn't meet requirements")
					continue
				}

				clientEmailArray := strings.Split(email, "@")
				if clientEmailArray[0] != "" &&
					registeredClients[clientEmailArray[0]] != "" {
					fmt.Println("OUTER")
					continue OUTER
				}

				registeredClients[clientEmailArray[0]] = ipArray[0]
			}
		}
	}

	generateDnsFile()

	return true
}

func generateDnsFile() {
	var fileContent string

	if len(registeredClients) != 0 {
		for clientName, clientIp := range registeredClients {
			fileContent += clientIp + " " + clientName + "." + dnsTld + "\n"
		}
	}

	fmt.Println(fileContent)

	// TODO: Write file
}
