package server

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

const (
	READ_DEADLINE         = 50
	FILE_NOT_FOUND_ERROR  = "iFile Not Found\r\n"
	NO_GOPHER_MAP_FOR_DIR = "iDirectory doesn't contain gopher\r\n"
)

type Server struct {
	listener   net.Listener
	addr       string
	server_dir string
	log_file   *os.File
}

func New_Instance() (ret Server) {
	conf_struct := struct {
		Ip           string `json:"ip"`
		Port         string `json:"port"`
		Root_dir     string `json:"root_dir"`
		Log_File_dir string `json:"log_file"`
	}{}
	conf_file, err := os.Open("config.json")
	if err != nil {
		log.Fatal(err)
	}
	defer conf_file.Close()
	conf_json, err := io.ReadAll(conf_file)
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal(conf_json, &conf_struct)
	ret.addr = conf_struct.Ip + ":" + conf_struct.Port
	ret.listener, err = net.Listen("tcp", ret.addr)
	if err != nil {
		log.Fatal(err)
	}
	ret.log_file, err = os.Create(conf_struct.Log_File_dir)
	if err != nil {
		ret.log_file = os.Stdin
	}
	log.SetOutput(ret.log_file)
	ret.server_dir = conf_struct.Root_dir
	return
}

func (server *Server) Server_Main() {
	defer server.listener.Close()

	for {
		conn, err := server.listener.Accept()
		if err != nil {
			log.Println(err)
		} else {
			go server.handle_con(conn)
		}
	}
}

func (server Server) handle_con(conn net.Conn) {
	defer conn.Close()
	conn.SetReadDeadline(time.Now().Add(time.Second * 5))
	conn_reader := bufio.NewReader(conn)
	request_buffer, err := conn_reader.ReadBytes('\n')
	log.Printf("Size: %d Request: %s\n", len(request_buffer), string(request_buffer))
	if err != nil {
		log.Println(err)
		return
	}
	if string(request_buffer) == "\r\n" || len(request_buffer) == 0 {
		server.load_and_write_gophermap(conn, server.server_dir)
	} else {
		file_stat, err := os.Stat(server.server_dir + string(request_buffer))
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				conn.Write([]byte(FILE_NOT_FOUND_ERROR))
			}
		}
		if file_stat.IsDir() {
			server.load_and_write_gophermap(conn, server.server_dir+string(request_buffer))
		}
	}
}

func (svr Server) load_and_write_gophermap(conn net.Conn, path string) {
	file, err := os.Open(path + "/gophermap")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			conn.Write([]byte(NO_GOPHER_MAP_FOR_DIR))
			return
		}
	}
	scanner := bufio.NewScanner(file)
	builder := strings.Builder{}
	builder.Grow(200)
	splitaddr := strings.Split(server.addr, ":")
	for scanner.Scan() {
		line := scanner.Text()
		if line[0] == 'i' {
			builder.WriteString(line + "\r\n")
		} else {
			builder.WriteString(scanner.Text() + "\t" + splitaddr[0] + "\t" + splitaddr[1] + "\r\n")
		}
	}
	built_map := builder.String()
	// log.Println(built_map)
	conn.Write([]byte(built_map))
}

func (svr Server) load_and_send_file(conn net.Conn, path string) {
	file, err := os.Open(path)
	if err != nil {
		log.Println(err)
		conn.Write([]byte(FILE_NOT_FOUND_ERROR))
	}
	scanner := bufio.NewScanner(file)
	buffer := make([]byte, 0, 4096)
	for scanner.Scan() {
		scanner.Bytes()
	}
}
