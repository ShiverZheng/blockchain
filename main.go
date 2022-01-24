package main

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/joho/godotenv"
)

// var 通道变量 chan 通道类型
// 通道类型：通道内的数据类型
// 通道变量：保存通道的变量
var bcServer chan []Block

var mutex = &sync.Mutex{}

func handleConn(conn net.Conn) {
	defer conn.Close()

	io.WriteString(conn, "Enter a new BPM:")

	scanner := bufio.NewScanner(conn)

	// 从stdin中获取BPM，在验证通过后将其加入区块
	go func() {
		for scanner.Scan() {
			bpm, err := strconv.Atoi(scanner.Text())
			if err != nil {
				log.Printf("%v not a number: %v", scanner.Text(), err)
				continue
			}

			newBlock, err := generateBlock(Blockchain[len(Blockchain)-1], bpm)
			if err != nil {
				log.Println(err)
				continue
			}

			if isBlockValid(newBlock, Blockchain[len(Blockchain)-1]) {
				newBlockchain := append(Blockchain, newBlock)
				replaceChain(newBlockchain)
			}

			// 将新建的区块链投入通道中
			bcServer <- Blockchain
			io.WriteString(conn, "\n Enter a new BPM:")
		}
	}()

	// 每隔30s向连接的控制台中写入新的区块链
	go func() {
		for {
			time.Sleep(30 * time.Second)
			mutex.Lock()
			// 将新的区块链转成 JSON 便于阅读
			output, err := json.Marshal(Blockchain)
			if err != nil {
				log.Fatal(err)
			}
			mutex.Unlock()
			// 广播
			io.WriteString(conn, string(output))
		}
	}()

	// 使用for-range不断从channel读取数据
	// 当channel关闭时，for循环会自动退出
	// 无需主动监测channel是否关闭，可以防止读取已经关闭的channel，造成读到数据为通道所存储的数据类型的零值
	for range bcServer {
		spew.Dump(Blockchain)
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	// 初始化无缓冲通道
	bcServer = make(chan []Block)

	// 创世块
	t := time.Now()
	genesisBlock := Block{0, t.String(), 0, "", ""}
	spew.Dump(genesisBlock)
	Blockchain = append(Blockchain, genesisBlock)

	// 启动 TCP 并服务 TCP 服务器
	server, err := net.Listen("tcp", ":"+os.Getenv("ADDR"))
	if err != nil {
		log.Fatal(err)
	}
	defer server.Close()

	for {
		conn, err := server.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go handleConn(conn)
	}
}
