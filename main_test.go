package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestRun(t *testing.T) {
	testCases := []struct {
		name     string
		root     string
		cfg      config
		expected string
	}{
		{name: "NoFilter", root: "testdata", cfg: config{ext: "", size: 0, list: true}, expected: "testdata\\dir.log\ntestdata\\dir2\\script.sh\n"},
		//{name: "NoFilter", root: "testdata", cfg: config{ext: "", size: 0, list: true}, expected: filepath.Join("testdata", "dir.log", "\n", "dir2", "script.sh", "\n")},
		{name: "FilterExtensionMatch", root: "testdata", cfg: config{ext: ".log", size: 0, list: true}, expected: "testdata\\dir.log\n"},
		{name: "FilterExtensionSizeMatch", root: "testdata", cfg: config{ext: ".log", size: 10, list: true}, expected: "testdata\\dir.log\n"},
		{name: "FilterExtensionSizeNoMatch", root: "testdata", cfg: config{ext: ".log", size: 20, list: true}, expected: ""},
		{name: "FilterExtensionNoMatch", root: "testdata", cfg: config{ext: ".gz", size: 0, list: true}, expected: ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buffer bytes.Buffer
			if err := run(tc.root, &buffer, tc.cfg); err != nil {
				t.Fatal(err)
			}
			res := buffer.String()
			if tc.expected != res {
				t.Errorf("Expected %q ,got %q instead\n", tc.expected, res)
			}
		})
	}
}

func createTempDir(t *testing.T, files map[string]int) (dirname string, cleanup func()) {
	t.Helper()
	tempDir, err := ioutil.TempDir("", "walktest")
	if err != nil {
		t.Fatal(err)
	}
	for k, n := range files {
		for j := 1; j <= n; j++ {
			fname := fmt.Sprintf("file%d%s", j, k)
			fpath := filepath.Join(tempDir, fname)
			if err := ioutil.WriteFile(fpath, []byte("dummy"), 0644); err != nil {
				t.Fatal(err)
			}
		}
	}
	return tempDir, func() { os.RemoveAll(tempDir) }
}

func TestRunDelExtension(t *testing.T) {
	testCases := []struct {
		name        string
		cfg         config
		extNoDelete string
		nDelete     int
		nNoDelete   int
		expected    string
	}{
		{name: "DeleteExtenstionNoMatch", cfg: config{ext: ".log", del: true}, extNoDelete: ".gz", nDelete: 0, nNoDelete: 10, expected: ""},
		{name: "DeleteExtenstionMatch", cfg: config{ext: ".log", del: true}, extNoDelete: "", nDelete: 0, nNoDelete: 0, expected: ""},
		{name: "DeleteExtenstionMixed", cfg: config{ext: ".log", del: true}, extNoDelete: ".gz", nDelete: 5, nNoDelete: 5, expected: ""},
	}
	var (
		logBuffer bytes.Buffer
	)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.cfg.wLog = &logBuffer
			var buffer bytes.Buffer
			tempDir, p := createTempDir(t, map[string]int{
				tc.cfg.ext:     tc.nDelete,
				tc.extNoDelete: tc.nNoDelete,
			})
			defer t.Cleanup(p)

			if err := run(tempDir, &buffer, tc.cfg); err != nil {
				t.Fatal(err)
			}
			res := buffer.String()

			if tc.expected != res {
				t.Errorf("Expected%q,got %q instead\n ", tc.expected, res)
			}
			filesLeft, err := ioutil.ReadDir(tempDir)
			if err != nil {
				t.Error(err)
			}
			if len(filesLeft) != tc.nNoDelete {
				t.Errorf("Expected %d files left,got %d instead\n", tc.nNoDelete, len(filesLeft))
			}

			expLogLines := tc.nDelete + 1
			lines := bytes.Split(logBuffer.Bytes(), []byte("\n"))
			if len(lines) != expLogLines {
				t.Errorf("Expected %d log lines,got %d instead\n", expLogLines, len(lines))
			}
		})

	}
}

func TestRunArchive(t *testing.T){
	testCases:=[]struct{
		name string
		cfg config
		extNoArchive string
		nArchive int
		nNoArchive int
	}
	{
		{name:"ArchiveExtensionNoMatch",cfg:config{ext:".log"},extNoArchive:".gz",nArchive:0,nNoArchive:10},
		{name:"ArchiveExtensionNoMatch",cfg:config{ext:".log"},extNoArchive:".log",nArchive:10,nNoArchive:0},
		{name:"ArchiveExtensionMixed",cfg:config{ext:".log"},extNoArchive:".log",nArchive:5,nNoArchive:5},

	}
	for _,tc:=range testCases{
		t.Run(tc.name,func(t *testing.T) {
			var buffer bytes.Buffer
			tempDir,cleanup:=createTempDir(t,map[str]int{
				tc.cfg.ext:tc.nArchive,
				tc.extNoArchive:tc.nNoArchive,

			})
			defer cleanup()
			archiveDir,cleanupArchive:=createTempDir(t,nil)
			defer cleanupArchive()

			tc.cfg.archive=archiveDir
			if err:=run(tempDir,&buffer,tc.cfg);err!=nil{
				t.Fatal(err)
			}
			pattern:=filepath.Join(tempDir,fmt.Sprintf("*%s",tc.cfg.ext))
			expFiles,err:=filepath.Glob(pattern)
			if err!=nil{
				t.Fatal(err)
			}
			expOut:=strings.Join(expFiles,"\n")
			res:=strings.TrimSpace(buffer.String())
			if expOut!=res{
				t.Errorf("Expected %q,got%q instead \n",expOut,res)
			}
			filesArchived,err:=ioutil.ReadDir(archiveDir)
			if err!=nil{
				t.Fatal(err)
			}
			if len(filesArchived)!=tc.nArchive{
				t.Errorf("Expected %d files archived,got %d instead",tc.nArchive,len(filesArchived))
			}
		})
	}
}
