package converter

import (
	"encoding/base64"
	"encoding/json"
	"net/url"
	"strconv"
	"strings"
)

func toRaw(accounts []interface{}, params url.Values) string {
	vpnType := strings.Split(params.Get("vpn"), ",")
	countryCodes := strings.Split(params.Get("cc"), ",")
	nw := strings.Split(params.Get("nw"), ",")
	region := strings.Split(params.Get("region"), ",")
	ports := strings.Split(params.Get("port"), ",")
	cdn := params.Get("cdn")
	sni := params.Get("sni")

	var result []string

	for _, account := range accounts {
		accMap := account.(map[string]interface{})
		accountVPN := accMap["vpn"].(string)
		accountCountryCode := accMap["country_code"].(string)
		accountServerPort := strconv.Itoa(int(accMap["server_port"].(float64)))
		accountTransport := accMap["transport"].(string)
		accountRegion := accMap["region"].(string)
		accountConnMode := accMap["conn_mode"].(string)

		if (len(vpnType) > 0 && !contains(vpnType, accountVPN)) ||
			(len(countryCodes) > 0 && !contains(countryCodes, accountCountryCode)) ||
			(len(ports) > 0 && !contains(ports, accountServerPort)) {
			continue
		}

		if (len(nw) > 0 && !contains(nw, accountTransport)) ||
			(len(region) > 0 && !contains(region, accountRegion)) ||
			(accountConnMode != "" && params.Get("mode") != accountConnMode) {
			continue
		}

		var url string
		switch accountVPN {
		case "vmess":
			url = getVmessURL(accMap, params, cdn, sni)
		case "vless", "trojan":
			url = getOtherURL(accMap, params, cdn, sni, accountVPN)
		case "shadowsocks":
			url = getShadowsocksURL(accMap)
		case "shadowsocksr":
			url = getShadowsocksrURL(accMap)
		default:
			continue
		}

		if url != "" {
			result = append(result, url)
		}
	}

	return strings.Join(result, "\n")
}

func getVmessURL(account map[string]interface{}, params url.Values, cdn, sni string) string {
	serverAddress := account["server"].(string)
	if params.Get("mode") == "cdn" && cdn != "" {
		serverAddress = cdn
	}
	serverHost := account["host"].(string)
	if params.Get("mode") == "sni" && sni != "" {
		serverHost = sni
	}

	vmess := map[string]interface{}{
		"v":      "2",
		"ps":     account["remark"].(string),
		"add":    serverAddress,
		"port":   int(account["server_port"].(float64)),
		"id":     account["uuid"].(string),
		"aid":    int(account["alter_id"].(float64)),
		"net":    account["transport"].(string),
		"path":   account["path"].(string),
		"tls":    account["tls"].(bool),
		"host":   account["host"].(string),
		"sni":    serverHost,
		"security": func() string {
			if account["security"] != nil {
				return account["security"].(string)
			}
			return ""
		}(),
		"allowInsecure": func() bool {
			if account["insecure"] != nil {
				return account["insecure"].(bool)
			}
			return false
		}(),
		"type": account["transport"].(string),
		"wsHeaders": map[string]string{
			"Host": account["host"].(string),
		},
	}

	if account["transport"].(string) == "grpc" {
		vmess["path"] = account["service_name"].(string)
	}

	vmessJSON, _ := json.Marshal(vmess)
	vmessBase64 := base64.StdEncoding.EncodeToString(vmessJSON)
	return "vmess://" + vmessBase64
}

func getOtherURL(account map[string]interface{}, params url.Values, cdn, sni, protocol string) string {
	serverAddress := account["server"].(string)
	if params.Get("mode") == "cdn" && cdn != "" {
		serverAddress = cdn
	}
	serverHost := account["host"].(string)
	if params.Get("mode") == "sni" && sni != "" {
		serverHost = sni
	}

	url := protocol + "://" + account["password"].(string) + "@" + serverAddress + ":" + strconv.Itoa(int(account["server_port"].(float64)))

	tls := ""
	if account["tls"] != nil && account["tls"].(bool) {
		tls = "tls"
	}

	switch protocol {
	case "vless", "trojan":
		url += "?security=" + tls + "&type=" + account["transport"].(string) + "&sni=" + account["sni"].(string) + "&allowInsecure=" + strconv.FormatBool(account["insecure"].(bool))
		switch account["transport"].(string) {
		case "ws":
			url += "&host=" + serverHost + "&path=" + url.QueryEscape(account["path"].(string)) + "#" + url.QueryEscape(account["remark"].(string))
		case "grpc":
			url += "&serviceName=" + url.QueryEscape(account["service_name"].(string)) + "#" + url.QueryEscape(account["remark"].(string))
		}
	}

	return url
}

func getShadowsocksURL(account map[string]interface{}) string {
	cred := base64.StdEncoding.EncodeToString([]byte(account["method"].(string) + ":" + account["password"].(string)))
	var plugin string
	if account["plugin"] != nil {
		plugin = "?plugin=" + account["plugin"].(string) + ";" + account["plugin_opts"].(string)
	}
	return "ss://" + cred + "@" + account["server"].(string) + ":" +
		strconv.Itoa(int(account["server_port"].(float64))) + plugin + "#" + url.QueryEscape(account["remark"].(string))
}

func getShadowsocksrURL(account map[string]interface{}) string {
	password := base64.StdEncoding.EncodeToString([]byte(account["password"].(string)))
	remarks := base64.StdEncoding.EncodeToString([]byte(account["remark"].(string)))
	protoParam := base64.StdEncoding.EncodeToString([]byte(account["protocol_param"].(string)))
	obfsParam := base64.StdEncoding.EncodeToString([]byte(account["obfs_param"].(string)))
	return "ssr://" + base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%d:%s:%s:%s:%s",
		account["server"].(string), int(account["server_port"].(float64)),
		account["protocol"].(string), account["method"].(string),
		account["obfs"].(string), password)))) +
		"?remarks=" + remarks + "&protoparam=" + protoParam + "&obfsparam=" + obfsParam
}

func contains(slice []string, element string) bool {
	for _, item := range slice {
		if item == element {
			return true
		}
	}
	return false
}
