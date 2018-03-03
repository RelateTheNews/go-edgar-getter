// Copyright 2018 Interconnect Analytics DBA RelateTheNews. All rights reserved.
// Use of this source code is governed by a Apache 2.0
// license that can be found in the LICENSE file.

// Package getter is used to manage retrieval, storge and decompression of Edgar filings from the SEC.
package getter

import (
	"edgarparser/pkgs/telemetry"
	"expvar"
	"github.com/anaskhan96/soup"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	// MAXRETRIEVALSIZE the maximum number of files to retrieve at any one time
	MAXRETRIEVALSIZE = 5000
	// MAXRETRY the maximum number of attempts to download a single resource
	MAXRETRY = 5
)

var (
	fileCounter, duration expvar.Int
)

// Getter provides a data structure to support this package
type Getter struct {
	RetrieveURI       string
	SaveLocation      string
	ValidFileSuffixes map[string]bool // ValidFileSuffixes lists the ONLY valid suffixes allowed when downloading
	Telemetry         *telemetry.Telemetry
}

// NewGetter initializes and performs any setup necessary for the getter package to function.
//
// Items currently Setup:
//   SaveLocation set to a default value of /tmp/edgar/
//   ValidFileSuffixes set to default values
func (getter *Getter) NewGetter() {
	getter.SaveLocation = "/tmp/edgar/"

	getter.ValidFileSuffixes = map[string]bool{
		"tgz":  true,
		"gz":   true,
		"xls":  true,
		"xlsx": true,
		"doc":  true,
		"docx": true}

}

// ErrorHandler a simple error checker and handler which panics if an error exists.
func (getter *Getter) ErrorHandler(e error) {
	if e != nil {
		panic(e)
	}
}

// DownloadableFile checks a file against the approved file suffix list
func (getter *Getter) DownloadableFile(filename string) bool {
	tokens := strings.Split(filename, ".")
	suffix := tokens[len(tokens)-1]

	if getter.ValidFileSuffixes[suffix] {
		return true
	}

	return false
}

// RetrieveSingleFile retrieves a single file/resource from the specified location.
// This function can be called concurrently by providing a channel.
func (getter *Getter) RetrieveSingleFile(location string, fileList chan<- string) (string, error) {
	errorCount := 0
	tokens := strings.Split(location, "/")
	fileName := tokens[len(tokens)-1]

	if !getter.DownloadableFile(fileName) {
		log.Println("File", fileName, "not allowed.")
		return "", nil
	}

	log.Println("Downloading", location, "to", fileName)

	output, err := os.Create(getter.SaveLocation + fileName)
	if err != nil {
		log.Println("Error while creating", fileName, "-", err)
		return "", nil
	}
	defer output.Close()

	response, err := http.Get(location)
	if err != nil {
		errorCount = errorCount + 1
		log.Println("Error while downloading", location, "-", err)
		log.Println("Retrying ", location, " Error Count: ", errorCount)

		for ok := true; ok; ok = (errorCount <= MAXRETRY) {
			response, err = http.Get(location)
			if err != nil {
				errorCount = errorCount + 1
				log.Println("Error while downloading", location, "-", err)
				log.Println("Retrying ", location, " Error Count: ", errorCount)
			}
		}
	}
	defer response.Body.Close()

	n, err := io.Copy(output, response.Body)
	if err != nil {
		log.Println("Error while copying file ", location, "-", err)
		return "", nil
	}

	log.Printf("File %v of %v bytes downloaded from %v\n", fileName, n, location)

	if fileList != nil {
		fileList <- fileName
	}
	return fileName, nil
} //RetrieveSingleFile

// RetrieveURIs concurrently retrieves all resources from the specified source
// returning the filenames retrieved.
func (getter *Getter) RetrieveURIs(sourceURI string, customLimit int) []string {
	var wg sync.WaitGroup
	var expvarTime *telemetry.TimeVar
	var links []soup.Root
	var single bool
	var i = 0
	var useLimit bool

	if customLimit == 0 {
		useLimit = false
	} else {
		useLimit = true
	}

	t := strings.Split(sourceURI, "/")
	filename := t[len(t)-1]

	start := time.Now()
	expvarTime = &telemetry.TimeVar{Value: start}
	if getter.Telemetry != nil {
		getter.Telemetry.TelemetryData.Set("RetrieveURIs Start Time", expvarTime)
	}
	if strings.Contains(filename, ".") {
		// sourceURI is a single file
		single = true
		doc := soup.HTMLParse("<a href='" + sourceURI + "'>" + sourceURI + "</a>")
		links = doc.FindAll("a")
	} else {
		resp, err := soup.Get(sourceURI)
		if err != nil {
			os.Exit(1)
		}
		doc := soup.HTMLParse(resp)
		links = doc.Find("table").FindAll("a")
	}

	log.Printf("Retrieved this many links %v", len(links))

	if len(links) > MAXRETRIEVALSIZE {
		log.Printf("Unable to continue. More files (%v) requested to download than allowed(%v)", len(links), MAXRETRIEVALSIZE)
		return nil
	}

	// NOTE remember that unbuffered channels hang
	allFileLocations := make(chan string, len(links))
	for _, link := range links {
		switch useLimit {
		case true:
			if i != customLimit {
				var fileURI string
				wg.Add(1)
				log.Printf("%v| Link: %v", link.Text(), link.Attrs()["href"])

				if single {
					fileURI = link.Attrs()["href"]
				} else {
					fileURI = sourceURI + link.Attrs()["href"]
				}

				go func(file string, wg *sync.WaitGroup) {
					defer wg.Done()
					if getter.Telemetry != nil {
						getter.Telemetry.TelemetryData.Add("Total URIs", 1)
					}
					getter.RetrieveSingleFile(file, allFileLocations)
				}(fileURI, &wg)

				i++
			}
			break
		case false:
			var fileURI string
			wg.Add(1)
			log.Printf("%v| Link: %v", link.Text(), link.Attrs()["href"])

			if single {
				fileURI = link.Attrs()["href"]
			} else {
				fileURI = sourceURI + link.Attrs()["href"]
			}

			go func(file string, wg *sync.WaitGroup) {
				defer wg.Done()
				if getter.Telemetry != nil {
					getter.Telemetry.TelemetryData.Add("Total URIs", 1)
				}
				getter.RetrieveSingleFile(file, allFileLocations)
			}(fileURI, &wg)
		} //switch
	} //for

	wg.Wait()
	close(allFileLocations)
	elapsed := time.Since(start)
	if getter.Telemetry != nil {
		getter.Telemetry.TelemetryData.Add("RetrieveURIs(ns)", int64(elapsed))
	}
	var fileList []string
	for file := range allFileLocations {
		fileList = append(fileList, file)
	}
	return fileList
}
