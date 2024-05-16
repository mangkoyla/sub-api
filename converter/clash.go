package converter

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func shuffleArray(array []interface{}) []interface{} {
	rand.Seed(time.Now().UnixNano())
	for i := len(array) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		array[i], array[j] = array[j], array[i]
	}
	return array
}

func toClash(accounts []interface{}, params url.Values) string {
	vpnType := strings.Split(params.Get("vpn"), ",")
	countryCodes := strings.Split(params.Get("cc"), ",")
	nw := strings.Split(params.Get("nw"), ",")
	region := strings.Split(params.Get("region"), ",")
	ports := strings.Split(params.Get("port"), ",")
	cdn := params.Get("cdn")
	sni := params.Get("sni")

	var filteredAccounts []interface{}

	for _, account := range accounts {
		accMap := account.(map[string]interface{})
		accountVPN := accMap["vpn"].(string)
		accountCountryCode := accMap["country_code"].(string)
		accountServerPort := strconv.Itoa(int(accMap["server_port"].(float64)))
		accountTransport := accMap["transport"].(string)
		accountRegion := accMap["region"].(string)
		accountConnMode := accMap["conn_mode"].(string)
		if (len(vpnType) == 0 || contains(vpnType, accountVPN)) &&
			(len(countryCodes) == 0 || contains(countryCodes, accountCountryCode)) &&
			(len(ports) == 0 || contains(ports, accountServerPort)) &&
			(len(nw) == 0 || contains(nw, accountTransport)) &&
			(len(region) == 0 || contains(region, accountRegion)) &&
			(accountConnMode == "" || !params.Has("mode") || params.Get("mode") == accountConnMode) {
			filteredAccounts = append(filteredAccounts, account)
		}
	}

	limit, _ := strconv.Atoi(params.Get("limit"))
	if limit == 0 {
		limit = len(filteredAccounts)
	}

	var result []string
	result = append(result, "proxies:")

	shuffledAccounts := shuffleArray(filteredAccounts)

	for _, account := range shuffledAccounts[:limit] {
		accMap := account.(map[string]interface{})
		proxy := []string{}
		proxy = append(proxy, fmt.Sprintf("  - name: %v", accMap["remark"]))
		proxy = append(proxy, fmt.Sprintf("    server: %v", accMap["server"]))
		proxy = append(proxy, fmt.Sprintf("    port: %v", int(accMap["server_port"].(float64))))
		proxy = append(proxy, fmt.Sprintf("    type: %v", accMap["vpn"]))

		switch accMap["vpn"] {
		case "vmess", "vless":
			proxy = append(proxy, fmt.Sprintf("    uuid: %v", accMap["uuid"]))
			proxy = append(proxy, "    cipher: auto")
			proxy = append(proxy, fmt.Sprintf("    tls: %v", accMap["tls"] == "1"))
			proxy = append(proxy, "    udp: true")
			proxy = append(proxy, "    skip-cert-verify: true")
			proxy = append(proxy, fmt.Sprintf("    servername: %v", accMap["host"]))
			proxy = append(proxy, fmt.Sprintf("    network: %v", accMap["transport"]))
			if accMap["vpn"] == "vmess" {
				proxy = append(proxy, fmt.Sprintf("    alterId: %v", accMap["alter_id"]))
			}
		case "trojan":
			proxy = append(proxy, fmt.Sprintf("    password: %v", accMap["password"]))
			proxy = append(proxy, "    udp: true")
			proxy = append(proxy, "    skip-cert-verify: true")
			proxy = append(proxy, fmt.Sprintf("    sni: %v", accMap["server"]))
			proxy = append(proxy, fmt.Sprintf("    network: %v", accMap["transport"]))
		case "shadowsocks":
			proxy = append(proxy, "    type: ss")
			proxy = append(proxy, fmt.Sprintf("    cipher: %v", accMap["method"]))
			proxy = append(proxy, fmt.Sprintf("    password: %v", accMap["password"]))
		case "shadowsocksr":
			proxy = append(proxy, "    type: ssr")
			proxy = append(proxy, fmt.Sprintf("    cipher: %v", accMap["method"]))
			proxy = append(proxy, fmt.Sprintf("    password: %v", accMap["password"]))
			proxy = append(proxy, fmt.Sprintf("    obfs: %v", accMap["obfs"]))
			proxy = append(proxy, fmt.Sprintf("    obfs-param: %v", accMap["obfs_param"]))
			proxy = append(proxy, fmt.Sprintf("    protocol: %v", accMap["protocol"]))
			proxy = append(proxy, fmt.Sprintf("    protocol-param: %v", accMap["protocol_param"]))
			proxy = append(proxy, "    udp: true")
		}

		switch accMap["transport"] {
		case "ws", "websocket":
			proxy = append(proxy, "    ws-opts:")
			proxy = append(proxy, fmt.Sprintf("      path: %v", accMap["path"]))
			proxy = append(proxy, "      headers:")
			proxy = append(proxy, fmt.Sprintf("        Host: %v", accMap["host"]))
		case "grpc":
			proxy = append(proxy, "    grpc-opts:")
			proxy = append(proxy, fmt.Sprintf("      grpc-service-name: %v", accMap["service_name"]))
		}

		result = append(result, strings.Join(proxy, "\n"))
	}

	return strings.Join(result, "\n")
}

func contains(slice []string, element string) bool {
	for _, item := range slice {
		if item == element {
			return true
		}
	}
	return false
}
