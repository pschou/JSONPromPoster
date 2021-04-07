package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/pschou/go-params"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

type masdat struct {
	Network    string `json:"network"`
	Observer   string `json:"observer"`
	Timestamp  uint64 `json:"timestamp"`
	Host       string `json:"host"`
	Key        string `json:"key"`
	Version    string `json:"@version"`
	Properties struct {
		Engine string `json:"engine"`
		Host   string `json:"host"`
		System string `json:"system"`
		Units  string `json:"units"`
		Site   string `json:"site"`
	} `json:"properties"`
	Value interface{} `json:"value"`
}
type site_system struct {
	Site   string
	System string
	Bytes  *bytes.Buffer
}

func fatalOnErr(err error, fmt string, vals ...interface{}) {
	if err != nil {
		log.Fatalf(fmt, vals...)
	}
}

var version = "0.0"

func main() {

	params.Usage = func() {
		fmt.Fprintf(os.Stderr, "JSON Prom Poster, written by Paul Schou (github.com/pschou/JSONPromPoster), Version: %s\n\n"+
			"Usage: %s [options...] [file]\n\n", version, os.Args[0])
		params.PrintDefaults()
	}
	var urls = params.StringSlice("post", "Endpoint to upload metrics to (use multiple times for multiple endpoints)", "URL", 1)

	// Indicate that we want all the flags indented for ease of reading
	params.CommandLine.Indent = 2

	// Let us parse everything!
	params.Parse()

	args := params.Args()
	if len(args) == 0 {
		params.Usage()
		os.Exit(1)
	}

	// buffer for all the sites seen
	var allSeen []site_system

	for _, file := range args {

		// Open our jsonFile
		jsonFile, err := os.Open(file)
		fatalOnErr(err, "Unable to open %s: %s", file, err)

		// defer the closing of our jsonFile so that we can parse it later on
		defer jsonFile.Close()

		//byteValue, err := ioutil.ReadAll(jsonFile)
		//fatalOnErr(err, "Unable to read in %s: %s", file, err)

		// JSON decoder, load from file stream
		decoder := json.NewDecoder(jsonFile)

		// regex to fix name to be compatible
		reName := regexp.MustCompile(`[^a-z0-9_]+`)

		for {
			// empty JSON structure
			var result = masdat{}

			// decode the next blob
			err = decoder.Decode(&result)
			if err == io.EOF {
				break
			} else if err != nil {
				fmt.Printf("Bad JSON / Unable to unmarhall %s: %s", file, err)
				continue
			}

			var cur_site_system site_system
			isNew := true
			for _, seen := range allSeen {
				if result.Properties.Site == seen.Site && result.Properties.System == seen.System {
					isNew = false
					cur_site_system = seen
					break
				}
			}
			if isNew {
				cur_site_system = site_system{
					Site:   result.Properties.Site,
					System: result.Properties.System,
					Bytes:  &bytes.Buffer{}}
				allSeen = append(allSeen, cur_site_system)
				cur_site_system.Bytes.WriteString(fmt.Sprintf("last_report_timestamp_seconds{site=%q,system=%q} %v %v\n",
					result.Properties.Site, result.Properties.System, result.Timestamp, time.Now().UnixNano()/1e6))
			}

			keyParts := strings.SplitN(result.Key, "/", 2)
			if len(keyParts) == 2 {
				name := reName.ReplaceAllString(strings.ToLower(keyParts[1]), "_")
				cur_site_system.Bytes.WriteString(fmt.Sprintf("%s{engine=%q,host=%q,site=%q,system=%q} %v %v\n",
					name, strings.ToLower(keyParts[0]), result.Properties.Host, result.Properties.Site, result.Properties.System,
					result.Value, result.Timestamp*1e3,
				))
			}
		}
	}

	// Post this to all the URLs available
	for _, url := range []string(*urls) {
		//fmt.Printf("urls = %#v", urls)
		fmt.Printf("Sending metrics to %v\n", url)
		for _, ss := range allSeen {
			fmt.Printf("  site: %v\tsystem: %v\n", ss.Site, ss.System)
			request, err := http.NewRequest("POST",
				fmt.Sprintf("%s/site/%s/system/%s", strings.TrimSuffix(url, "/"), ss.Site, ss.System),
				ss.Bytes)
			client := &http.Client{
				Transport: &http.Transport{TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				}},
			}

			response, err := client.Do(request)
			if err != nil {
				continue
			}
			defer response.Body.Close()

			content, err := ioutil.ReadAll(response.Body)
			if err != nil {
				continue
			}

			fmt.Println(string(content))
			//fmt.Printf("%#v\n\n", site)
			//fmt.Println(site.Bytes.String())
		}
	}
}
