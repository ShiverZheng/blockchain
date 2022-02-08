package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"sync"
	"time"

	golog "github.com/ipfs/go-log"
	peer "github.com/libp2p/go-libp2p-core/peer"
	pstore "github.com/libp2p/go-libp2p-core/peerstore"
	multiAddr "github.com/multiformats/go-multiaddr"
)

var Blockchain []Block

var mutex = &sync.Mutex{}

func main() {
	// 创建创世块
	t := time.Now()
	genesisBlock := Block{}
	genesisBlock = Block{0, t.String(), 0, calculateHash(genesisBlock), ""}
	Blockchain = append(Blockchain, genesisBlock)

	// 使用 golog 库的记录器来处理日志记录，这是可选的
	golog.SetAllLoggers(golog.LevelInfo)

	/**** 设置所有的命令行标志 Start ****/
	// 打开希望允许连接的端口，这意味着我们正在充当主机
	listenF := flag.Int("l", 0, "wait for incoming connections")

	// 让我们指定要连接的另一台主机的地址，这意味着如果我们使用此标志，我们将充当主机的对等方
	target := flag.String("d", "", "target peer to dial")

	// 安全流，确保在运行程序时始终通过设置标志来使用它
	secio := flag.Bool("secio", false, "enable secio")

	// 可选的随机种子，用于构建地址，其他对等点可以用来连接到我们
	seed := flag.Int64("seed", 0, "set random seed for id generation")

	flag.Parse()

	if *listenF == 0 {
		log.Fatal("Please provide a port to bind on with -l")
	}
	/**** 设置所有的命令行标志 End ****/

	// 创建新的主机
	host, err := makeBasicHost(*listenF, *secio, *seed)
	if err != nil {
		log.Fatal(err)
	}

	// 如果我们只是作为一个主机（即我们没有连接到其他主机）
	if *target == "" {
		log.Println("listening for connections")
		// 在主机 A 上设置流处理程序。/p2p/1.0.0 是用户定义的协议名称。
		host.SetStreamHandler("/p2p/1.0.0", handleStream)

		// 使程序一直运行
		select {}
		/**** 监听器代码结束 ****/
	} else { // 如果想要连接到另一台主机
		// 再次设置处理程序，以便充当主机以及连接对等方
		host.SetStreamHandler("/p2p/1.0.0", handleStream)

		// 解构 target 字符串，以便找到想要连接的主机
		ipfsAddr, err := multiAddr.NewMultiaddr(*target)
		if err != nil {
			log.Fatalln(err)
		}

		pid, err := ipfsAddr.ValueForProtocol(multiAddr.P_IPFS)
		if err != nil {
			log.Fatalln(err)
		}

		// 获取对等方 ID
		peerid, err := peer.Decode(pid)
		if err != nil {
			log.Fatalln(err)
		}

		// 获取目标地址
		targetPeerAddr, _ := multiAddr.NewMultiaddr(
			fmt.Sprintf("/ipfs/%s", peer.Encode(peerid)))
		targetAddr := ipfsAddr.Decapsulate(targetPeerAddr)

		// 将对等方 ID 与目标地址记录到“存储”中，这样​​就可以跟踪连接的对象
		host.Peerstore().AddAddr(peerid, targetAddr, pstore.PermanentAddrTTL)

		log.Println("opening stream")

		// 与想要连接的对等方创建连接流
		s, err := host.NewStream(context.Background(), peerid, "/p2p/1.0.0")
		if err != nil {
			log.Fatalln(err)
		}

		// 为了能够接收和发送他们的数据流（我们的区块链）
		// 为 writeData 和 readData 创建单独的线程
		rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
		go writeData(rw)
		go readData(rw)

		// 通过一个空语句阻塞来结束，所以程序不会完成后退出
		select {}
	}
}
