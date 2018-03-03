// Copyright 2018 Interconnect Analytics DBA RelateTheNews. All rights reserved.
// Use of this source code is governed by a Apache 2.0
// license that can be found in the LICENSE file.

// Package getter is used to manage retrieval, storge and decompression of Edgar filings from the SEC.
package getter

import (
	"testing"
)

const (
	testURI       = "https://www.sec.gov/Archives/edgar/Feed/2013/QTR1/"
	testSingleURI = "https://www.sec.gov/Archives/edgar/Feed/2013/QTR1/20130109.nc.tar.gz"
)

func TestRetrieveURIs(t *testing.T) {
	var g Getter

	g.NewGetter()

	files := g.RetrieveURIs(testURI, 1)

	if len(files) != 1 {
		t.Errorf("Number of files was incorrect, got: %d, want: %d.", len(files), 1)
	}

}

func TestRetrieveURIs_Single(t *testing.T) {
	var g Getter

	g.NewGetter()
	files := g.RetrieveURIs(testSingleURI, 0)

	if len(files) != 1 {
		t.Errorf("Number of files was incorrect, got: %d, want: %d.", len(files), 1)
	}
}

func TestRetrieveSingleFile(t *testing.T) {
	var g Getter

	g.NewGetter()

	files, _ := g.RetrieveSingleFile(testSingleURI, nil)
	if len(files) != 18 {
		// NOTE different length here because its the length of the string returned
		// not the slice
		t.Errorf("Number of files was incorrect, got: %d, want: %d.", len(files), 18)
	}
}

func TestDownloadableFile(t *testing.T) {
	var g Getter
	var allowed = false

	g.NewGetter()

	allowedFile := "testfile.gz"
	disallowedFile := "testfile.js"

	allowed = g.DownloadableFile(allowedFile)

	if !allowed {
		t.Errorf("File %v should have been allowed.", allowedFile)
	}

	allowed = true

	allowed = g.DownloadableFile(disallowedFile)

	if allowed {
		t.Errorf("File %v should have been disallowed.", allowedFile)
	}
}


