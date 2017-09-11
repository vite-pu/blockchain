package core

import (
	"fmt"
	"reflect"
	"time"

	"github.com/izqui/helpers"
)

type TransactionsQueue chan *Transaction
type BlocksQueue chan Block

type Blockchain struct { // 区块链结构
	CurrentBlock Block
	BlockSlice

	TransactionsQueue
	BlocksQueue
}

func SetupBlockchan() *Blockchain {
	// 配置区块链环境
	bl := new(Blockchain)
	bl.TransactionsQueue, bl.BlocksQueue = make(TransactionsQueue), make(BlocksQueue)

	// Read blcokchain from file and stuff...

	bl.CurrentBlock = bl.CreateNewBlock()

	return bl
}

func (bl *Blockchain) CreateNewBlock() Block {
	// 创建一个新区块，先得到当前的顶端区块，哈希值作为新区块的父哈希字段
	PrevBlock := bl.BlockSlice.PreviousBlock()
	PrevBlockHash := []byte{}
	if PrevBlock != nil {

		PrevBlockHash = PrevBlock.Hash()
	}

	b := NewBlock(previousBlock)
	b.BlockHeader.Origin = Core.Keypair.Public

	return b
}

func (bl *Blockchain) AddBlock(b Block) {
	// 增加一个区块
	bl.BlockSlice = append(bl.BlockSlice, b)
}

func (bl *Blockchain) Run() {
	//
	interruptBlcokGen := bl.GenerateBlocks()
	for {
		select {
		case tr := <-bl.TransactionsQueue:

			if bl.CurrentBlock.TransactionSlice.Exists(*tr) {
				continue
			}
			if !tr.VerifyTransaction(TRANSACTION_POW) {
				fmt.Println("Recieved non valid transaction", tr)
				continue
			}

			bl.CurrentBlock.AddTransaction(tr)
			interruptBlcokGen <- bl.CurrentBlock

			// 广播交易到网络中
			mes := NewMessage(MESSAGE_SEND_TRANSACTION)
			mes.Data, _ = tr.MarshalBinary()

			time.Sleep(300 * time.Millisecond)
			Core.Network.BroadcastQueue <- *mes

		case b := <-bl.BlocksQueue:

			if bl.BlockSlice.Exists(b) {
				fmt.Println("block exists")
				continue
			}

			if !b.VerifyBlock(BLCOK_POW) {
				fmt.Println("block verification fails")
				continue
			}

			if reflect.DeepEqual(b.PrevBlock, bl.CurrentBlock.Hash()) {
				// I'm missing some blocks in the middle.Request'em.
				fmt.Println("Missing blocks in between")
			} else {

				fmt.Println("New block!", b.Hash())

				transDiff := TransactionSlice{}

				if !reflect.DeepEqual(b.BlockHeader.MerkleRoot, bl.CurrentBlock.MerkleRoot) {
					// Transactions are different
					fmt.Println("Transactions are different. finding diff")
					transDiff = DiffTransactionSlice(*bl.CurrentBlock.TransactionSlice, b.TransactionSlice)

				}

				bl.AddBlock(b)

				// Broadcast block and shit
				mes := NewMessage(MESSAGE_SEND_BLOCK)
				mes.Data, _ = b.MarshalBinary()
				Core.Network.BroadcastQueue <- *mes

				// New Block
				bl.CurrentBlock = bl.CreateNewBlock()
				bl.CurrentBlock.TransactionSlice = &transDiff

				interruptBlcokGen <- bl.CurrentBlock
			}
		}
	}

}

func DiffTransactionSlice(a, b TransactionSlice) (diff TransactionSlice) {
	// Assumes transaction array are sorted (which maybe is too big of an assumption)
	lastj := 0
	for _, t := range a {
		found := false
		for j := lastj; j < len(b); j++ {
			if reflect.DeepEqual(b[j].Signature, t.Signature) {
				found = true
				lastj = j
				break
			}
		}
		if !found {
			diff = append(diff, t)
		}
	}

	return
}

func (bl *Blockchain) GenerateBlocks() chan Block {

	interrupt := make(chan Block)

	go func() {

		block := <-interrupt
	loop:
		fmt.Println("Starting Proof of Work...")
		block.BlockHeader.MerkleRoot = block.GenerateMerkleRoot()
		block.BlockHeader.Nonce = 0
		block.BlockHeader.Timestamp = uint32(time.Now().Unix())

		for true { // 挖矿过程

			sleepTime := time.Nanosecond
			if block.TransactionSlice.Len() > 0 {

				if CheckProofOfWork(BLCOK_POW, block.Hash()) {

					block.Signature = block.Sign(Core.Keypair)
					bl.BlocksQueue <- block
					sleepTime = time.Hour * 24
					fmt.Println("Found Block!")

				} else {

					block.BlockHeader.Nonce += 1
				}

			} else {
				sleepTime = time.Hour * 24
				fmt.Println("No trans sleep")
			}

			select {
			case block = <-interrupt:
				goto loop
			case <-helpers.Timeout(sleepTime):
				continue
			}
		}
	}()

	return interrupt
}
