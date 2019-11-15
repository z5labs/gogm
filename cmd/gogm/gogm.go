package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)



func main() {
	directory := "/home/erictg97/mindstand/repos/gogm/testing_"


	confs := map[string][]*relConf{}
	imps := map[string][]string{}

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		if err != nil {
			log.Println("failed here")
			return err
		}

		if strings.Contains(path, ".go") {
			err := parseFile(path, &confs, imps)
			if err != nil {
				log.Fatal(err)
			}
		}

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Println(confs)
	log.Println(imps)
}

