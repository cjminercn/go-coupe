// Copyright 2015 The go-coupe Authors
// This file is part of the go-coupe library.
//
// The go-coupe library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-coupe library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-coupe library. If not, see <http://www.gnu.org/licenses/>.

package api

import (
	"strings"

	"fmt"

	"github.com/cjminercn/go-coupe/eth"
	"github.com/cjminercn/go-coupe/rpc/codec"
	"github.com/cjminercn/go-coupe/rpc/shared"
	"github.com/cjminercn/go-coupe/xeth"
)

var (
	// Mapping between the different methods each api supports
	AutoCompletion = map[string][]string{
		"admin": []string{
			"addPeer",
			"datadir",
			"enableUserAgent",
			"exportChain",
			"getContractInfo",
			"httpGet",
			"importChain",
			"nodeInfo",
			"peers",
			"register",
			"registerUrl",
			"saveInfo",
			"setGlobalRegistrar",
			"setHashReg",
			"setUrlHint",
			"setSolc",
			"sleep",
			"sleepBlocks",
			"startNatSpec",
			"startRPC",
			"stopNatSpec",
			"stopRPC",
			"verbosity",
		},
		"db": []string{
			"getString",
			"putString",
			"getHex",
			"putHex",
		},
		"debug": []string{
			"dumpBlock",
			"getBlockRlp",
			"metrics",
			"printBlock",
			"processBlock",
			"seedHash",
			"setHead",
		},
		"eth": []string{
			"accounts",
			"blockNumber",
			"call",
			"contract",
			"coinbase",
			"compile.lll",
			"compile.serpent",
			"compile.solidity",
			"contract",
			"defaultAccount",
			"defaultBlock",
			"estimateGas",
			"filter",
			"getBalance",
			"getBlock",
			"getBlockTransactionCount",
			"getBlockUncleCount",
			"getCode",
			"getNatSpec",
			"getCompilers",
			"gasPrice",
			"getStorageAt",
			"getTransaction",
			"getTransactionCount",
			"getTransactionFromBlock",
			"getTransactionReceipt",
			"getUncle",
			"hashrate",
			"mining",
			"namereg",
			"pendingTransactions",
			"resend",
			"sendRawTransaction",
			"sendTransaction",
			"sign",
			"syncing",
		},
		"miner": []string{
			"hashrate",
			"makeDAG",
			"setEtherbase",
			"setExtra",
			"setGasPrice",
			"startAutoDAG",
			"start",
			"stopAutoDAG",
			"stop",
		},
		"net": []string{
			"peerCount",
			"listening",
		},
		"personal": []string{
			"listAccounts",
			"newAccount",
			"unlockAccount",
		},
		"shh": []string{
			"post",
			"newIdentity",
			"hasIdentity",
			"newGroup",
			"addToGroup",
			"filter",
		},
		"txpool": []string{
			"status",
		},
		"web3": []string{
			"sha3",
			"version",
			"fromWei",
			"toWei",
			"toHex",
			"toAscii",
			"fromAscii",
			"toBigNumber",
			"isAddress",
		},
	}
)

// Parse a comma separated API string to individual api's
func ParseApiString(apistr string, codec codec.Codec, xeth *xeth.XEth, eth *eth.cjminercn) ([]shared.cjminercnApi, error) {
	if len(strings.TrimSpace(apistr)) == 0 {
		return nil, fmt.Errorf("Empty apistr provided")
	}

	names := strings.Split(apistr, ",")
	apis := make([]shared.cjminercnApi, len(names))

	for i, name := range names {
		switch strings.ToLower(strings.TrimSpace(name)) {
		case shared.AdminApiName:
			apis[i] = NewAdminApi(xeth, eth, codec)
		case shared.DebugApiName:
			apis[i] = NewDebugApi(xeth, eth, codec)
		case shared.DbApiName:
			apis[i] = NewDbApi(xeth, eth, codec)
		case shared.EthApiName:
			apis[i] = NewEthApi(xeth, eth, codec)
		case shared.MinerApiName:
			apis[i] = NewMinerApi(eth, codec)
		case shared.NetApiName:
			apis[i] = NewNetApi(xeth, eth, codec)
		case shared.ShhApiName:
			apis[i] = NewShhApi(xeth, eth, codec)
		case shared.TxPoolApiName:
			apis[i] = NewTxPoolApi(xeth, eth, codec)
		case shared.PersonalApiName:
			apis[i] = NewPersonalApi(xeth, eth, codec)
		case shared.Web3ApiName:
			apis[i] = NewWeb3Api(xeth, codec)
		default:
			return nil, fmt.Errorf("Unknown API '%s'", name)
		}
	}

	return apis, nil
}

func Javascript(name string) string {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case shared.AdminApiName:
		return Admin_JS
	case shared.DebugApiName:
		return Debug_JS
	case shared.DbApiName:
		return Db_JS
	case shared.EthApiName:
		return nch_JS
	case shared.MinerApiName:
		return Miner_JS
	case shared.NetApiName:
		return Net_JS
	case shared.ShhApiName:
		return Shh_JS
	case shared.TxPoolApiName:
		return TxPool_JS
	case shared.PersonalApiName:
		return Personal_JS
	}

	return ""
}
