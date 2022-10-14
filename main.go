package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

const settingsFilename = "settings.json"
const inputDir = "test-input/"
const outputDir = "test-output/"

type Settings struct {
	BaseUrl   string   `json:"baseurl"`
	UseSSL    bool     `json:"usessl"`
	Endpoints []string `json:"endpoints"`
}

type UploadData struct {
	filename string
	file     []byte
}

func InitSettings() Settings {
	//Initilize settings
	var settings Settings

	//Try to open settings file
	settingsFile, err := os.Open(settingsFilename)
	if err != nil {
		//Create new settingsfile
		fmt.Println("No settings file! Let's create one")

		reader := bufio.NewReader(os.Stdin)
		index := 0
		for {
			fmt.Println("------------------")
			if index == 0 {
				fmt.Println("What base url do you want?")
				fmt.Print("--> ")

				text, _ := reader.ReadString('\n')
				text = strings.Replace(text, "\n", "", -1)

				settings.BaseUrl = text

				index++
			} else if index == 1 {
				fmt.Println("Do you want use ssl verification?")
				fmt.Print("--> ")

				text, _ := reader.ReadString('\n')
				text = strings.Replace(text, "\n", "", -1)
				boolValue, err := strconv.ParseBool(text)

				if err != nil {
					settings.UseSSL = false
				} else {
					settings.UseSSL = boolValue
				}

				index++
			} else if index == 2 {
				fmt.Println("What endpoints do you want to test(seperated by ,)?")
				fmt.Print("--> ")

				text, _ := reader.ReadString('\n')
				text = strings.Replace(text, "\n", "", -1)
				arrValue := strings.Split(text, ",")

				for i := 0; i < len(arrValue); i++ {
					settings.Endpoints = append(settings.Endpoints, arrValue[i])
				}

				index++
			} else {
				break
			}
		}

		file, _ := json.MarshalIndent(settings, "", "")
		_ = ioutil.WriteFile(settingsFilename, file, 0644)
		fmt.Println("Settings file created")
	} else {
		// Load settingsfile
		jsonByteValue, _ := ioutil.ReadAll(settingsFile)
		json.Unmarshal(jsonByteValue, &settings)
	}

	defer settingsFile.Close()

	return settings
}

func ChooseEndpoint(settings Settings) string {
	if len(settings.Endpoints) == 0 {
		fmt.Println("What endpoint do you want to use?")

		reader := bufio.NewReader(os.Stdin)
		choosen := false
		for {
			if choosen {
				break
			}

			fmt.Print("-->")

			text, _ := reader.ReadString('\n')
			text = strings.Replace(text, "\n", "", -1)
			return text
		}
	} else if len(settings.Endpoints) == 1 {
		return settings.Endpoints[0]
	} else {
		reader := bufio.NewReader(os.Stdin)

		for i := 0; i < len(settings.Endpoints); i++ {
			fmt.Println(fmt.Sprintf("%d: ", i+1) + settings.Endpoints[i])
		}

		fmt.Println("What endpoint do you want to use?")
		choosen := 0

		for {
			if choosen > 0 {
				break
			}

			fmt.Print("-->")

			text, _ := reader.ReadString('\n')
			text = strings.Replace(text, "\n", "", -1)
			intValue, err := strconv.Atoi(text)

			if err != nil || intValue > len(settings.Endpoints) {
				choosen = 1
			} else {
				choosen = intValue
			}
		}

		return settings.Endpoints[choosen-1]
	}
	return ""
}

func GrabJson(pathname string) []byte {
	jsonFile, err := os.Open(pathname)

	if err != nil {
		return nil
	} else {
		jsonByteValue, _ := ioutil.ReadAll(jsonFile)
		return jsonByteValue
	}
}

func Upload(settings Settings, endpoint string, uploadFile UploadData) {
	fmt.Println("Uploading --> " + uploadFile.filename)

	requestBody := bytes.NewBuffer(uploadFile.file)
	resp, err := http.Post(settings.BaseUrl+endpoint, "application/json", requestBody)

	if err != nil {
		fmt.Println("Post failed")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		fmt.Println("Failed to read body")
	} else {
		ioutil.WriteFile(outputDir+uploadFile.filename, body, 0644)
	}

}

func InitUpload(settings Settings) {
	useEndpoint := ChooseEndpoint(settings)

	dirItems, _ := ioutil.ReadDir(inputDir)

	var filesToUpload []UploadData

	for _, item := range dirItems {
		if strings.Contains(item.Name(), ".geojson") || strings.Contains(item.Name(), ".json") {
			byteArray := GrabJson(inputDir + item.Name())
			filesToUpload = append(filesToUpload, UploadData{item.Name(), byteArray})

		}
	}

	var wg sync.WaitGroup
	wg.Add(len(filesToUpload))

	for i := 0; i < len(filesToUpload); i++ {
		go func(i int) {
			Upload(settings, useEndpoint, filesToUpload[i])
			wg.Done()
		}(i)
	}

	wg.Wait()
}

func main() {
	settings := InitSettings()

	InitUpload(settings)
}
