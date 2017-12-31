package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

func main() {
	command := "help"
	fmt.Print("Commands: (help, config, fetch, sync, exit): ")
	isConfigSet()
	fmt.Scanln(&command)
	execute(command)
}

func execute(commandName string) {
	funcs := map[string]func(){"fetch": fetch,
		"help": help, "sync": sync, "config": config, "exit": exit}
	command := funcs[commandName]
	if command == nil {
		help()
	}
	command()
}

func fetch() {
	var serviceName string
	var rules string
	fmt.Println("")
	fmt.Print("Fetch which Service Rule? : ")
	fmt.Scanln( &serviceName)

	fmt.Println("")
	fmt.Println("Fetching ... ")
	rules = getServiceRule(serviceName)
	persistToFile(serviceName, rules)
	fmt.Println("Done ! Find it in the service_rules folder :) ")

}

func help() {
	fmt.Println("")
	fmt.Println("help: Shows you this")
	fmt.Println("config: Set Devless instance details ")
	fmt.Println("fetch: get rules for a service")
	fmt.Println("sync: sync service files with Devless instance")
	fmt.Println("")
	main()
}

func sync() {
	var serviceName string
	println("")
	print("Which service will you like to sync? : ")
	fmt.Scanln( &serviceName)
	rules := getRulesFromFile(serviceName)
	pushToUpstream(serviceName, rules)

}

func config() {
	var instUrl string
	var extEditorKey string

	fmt.Print("Enter Instance URL: ")
	fmt.Scanln( &instUrl)
	fmt.Println("")
	fmt.Print("Enter extEditor key: ")
	fmt.Scanln( &extEditorKey)

	config := []byte("{\"instanceURL\":\"" + instUrl + "\",\"extEditorKey\":\"" + extEditorKey + "\"}")
	err := ioutil.WriteFile("config.json", config, 0644)
	if err == nil {
		fmt.Println("")
		fmt.Println("Configuration set successfully")
		fmt.Println("")
		main()

	} else {
		fmt.Print(err)
	}
}

func isConfigSet() {
	_, err := ioutil.ReadFile("config.json")
	if err != nil {
		fmt.Println("")
		fmt.Println("You need to configure first ...")
		fmt.Println("")
		config()
	}

}

func getConfig() (string, string) {
	type configStruct struct {
		InstanceURL  string
		ExtEditorKey string
	}

	var config configStruct
	configByt, _ := ioutil.ReadFile("config.json")

	json.Unmarshal(configByt, &config)
	return config.InstanceURL, config.ExtEditorKey
}
func exit() {
	os.Exit(1)
}

func getServiceRule(serviceName string) string {

	instUrl, extEditorKey := getConfig()
	url := instUrl + "/service/extEditor/view/index/?service_name=" + serviceName + "&extEditor_key=" + extEditorKey

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("content-type", "application/json")
	req.Header.Add("cache-control", "no-cache")

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	return string(body)
}

func pushToUpstream(serviceName string, rules string) {
	type JSONStruct struct {
		Rules string `json:"rules"`
	}

	instanceURL, extEditorKey := getConfig()

	url := instanceURL + "/service/extEditor/view/index/?service_name=" + serviceName + "&extEditor_key=" + extEditorKey
	var data JSONStruct
	data.Rules = rules
	jsonPayload, _ := json.Marshal(data)
	payload := strings.NewReader(string(jsonPayload))
	req, _ := http.NewRequest("POST", url, payload)

	req.Header.Add("content-type", "application/json")
	req.Header.Add("cache-control", "no-cache")

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		println("Write to Instance failed :(")
		exit()
	}
	print(string(body))
	println("Write to Instance was successful :)")
	exit()
}

func persistToFile(serviceName string, data string) {
	type payloadStruct struct {
		Rules string
	}
	var payload payloadStruct
	rules := []byte(data)

	json.Unmarshal(rules, &payload)
	if _, err := os.Stat("service_rules"); os.IsNotExist(err) {
		os.Mkdir("service_rules", 0644)
	}

	err := WriteStringToFile("service_rules/"+serviceName+".rl", payload.Rules)
	if err != nil {
		fmt.Println("")
		fmt.Println(err)
		fmt.Println("Something snapped. :( Please try again")
		fmt.Println("")
		main()
	}
}

func WriteStringToFile(filepath, s string) error {
	fo, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer fo.Close()

	_, err = io.Copy(fo, strings.NewReader(s))
	if err != nil {
		return err
	}

	return nil
}

func getRulesFromFile(serviceName string) string {
	configByt, err := ioutil.ReadFile("service_rules/" + serviceName + ".rl")
	if err != nil {
		println("Sorry but there is no such service {" + serviceName + " }")
		exit()
	}
	rules := string(configByt)
	return rules
}
