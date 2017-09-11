package core

import (
	"bytes"
	"encoding/binary"
	"errors"
	"reflect"
	"time"

	"github.com/izqui/helpers"
)

type Transaction struct { // 定义交易结构
	Header    TransactionHeader
	Signature []byte
	Payload   []byte
}

type TransactionHeader struct { // 交易头结构
	From          []byte
	To            []byte
	Timestamp     uint32
	PayloadHash   []byte
	PayloadLength uint32
	Nonce         uint32
}

func NewTransaction(from, to, payload []byte) *Transaction {
	// 新生成一条交易
	t := Transaction{Header: TransactionHeader{From: from, To: to}, Payload: payload}

	t.Header.Timestamp = uint32(time.Now().Unix())
	t.Header.PayloadHash = helpers.SHA256(t.Payload)
	t.Header.PayloadLength = uint32(len(t.Payload))

	return &t
}

func (t *Transaction) Hash() []byte {
	// 得到交易哈希值
	headerBytes, _ := t.Header.MarshalBinary()
	return helpers.SHA256(headerBytes)
}

func (t *Transaction) Sign(keypair *Keypair) []byte {
	// 对交易进行签名
	s, _ := keypair.Sign(t.Hash())

	return s
}

func (t *Transaction) VerifyTransaction(pow []byte) bool {
	// 对交易进行验证
	headerHash := t.Hash()
	payloadHash := helpers.SHA256(t.Payload)

	return reflect.DeepEqual(payloadHash, t.Header.PayloadHash) && CheckProofOfWork(pow, headerHash) && SignatureVerify(t.Header.From, t.Signature, headerHash)

}

func (t *Transaction) GenerateNonce(prefix []byte) uint23 {
	//  生成Nonce值
	newT := t

	for {
		if CheckProofOfWork(prefix, newT.Hash()) {
			break
		}

		newt.Header.Nonce++
	}

	return newT.Header.Nonce
}

func (t *Transaction) MarshalBinary() ([]byte, error) {
	// 交易序列化

	headerBytes := t.Header.MarshalBinary()

	if len(headerBytes) != TRANSACTION_HEADER_SIZE {
		return nil, errors.New("Header marsahlling error")
	}

	return append(append(headerBytes, helpers.FitBytesInto(t.Signature, NETWORK_KEY_SIZE)...), t.Payload...), nil
}

func (t *Transaction) UnmarshalBinary(d []byte) ([]byte, error) {
	// 交易反序列化

	buf := bytes.NewBuffer(d)

	if len(d) < TRANSACTION_HEADER_SIZE+NETWORK_KEY_SIZE {
		return nil, errors.New("Insuficient bytes for unmarshalling transaction")
	}

	header := &TransactionHeader{}
	if err := header.UnmarshalBinary(buf.Next(TRANSACTION_HEADER_SIZE)); err != nil {
		return nil, err
	}

	t.Header = *header

	t.Signature = helpers.StripByte(buf.Next(NETWORK_KEY_SIZE), 0)
	t.Payload = buf.Next(int(t.Header.PayloadLength))

	return buf.Next(helpers.MaxInt), nil
}

func (th *TransactionHeader) MarshalBinary() ([]byte, error) {
	// 交易头序列化
	buf := new(bytes.Buffer)

	buf.Write(helpers.FitBytesInto(th.From, NETWORK_KEY_SIZE))
	buf.Write(helpers.FitBytesInto(th.To, NETWORK_KEY_SIZE))
	binary.Write(buf, binary.LittleEndian, th.Timestamp)
	buf.Write(helpers.FitBytesInto(th.PayloadHash, 32))
	binary.Write(buf, binary.LittleEndian, th.PayloadLength)
	binary.Write(buf, binary.LittleEndian, th.Nonce)

	return buf.Bytes(), nil
}

func (th *TransactionHeader) UnmarshalBinary(d []byte) error {
	// 交易头反序列化
	buf := bytes.NewBuffer(d)

	th.From = helpers.StripByte(buf.Next(NETWORK_KEY_SIZE), 0)
	th.To = helpers.StripByte(buf.Next(NETWORK_KEY_SIZE), 0)
	binary.Read(bytes.NewBuffer(buf.Next(4)), binary.LittleEndian, &th.Timestamp)
	th.PayloadHash = buf.Next(32)
	binary.Read(bytes.NewBuffer(buf.Next(4)), binary.LittleEndian, &th.PayloadLength)
	binary.Read(bytes.NewBuffer(buf.Next(4)), binary.LittleEndian, &th.Nonce)

	return nil
}

type TransactionSlice []Transaction

func (slice TransactionSlice) Len() int {
	return len(slice)
}

func (slice TransactionSlice) Exists(tr Transaction) bool {
	for _, t := range slice {
		if reflect.DeepEqual(t.Signature, tr.Signature) {
			return true
		}
	}

	return false
}

func (slice TransactionSlice) AddTransaction(t Transaction) TransactionSlice {
	// 添加交易, 按时间戳有序插入

	for i, tr := range slice {
		if tr.Header.Timestamp >= t.Header.Timestamp {
			return append(append(slice[:i], t), slice[i:]...) // ...的作用，说明是不定参数，相当于把slice[i:]打散后加入到前面的切片中的
		}
	}

	return append(slice, t)
}

func (slice *TransactionSlice) MarshalBinary() ([]byte, error) {
	// 序列化

	buf := new(bytes.Buffer)

	for _, t := range *slice {
		bs, err := t.MarshalBinary()

		if err != nil {
			return nil, err
		}

		buf.Write(bs)
	}

	return buf.Bytes(), nil
}

func (slice *TransactionSlice) UnmarshalBinary(d []byte) error {
	remaining := d

	for len(remaining) > TRANSACTION_HEADER_SIZE+NETWORK_KEY_SIZE {
		t := new(Transaction)
		rem, err := t.UnmarshalBinary(remaining)

		if err != nil {
			return nil
		}

		(*slice) = append((*slice), *t)
		remaining = rem
	}
	return nil
}
