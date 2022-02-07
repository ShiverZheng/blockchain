package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"time"
)

// 每个块的内容
type Block struct {
	Index     int    // 增量
	Timestamp string // 当前时间的字符串表示
	BPM       int    // 心率
	Hash      string // 当前区块的 Hash
	PrevHash  string // 前一个区块的 Hash
	Validator string // 验证者
}

// 官方区块链：一系列经过验证的区块
var Blockchain []Block

// 在选出正确的区块之前的暂时容器
var tempBlocks []Block

// 区块通道：每个提出新块的节点都将其发送到该通道
var candidateBlocks = make(chan Block)

// 广播通道：TCP服务器向所有节点广播最新区块链
var announcements = make(chan string)

// 锁
var mutex = &sync.Mutex{}

// 节点与已质押的代币数量的映射
var validators = make(map[string]int)

// 接受一个字符串并返回其 SHA256 哈希表示
func calculateHash(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

// 连接其所有字段来散列块的内容
func calculateBlockHash(block Block) string {
	record := string(block.Index) + block.Timestamp + string(block.BPM) + block.PrevHash
	return calculateHash(record)
}

// 生成新的块
func generateBlock(oldBlock Block, BPM int, address string) (Block, error) {

	var newBlock Block

	t := time.Now()

	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = t.String()
	newBlock.BPM = BPM
	newBlock.PrevHash = oldBlock.Hash
	newBlock.Hash = calculateBlockHash(newBlock)
	newBlock.Validator = address

	return newBlock, nil
}

// 检查 Hash 与 PrevHash 以确保链没有损坏
func isBlockValid(newBlock, oldBlock Block) bool {
	if oldBlock.Index+1 != newBlock.Index {
		return false
	}

	if oldBlock.Hash != newBlock.PrevHash {
		return false
	}

	if calculateBlockHash(newBlock) != newBlock.Hash {
		return false
	}

	return true
}

// 1、允许输入代币余额（不会执行任何余额检查，因为没有钱包逻辑）
// 2、接收最新区块链的广播
// 3、接收网络中哪个验证者赢得最新区块的广播
// 4、将自己添加到验证器的整体列表中
// 5、输入区块数据 BPM (每个验证者的脉搏率)
// 6、提出一个新区块
func handleConn(conn net.Conn) {
	defer conn.Close()

	// 接收并打印出来自 TCP 服务器的任何通知
	// 当一个被选中时，这些公告将是获胜的验证者
	go func() {
		for {
			msg := <-announcements
			io.WriteString(conn, msg)
		}
	}()

	// 验证器地址
	var address string

	// 允许用户分配要质押的代币数量
	// 代币数量越多，生成新区块的机会就越大
	// 为验证者分配一个 SHA256 地址，并将该地址与新验证者的质押代币数量添加到 validators 全局映射中
	io.WriteString(conn, "Enter token balance:")
	scanBalance := bufio.NewScanner(conn)
	for scanBalance.Scan() {
		balance, err := strconv.Atoi(scanBalance.Text())
		if err != nil {
			log.Printf("%v not a number: %v", scanBalance.Text(), err)
			return
		}
		t := time.Now()
		address = calculateHash(t.String())
		validators[address] = balance
		fmt.Println(validators)
		break
	}

	// 输入 BPM 验证者的脉搏频率
	io.WriteString(conn, "\nEnter a new BPM:")

	scanBPM := bufio.NewScanner(conn)

	// 创建一个单独的线程来处理块逻辑
	go func() {
		for {
			// 从标准输入中获取 BPM 并在进行必要的验证后将其添加到区块链中
			for scanBPM.Scan() {
				bpm, err := strconv.Atoi(scanBPM.Text())

				// 如果验证者试图提出一个被污染的块
				// 例如在这里输入一个不是整数的 BPM ，会引发错误
				// 这将会立即从验证者列表中删除该验证者
				// 他们不再有资格锻造新的区块并且他们失去了质押代币

				// 这种失去代币余额的可能性是确保【权益证明】通常是安全的主要原因
				// 如果有人试图为了自己的利益而改变区块链并且被抓住了
				// 这个人就会失去他所有的质押代币
				// 这是对不良参与者的主要威慑
				if err != nil {
					log.Printf("%v not a number: %v", scanBPM.Text(), err)
					delete(validators, address)
					conn.Close()
				}

				mutex.Lock()
				oldLastBlock := Blockchain[len(Blockchain)-1]
				mutex.Unlock()

				// 创建一个新块，并将其发送到 candidateBlocks 通道进行进一步处理
				newBlock, err := generateBlock(oldLastBlock, bpm, address)
				if err != nil {
					log.Println(err)
					continue
				}
				if isBlockValid(newBlock, oldLastBlock) {
					candidateBlocks <- newBlock
				}
				io.WriteString(conn, "\nEnter a new BPM:")
			}
		}
	}()

	// 模拟接收广播
	// 定期打印最新的区块链，因此每个验证者都知道最新状态
	for {
		time.Sleep(time.Minute)
		mutex.Lock()
		output, err := json.Marshal(Blockchain)
		mutex.Unlock()
		if err != nil {
			log.Fatal(err)
		}
		io.WriteString(conn, string(output)+"\n")
	}
}

// 每30秒选出一个获胜者，让每个验证者有时间提出一个新的区块
func pickWinner() {
	time.Sleep(30 * time.Second)
	mutex.Lock()
	temp := tempBlocks
	mutex.Unlock()

	// 创建一个 lotterPool 包含可以被选为获胜者的验证者的地址
	lotteryPool := []string{}
	// 在继续执行逻辑之前，检查是否在提议块的临时存储罐中存有一些提议的块
	if len(temp) > 0 {
		// 在 OUTER for 循环中
		// 检查在临时切片中没有相同的验证器
		// 否则跳过该块并寻找下一个唯一的验证者
	OUTER:
		for _, block := range temp {
			for _, node := range lotteryPool {
				if block.Validator == node {
					continue OUTER
				}
			}

			mutex.Lock()
			setValidators := validators
			mutex.Unlock()

			// 确保从 temp 中的块数据中获得的验证器实际上是位于验证器映射中的合格验证器
			// 如果它们存在，那么我们将它们添加到 lotteryPool 中
			k, ok := setValidators[block.Validator]
			if ok {
				for i := 0; i < k; i++ {
					lotteryPool = append(lotteryPool, block.Validator)
				}
			}
		}

		// 如何根据验证者质押的代币数量为他们分配适当的权重？
		// 用验证者地址的副本填充我们的彩票池
		// 他们会因他们质押的每个代币获得一份副本
		// 因此，投入 100 个代币的验证者将在彩票池中获得 100 个条目
		// 仅放入 1 个令牌的验证者将仅获得 1 个条目
		s := rand.NewSource(time.Now().Unix())
		r := rand.New(s)
		// 从 lotteryPool 中随机挑选获胜者并将他们的地址分配给 lotteryWinner
		// Intn 以 int 形式返回 [0,n) 中的非负伪随机数。 如果 n <= 0，它会panic。
		lotteryWinner := lotteryPool[r.Intn(len(lotteryPool))]

		for _, block := range temp {
			// 将他们的区块添加到我们的区块链中
			if block.Validator == lotteryWinner {
				mutex.Lock()
				Blockchain = append(Blockchain, block)
				mutex.Unlock()

				// 向其他节点宣布获胜者
				for range validators {
					announcements <- "\nwinning validator: " + lotteryWinner + "\n"
				}
				break
			}
		}

	}

	// 清理 tempBlocks 存储，以便可以再次用下一组提议的区块填充它。
	mutex.Lock()
	tempBlocks = []Block{}
	mutex.Unlock()
}
