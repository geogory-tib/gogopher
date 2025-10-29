package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func check_if_contains_a_keyword(s string, keywords []string) bool {
	for _, keyword := range keywords {
		if strings.Contains(s, keyword) {
			return true
		}
	}
	return false
}

func main() {
	var gopher_map_data []byte
	scanner := bufio.NewScanner(os.Stdin)
	text := ""
	for scanner.Scan() {
		text += scanner.Text()
	}
	Split_message := strings.Split(text, "\t")
	server_host := Split_message[0]
	server_port := Split_message[1]
	root_dir := Split_message[2]
	keywords := strings.Split(Split_message[3], "+")
	walk_func := func(path string, d fs.DirEntry, err error) error {
		if d.Name() == "gophermap" && !d.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := scanner.Text()
				if len(line) > 0 {
					if check_if_contains_a_keyword(line, keywords) && line[0] != 'i' {
						gopher_map_data = append(gopher_map_data, []byte(line+"\t"+server_host+"\t"+server_port+"\r\n")...)
					}
				}
			}
		}
		return nil
	}
	filepath.WalkDir(root_dir, walk_func)
	fmt.Print(string(gopher_map_data))
}
