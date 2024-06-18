package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

// headers 是一个自定义的类型，用于存储多个 Header
type headers map[string]string

// String 方法用于实现 flag.Value 接口中的 String 方法
func (h *headers) String() string {
	var result []string
	for key, val := range *h {
		result = append(result, fmt.Sprintf("%s: %s", key, val))
	}
	return strings.Join(result, ", ")
}

// Set 方法用于实现 flag.Value 接口中的 Set 方法，用于解析并存储 Header
func (h *headers) Set(value string) error {
	parts := strings.SplitN(value, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid header format, should be key:value")
	}
	key := strings.TrimSpace(parts[0])
	val := strings.TrimSpace(parts[1])
	(*h)[key] = val
	return nil
}

func main() {
	var customHeaders headers = make(map[string]string)
	// 定义一个 -H 参数，用于设置自定义 Header
	flag.Var(&customHeaders, "H", "HTTP header in key:value format")
	env := flag.String("h", "", "Host eg: https://www.baidu.com embed prod or dev or local")
	path := flag.String("p", "", "Request path")
	method := flag.String("X", "GET", "HTTP method")
	port := flag.Int("P", 59447, "Port")
	body := flag.String("d", "", "Request body")
	token := flag.String("t", "", "Authorization token")

	flag.Parse()
	// 如果用户没有输入参数，则打印帮助信息
	if *env == "" {
		flag.Usage()
		os.Exit(0)
	}

	if token != nil {
		customHeaders["Authorization"] = *token
	}

	requestUrl := getRequestUrl(*env, *port, *path)
	doRequest(*method, requestUrl, *body, &customHeaders)
}

func doRequest(method string, url string, reqBody string, h *headers) {
	req, err := http.NewRequest(method, url, strings.NewReader(reqBody))
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}
	if method == "POST" || method == "PUT" || method == "DELETE" {
		req.Header.Add("Content-Type", "application/json")
	}

	for key, val := range *h {
		req.Header.Add(key, val)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to perform request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}

	fmt.Printf("Response from %s:\n%s\n", url, string(body))
}

func getRequestUrl(env string, port int, path string) string {
	url := ""
	switch env {
	case "prod":
		url = "https://www.prod.com" // 替换为你的生产环境 url
	case "dev":
		url = "https://www.dev.com" // 替换为你的开发环境 url
	case "local":
		url = "http://localhost:" + fmt.Sprint(port) // 替换为你的本地环境 url
	default:
		url = env
	}
	url = url + path
	return url
}
