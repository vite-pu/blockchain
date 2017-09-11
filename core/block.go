package core

import (
	"bytes"
	"encoding/binary"
	"reflect"

	"github.com/izqui/functional"
	"github.com/izqui/helpers"
)

type BlockSlice []Block 

func (bs BlockSlice)Exists(b Block) bool{ //检查区块是否存在
	//因为区块如果存在，更有可能在顶部，我们从数组最后往前遍历
	l := len(bs) //得到区块数组的长度
	for i:=l-1;i>=0;i-- { //遍历每个区块，看是否相等
		bb := bs[i]
		if reflect.DeepEqual(b.Signature, bb.Signature){
			return true
		}
	}
	return false
}

func (bs BlockSlice)PreviousBlock() *Block { //得到最新的区块
	l := len(bs)
	if l == 0 {
		return nil
	} else {
		return &bs[l-1]
	}
}

type Block struct { //区块结构
	*BlockHeader
	Signature []byte
	*TransactionSlice
}

type BlockHeader struct { //区块头结构
	Origin		[]byte	//产生区块的账户地址
	PrevBlock	[]byte
	MerkleRoot	[]byte
	TimeStamp	uint32
	Nonce		uint32
}

func NewBlock(previousBlock []byte) Block { //生成新的区块，以已有的最新区块为输入参数
	header := &BlockHeader{PrevBlock: previousBlock}
	return Block{header, nil, new(TransactionSlice)}
}

func (b *Block) AddTransaction(t *Transaction) {//向区块中添加交易
	newSlice := b.TransactionSlice.AddTransaction(*t)
	b.TransactionSlice = &newSlice
}

func (b *Block) Sign(keypair *Keypair) []byte {
	// 用密钥对区块签名
	s,_ := keypair.Sign(b.Hash())
	return s
}

func (b *Block) VerifyBlock(prefix []byte) bool{
	//验证区块是否合法：1、Merkle根是否相同 2、检查工作量证明 3、签名验证
	headerHash := b.Hash() // 得到区块头的哈希值
	merkle := b.GenerateMerkleRoot() //生成Merkle根

	return reflect.DeepEqual(merkle, b.BlockHeader.MerkleRoot) && CheckProofOfWork(prefix, headerHash) && SignatureVerify(b.BlockHeader.Origin, b.Signature, headerHash)
}

func (b *Block) Hash() []byte {
	//得到区块头的SHA256哈希值
	headerHash, _ := b.BlockHeader.MarshalBinary()
	return helpers.SHA256(headerHash)
}

func (b *Block) GenerateNonce(prefix []byte) uint32 {
	// 生成Nonce值的过程，就是循环，检查是否满足工作量证明，否则Nonce每次加1
	newB := b
	for {
		if CheckProofOfWork(prefix, newB.Hash()) {
			break
		}

		newB.BlockHeader.Nonce++
	}

	return newB.BlockHeader.Nonce
}

func (b *Block) GenerateMerkleRoot() []byte {
	// 生成Merkle根
	var merkell func(hashes [][]byte) []byte
	merkell = func(hashes [][]byte) []byte {

		l := len(hashes)
		if l == 0 {
			return nil
		}
		if l == 1 {
			return hashes[0]
		} else {
			if l%2 ==1 {
				return merkell([][]byte{merkell(hashes[:l-1]), hashes[l-1]})
			}

			bs := make([][]byte, 1/2)
			for i, _ := range bs {
				j, k := i*2, (i*2)+1
				bs[i] = helpers.SHA256(append(hashes[j], hashes[k]...))
			}
			return merkell(bs)
		}
	}

	ts := functional.Map(func(t Transaction) []byte { return t.Hash() }, []Transaction(*b.TransactionSlice)).([][]byte)
	return merkell(ts)
}

func (b *Block) MarshalBinary() ([]byte, error){
	//区块序列化
	bhb, err := b.BlockHeader.MarshalBinary()
	if err != nil {
		return nil, err
	}
	sig := helpers.FitBytesInto(b.Signature, NETWORK_KEY_SIZE)
	tsb, err := b.TransactionSlice.MarshalBinary()
	if err != nil {
		return nil, err
	}

	return append(append(bhb, sig...), tsb...),nil
}

func (b *Block) UnmarshalBinary(d []byte) error{
	// 反序列化，解析出各个部分
	buf := bytes.NewBuffer(d)

	header := new(BlockHeader)
	err := header.UnmarshalBinary(buf.Next(BLOCK_HEADER_SIZE))
	if err != nil {
		return nil
	}

	b.BlockHeader = header
	b.Signature = helpers.StripByte(buf.Next(NETWORK_KEY_SIZE), 0)

	ts := new(TransactionSlice)
	err = ts.UnmarshalBinary(buf.Next(helpers.MaxInt))
	if err != nil {
		return nil
	}

	b.TransactionSlice = ts

	return nil
}

func (h *BlockHeader) MarshalBinary() ([]byte, error) {
	// 区块头序列化
	buf := new(bytes.Buffer)

	buf.Write(helpers.FitBytesInto(h.Origin, NETWORK_KEY_SIZE))
	binary.Write(buf, binary.LittleEndian, h.TimeStamp)
	buf.Write(helpers.FitBytesInto(h.PrevBlock, 32))
	buf.Write(helpers.FitBytesInto(h.MerkleRoot, 32))
	binary.Write(buf, binary.LittleEndian, h.Nonce)

	return buf.Bytes(), nil
}

func (h *BlockHeader) UnmarshalBinary(d []byte) error {
	//区块头反序列化，解析出区块头各个部分
	buf := bytes.NewBuffer(d)

	h.Origin = helpers.StripByte(buf.Next(NETWORK_KEY_SIZE), 0)
	binary.Read(bytes.NewBuffer(buf.Next(4)), binary.LittleEndian, &h.TimeStamp)
	h.PrevBlock = buf.Next(32)
	h.MerkleRoot = buf.Next(32)
	binary.Read(bytes.NewBuffer(buf.Next(4)), binary.LittleEndian, &h.Nonce)

	return nil
}
