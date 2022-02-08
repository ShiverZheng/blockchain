package main

import (
	"bufio"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	mathRand "math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/libp2p/go-libp2p"
	crypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	net "github.com/libp2p/go-libp2p-core/network"
	multiAddr "github.com/multiformats/go-multiaddr"
)

// makeBasicHost 创建一个具有随机对等 ID 的 LibP2P 主机，监听给定的多个地址，secio 为 true 的情况启用
// listenPort 命令行标志中指定的其他对等方连接的端口
// secio 打开和关闭安全数据流的布尔值。它代表“安全输入/输出”，通常启用它是更好的选择
// randseed 是一个可选的命令行标志，它允许提供一个种子来为主机创建一个随机地址
func makeBasicHost(listenPort int, secio bool, randseed int64) (host.Host, error) {
	var r io.Reader

	// 为主机生成密钥
	// 如果 seed 为 0，使用真实的随机密码
	// 否则，使用确定的随机源使得多次运行中生成的 key 不变
	if randseed == 0 {
		r = rand.Reader
	} else {
		r = mathRand.New(mathRand.NewSource(randseed))
	}

	// 为该主机生成密钥对，将使用它获取有效的主机 ID
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		return nil, err
	}

	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", listenPort)),
		libp2p.Identity(priv),
	}

	// 基于指定配置创建一个 host
	basicHost, err := libp2p.New(opts...)
	if err != nil {
		return nil, err
	}

	// 返回主机地址 /p2p/<BASIC_HOST_ID>
	hostAddr, _ := multiAddr.NewMultiaddr(fmt.Sprintf("/ipfs/%s", basicHost.ID().Pretty()))

	// [/ip4/127.0.0.1/tcp/<l>]
	addrs := basicHost.Addrs()

	var addr multiAddr.Multiaddr

	// 找到 ip4 开头的地址
	for _, i := range addrs {
		if strings.HasPrefix(i.String(), "/ip4") {
			addr = i
			break
		}
	}

	// 完整地址 /ip4/127.0.0.1/tcp/<l>/p2p/<BASIC_HOST_ID>
	fullAddr := addr.Encapsulate(hostAddr)
	log.Printf("I am %s\n", fullAddr)

	if secio {
		log.Printf("Now run \"go run . -l %d -d %s -secio\" on a different terminal\n", listenPort+1, fullAddr)
	} else {
		log.Printf("Now run \"go run . -l %d -d %s\" on a different terminal\n", listenPort+1, fullAddr)
	}

	return basicHost, nil
}

// 需要让主机处理传入的数据流
// 当另一个节点连接到主机并想要提出一个新的区块链来覆盖主机中的区块链时
// 需要确定是否应该接受它
func handleStream(s net.Stream) {
	log.Println("Got a new stream!")

	// 创建一个非阻塞的读写缓冲流
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	go readData(rw)
	go writeData(rw)

	// stream 's' 将保持打开状态，直到关闭它（或另一边关闭它）
}

func readData(rw *bufio.ReadWriter) {
	// 读取函数需要对传入的区块链保持开放，采用无限循环
	for {
		// 从对等点解析传入的区块链字符串
		// ReadString 读取直到输入中第一次出现分隔符，返回一个字符串，其中包括分隔符的数据。
		str, err := rw.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		if str == "" {
			return
		}

		if str != "\n" {
			chain := make([]Block, 0)
			if err := json.Unmarshal([]byte(str), &chain); err != nil {
				log.Fatal(err)
			}

			mutex.Lock()
			// 检查区块链长度，将更长的链作为最新的区块链
			// 可修改为其他校验方式
			if len(chain) > len(Blockchain) {
				Blockchain = chain
				bytes, err := json.MarshalIndent(Blockchain, "", " ")
				if err != nil {
					log.Fatal(err)
				}
				// 绿色控制台颜色: 	\x1b[32m
				// 重置控制台颜色: 	\x1b[0m
				fmt.Printf("\x1b[32m%s\x1b[0m> ", string(bytes))
			}
			mutex.Unlock()
		}
	}
}

func writeData(rw *bufio.ReadWriter) {
	// 每隔5秒向对等方广播区块链的最新状态
	// 如果区块链长度比对等方的短，对等方收到后将会丢弃，否则将会接受
	// 所有对等点都在不断通过网络的最新状态更新区块链
	go func() {
		for {
			time.Sleep(5 * time.Second)
			mutex.Lock()
			bytes, err := json.Marshal(Blockchain)
			if err != nil {
				log.Println(err)
			}
			mutex.Unlock()

			mutex.Lock()
			rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
			rw.Flush()
			mutex.Unlock()
		}
	}()

	// 创建新的阅读器
	stdReader := bufio.NewReader(os.Stdin)

	// bufio.NewReader以便它可以阅读stdin（控制台输入）
	// 因为希望能够不断添加新块，所以将其置于无限循环中
	for {
		fmt.Print("> ")
		sendData, err := stdReader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		sendData = strings.Replace(sendData, "\n", "", -1)
		// 确保输入的 BPM 是一个整数
		bpm, err := strconv.Atoi(sendData)
		if err != nil {
			log.Fatal(err)
		}

		// 创建新块并校验，如果正确就可以添加
		newBlock := generateBlock(Blockchain[len(Blockchain)-1], bpm)

		if isBlockValid(newBlock, Blockchain[len(Blockchain)-1]) {
			mutex.Lock()
			Blockchain = append(Blockchain, newBlock)
			mutex.Unlock()
		}

		// 格式化区块链并打印到控制台
		bytes, err := json.Marshal(Blockchain)
		if err != nil {
			log.Println(err)
		}
		spew.Dump(Blockchain)

		// 广播到连接着的对等网络中
		mutex.Lock()
		rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
		rw.Flush()
		mutex.Unlock()
	}
}
