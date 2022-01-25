package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// 散列0的数量，所需的0越多，越难找到正确的Hash值
const difficulty = 1

type Block struct {
	Index      int    // 增量
	Timestamp  string // 当前时间的字符串表示
	BPM        int    // 心率
	Hash       string // 当前区块的Hash
	PrevHash   string // 前一个区块的Hash
	Difficulty int    // 难度系数，与Hash前缀0的数量相等
	Nonce      string // 从零开始的自增16进制
}

var Blockchain []Block

type Message struct {
	BPM int
}

func isBlockValid(newBlock, oldBlock Block) bool {
	if oldBlock.Index+1 != newBlock.Index {
		return false
	}

	if oldBlock.Hash != newBlock.PrevHash {
		return false
	}

	if calculateHash(newBlock) != newBlock.Hash {
		return false
	}

	return true
}

// 将Nonce加入Hash计算
func calculateHash(block Block) string {
	record := strconv.Itoa(block.Index) + block.Timestamp + strconv.Itoa(block.BPM) + block.PrevHash + block.Nonce
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

// 验证Hash的前缀零是否与难度数量相等
func isHashValid(hash string, difficulty int) bool {
	prefix := strings.Repeat("0", difficulty)
	return strings.HasPrefix(hash, prefix)
}

// 生成返回符合工作量证明的区块
func generateBlock(oldBlock Block, BPM int) Block {
	var newBlock Block

	t := time.Now()

	// 初始化区块
	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = t.String()
	newBlock.BPM = BPM
	newBlock.PrevHash = oldBlock.Hash
	newBlock.Difficulty = difficulty

	for i := 0; ; i++ {
		// 将i转换成16进制，并赋值给Nonce
		hex := fmt.Sprintf("%x", i)
		fmt.Println(hex)
		// 通过i计算Hash，Nonce从0开始，检查Hash结果是否与定义的difficulty相同
		newBlock.Nonce = hex
		if !isHashValid(calculateHash(newBlock), newBlock.Difficulty) {
			fmt.Println(calculateHash(newBlock), " do more work!")
			// 模拟需要一些时间来解决工作量证明
			time.Sleep(time.Second)
			// 一直循环，直到得到想要的前导零的数量，表示完成了工作证明
			continue
		} else {
			hash := calculateHash(newBlock)
			fmt.Println(hash, " Work Done!")
			newBlock.Hash = hash
			break
		}
	}

	return newBlock
}
