package main

/*
 * kube-file is a stub to handle the disk/os side of kube-addons by
 * going through a specified directory and checking for changes.
 *
 * Currently, I'm using a metadata file to contain the filename and
 * timestamp of registration to provide a basis for comparison, but
 * this will be replaced with calls into k8s to determine if the
 * yaml/json specified add-on is registered.
 *
 * Things to do for this file:
 * - Daemonize such that it runs in the background and logs its actions
 *   to a proper logger
 * - Add hooks to k8s.
 */

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"strconv"
	"time"
)

type Status struct {
	name string
	ts time.Time
}

func ReadMetadata(filename string) (map[string]Status, error) {
	status := make(map[string]Status)
	file, err := ioutil.ReadFile(filename)
	f := string(file[:])

	if os.IsNotExist(err) {
		return status, nil
	}

	if err != nil {
		return nil, err
	}

	/* file format is name,ts */
	for _, line := range strings.Split(f, "\n") {
		if len(line) == 0 {
			break
		}

		vals := strings.Split(line, ",")
		ts, err := strconv.ParseInt(vals[1], 10, 64); if err != nil {
			return nil, err
		}

		t := time.Unix(ts, 0)
		status[vals[0]] = Status{vals[0], t}
	}

	return status, nil
}

func WriteMetadata(filename string, status map[string]Status) error {
	var str string

	file, err := os.Create(filename)
	if err != nil {
		return err
	}

	defer file.Close()

	for _,v := range status {
		var s = make([]string, 2)

		s[0] = v.name
		s[1] = strconv.FormatInt(v.ts.Unix(), 10)

		str = str + strings.Join(s , ",") + "\n"
	}

	err = file.Truncate(0)
	if err != nil {
		return err
	}

	buffer := bytes.NewBufferString(str)
	_, err = file.Write(buffer.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Invalid number of arguments.  Expect a directory to parse")
	}

	dir, err := ioutil.ReadDir(os.Args[1]); if err != nil {
		log.Fatal(err)
	}

	md, err := ReadMetadata(os.Args[1] + ".metadata")
	if err != nil {
		log.Fatal(err)
	}

	dirStatus := make(map[string]Status)
	for _, v := range dir {
		d := Status{v.Name(), v.ModTime()}
		m, mExists := md[v.Name()]

		dirStatus[v.Name()] = d

		if !mExists {
			fmt.Printf("Adding: %s:%d\n", d.name, d.ts.Unix())
			md[v.Name()] = d
		} else if m.ts.Unix() < d.ts.Unix() {
			fmt.Printf("Updating: %s:%d\n", d.name, d.ts.Unix())
			md[v.Name()] = d
		}
	}

	for k, _ := range md {
		_, exists := dirStatus[k]; if !exists {
			fmt.Printf("Deleting: %s", k)
			delete(md, k)
		}
	}

	err = WriteMetadata(os.Args[1] + ".metadata", md)
	if err != nil {
		log.Fatal(err)
	}
}
