package main

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"time"
)

type Block struct {
	Index     int    // 数据记录在区块中的位置
	Timestamp string // 自动确定的写入data的时间
	BPM       int    // 自定义数据（这里用来记录每分钟的心跳）
	Hash      string // 此数据记录的SHA256标识符
	PrevHash  string // 上一条记录的SHA256标识符
}

var Blockchain []Block // 区块链

// 计算区块索引、时间戳、数据以及上一个区块hash的hash
func calculateHash(block Block) string {
	record := strconv.Itoa(block.Index) + block.Timestamp + strconv.Itoa(block.BPM) + block.PrevHash
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

// 生成区块
func generateBlock(oldBlock Block, BPM int) (Block, error) {
	var newBlock Block
	t := time.Now()

	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = t.String()
	newBlock.BPM = BPM
	newBlock.PrevHash = oldBlock.Hash
	newBlock.Hash = calculateHash(newBlock)

	return newBlock, nil
}

// 区块验证
func isBlockValid(newBlock, oldBlock Block) bool {
	// 检查 Index 确保它们按预期加1
	if oldBlock.Index+1 != newBlock.Index {
		return false
	}

	// 检查 PrevHash 与前一个块的 Hash 相同
	if oldBlock.Hash != newBlock.PrevHash {
		return false
	}

	// 在当前块上计算 Hash 来确认哈希值是否一致
	if calculateHash(newBlock) != newBlock.Hash {
		return false
	}

	return true
}

// 选择较长的链来作为正确的区块链
func replaceChain(newBlocks []Block) {
	mutex.Lock()
	if len(newBlocks) > len(Blockchain) {
		Blockchain = newBlocks
	}
	mutex.Unlock()
}
