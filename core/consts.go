package core

import _ "fmt"

const (
	BLOCKCHAIN_PORT			= "9119"
	MAX_NODE_CONNECTIONS	= 400

	NETWORK_KEY_SIZE		= 80

	TRANSACTION_KEY_SIZE	= NETWORK_KEY_SIZE /* from key */ + NETWORK_KEY_SIZE /* to key */ + 4 /* int32 timestamp */ +32 /* sha256 payload hash */ + 4 /* int32 payload length */ + 4 /* int32 nonce */
	BLOCK_HEADER_SIZE		= NETWORK_KEY_SIZE /* origin key */ + 4 /* int32 timestamp */ + 32 /* prev block hash */ + 32 /* merkle tree hash */ + 4 									 /* int32 nonce */

	KEY_POW_COMPLEXITY		= 0

	TRANSACTION_POW_COMPLEXITY = 1

	BLCOK_POW_COMPLEXITY 	= 2

	KEY_SIZE = 28

	POW_PREFIX	= 0

	MESSAGE_TYPE_SIZE	= 1
	MESSAGE_OPTIONS_SIZE = 4
)

const (
	MESSAGE_GET_NODES	= iota + 20
	MESSAGE_SEND_NODES

	MESSAGE_GET_TRANSACTION
	MESSAGE_SEND_TRANSACTION

	MESSAGE_GET_BLOCK
	MESSAGE_SEND_BLOCK
)

func SEED_NODES() []string {
	nodes := []string{"219.216.65.127"}

	return nodes
}