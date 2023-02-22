package main

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"time"
)

type Server struct {
	serverName string
	serverURL  string
	timeSpent  float64
	status     int
	failTime   string
}

func createListServers(serverList *os.File) []Server {
	csvReader := csv.NewReader(serverList)
	data, err := csvReader.ReadAll()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	var servers []Server
	for i, line := range data {
		if i > 0 {
			server := Server{
				serverName: line[0],
				serverURL:  line[1],
			}
			servers = append(servers, server)
		}
	}
	return servers
}

func checkServer(servers []Server) []Server {
	var downServers []Server
	now := time.Now()

	for _, server := range servers {

		dateNow := time.Now()
		get, err := http.Get(server.serverURL)
		if err != nil {
			fmt.Printf("Server %s is down [%s]\n", server.serverName, err.Error())
			server.status = 0
			server.failTime = now.Format("02/01/2006 15:04:05")
			downServers = append(downServers, server)
			continue

		}
		server.status = get.StatusCode
		if server.status != 200 {
			server.failTime = now.Format("02/01/2006 15:04:05")
			downServers = append(downServers, server)
		}
		server.timeSpent = time.Since(dateNow).Seconds()
		fmt.Printf("Status: [%d] Tempo de carga: [%f] URL: [%s]\n", server.status, server.timeSpent, server.serverURL)
	}
	return downServers

}

func openFiles(serverListFile string, downtimeFile string) (*os.File, *os.File) {
	serverList, err := os.OpenFile(serverListFile, os.O_RDONLY, 0666)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	downtimeList, err := os.OpenFile(downtimeFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return serverList, downtimeList
}

func generateDowntime(downtimeList *os.File, downServers []Server) {
	csvWriter := csv.NewWriter(downtimeList)
	for _, server := range downServers {
		line := []string{server.serverName, server.serverURL, server.failTime, fmt.Sprintf("%f", server.timeSpent), fmt.Sprintf("%d", server.status)}
		csvWriter.Write(line)
	}
	csvWriter.Flush()
}

func main() {
	serverList, downtimeList := openFiles(os.Args[1], os.Args[2])
	defer serverList.Close()
	defer downtimeList.Close()
	servers := createListServers(serverList)

	downServers := checkServer(servers)
	generateDowntime(downtimeList, downServers)

}
