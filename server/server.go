package server

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
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
	search_dir string
}

func New_Instance() (ret Server) {
	conf_struct := struct {
		Ip           string `json:"ip"`
		Port         string `json:"port"`
		Root_dir     string `json:"root_dir"`
		Log_File_dir string `json:"log_file"`
		Search_Dir   string `json:"search_dir"`
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
	ret.search_dir = conf_struct.Search_Dir
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
	} else if strings.Contains(string(request_buffer), server.search_dir) {
		if len(request_buffer) > len(server.search_dir)+2 {
			server.handle_search(conn, string(request_buffer))
		}
	} else {
		path := server.server_dir + string(request_buffer)
		path = strings.Trim(path, "\n\r")
		file_stat, err := os.Stat(path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				conn.Write([]byte(FILE_NOT_FOUND_ERROR))
				return
			}
		}
		if file_stat.IsDir() {
			server.load_and_write_gophermap(conn, path)
		} else {
			server.load_and_send_file(conn, path)
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
	defer file.Close()
	scanner := bufio.NewScanner(file)
	builder := strings.Builder{}
	splitaddr := strings.Split(svr.addr, ":")
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 0 {
			if line[0] == 'i' {
				builder.WriteString(line + "\t" + "\" + \t" + splitaddr[0] + "\t" + splitaddr[1] + "\r\n")
			} else {
				builder.WriteString(line + "\t" + splitaddr[0] + "\t" + splitaddr[1] + "\r\n")
			}
		}
	}
	builder.WriteString(".\r\n")
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
	file.Seek(0, io.SeekStart)
	defer file.Close()
	scanner := bufio.NewReader(file)
	buffer, err := io.ReadAll(scanner)
	if err != nil {
		log.Println(err)
		return
	}
	conn.Write(buffer)
}

func (svr Server) handle_search(conn net.Conn, request string) {
	split_request := strings.Split(request, "\t")

	search_prg := exec.Command(svr.server_dir + svr.search_dir)
	stdin_pipe, err := search_prg.StdinPipe()
	if err != nil {
		log.Println(err)
		return
	}
	adress_and_port := strings.Split(svr.addr, ":")
	go func() {
		defer stdin_pipe.Close()
		stdin_pipe.Write([]byte(adress_and_port[0] + "\t" + adress_and_port[1] + "\t" + svr.server_dir + "\t" + split_request[1]))
	}()
	search_result, err := search_prg.CombinedOutput()
	log.Println(err)
	log.Println(string(search_result))
	conn.Write(search_result)
}
