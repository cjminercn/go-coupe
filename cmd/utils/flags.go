// Copyright 2015 The go-coupe Authors
// This file is part of go-coupe.
//
// go-coupe is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-coupe is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-coupe. If not, see <http://www.gnu.org/licenses/>.

package utils

import (
	"crypto/ecdsa"
	"fmt"
	"log"
	"math"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/codegangsta/cli"
	"github.com/cjminercn/ethash"
	"github.com/cjminercn/go-coupe/accounts"
	"github.com/cjminercn/go-coupe/common"
	"github.com/cjminercn/go-coupe/core"
	"github.com/cjminercn/go-coupe/core/vm"
	"github.com/cjminercn/go-coupe/crypto"
	"github.com/cjminercn/go-coupe/eth"
	"github.com/cjminercn/go-coupe/ethdb"
	"github.com/cjminercn/go-coupe/event"
	"github.com/cjminercn/go-coupe/logger"
	"github.com/cjminercn/go-coupe/logger/glog"
	"github.com/cjminercn/go-coupe/metrics"
	"github.com/cjminercn/go-coupe/p2p/nat"
	"github.com/cjminercn/go-coupe/params"
	"github.com/cjminercn/go-coupe/rpc/api"
	"github.com/cjminercn/go-coupe/rpc/codec"
	"github.com/cjminercn/go-coupe/rpc/comms"
	"github.com/cjminercn/go-coupe/rpc/shared"
	"github.com/cjminercn/go-coupe/rpc/useragent"
	"github.com/cjminercn/go-coupe/xeth"
)

func init() {
	cli.AppHelpTemplate = `{{.Name}} {{if .Flags}}[global options] {{end}}command{{if .Flags}} [command options]{{end}} [arguments...]

VERSION:
   {{.Version}}

COMMANDS:
   {{range .Commands}}{{.Name}}{{with .ShortName}}, {{.}}{{end}}{{ "\t" }}{{.Usage}}
   {{end}}{{if .Flags}}
GLOBAL OPTIONS:
   {{range .Flags}}{{.}}
   {{end}}{{end}}
`

	cli.CommandHelpTemplate = `{{.Name}}{{if .Subcommands}} command{{end}}{{if .Flags}} [command options]{{end}} [arguments...]
{{if .Description}}{{.Description}}
{{end}}{{if .Subcommands}}
SUBCOMMANDS:
	{{range .Subcommands}}{{.Name}}{{with .ShortName}}, {{.}}{{end}}{{ "\t" }}{{.Usage}}
	{{end}}{{end}}{{if .Flags}}
OPTIONS:
	{{range .Flags}}{{.}}
	{{end}}{{end}}
`
}

// NewApp creates an app with sane defaults.
func NewApp(version, usage string) *cli.App {
	app := cli.NewApp()
	app.Name = filepath.Base(os.Args[0])
	app.Author = ""
	//app.Authors = nil
	app.Email = ""
	app.Version = version
	app.Usage = usage
	return app
}

// These are all the command line flags we support.
// If you add to this list, please remember to include the
// flag in the appropriate command definition.
//
// The flags are defined here so their names and help texts
// are the same for all commands.

var (
	// General settings
	DataDirFlag = DirectoryFlag{
		Name:  "datadir",
		Usage: "Data directory for the databases and keystore",
		Value: DirectoryString{common.DefaultDataDir()},
	}
	NetworkIdFlag = cli.IntFlag{
		Name:  "networkid",
		Usage: "Network identifier (integer, 0=Olympic, 1=Frontier, 2=Morden)",
		Value: eth.NetworkId,
	}
	OlympicFlag = cli.BoolFlag{
		Name:  "olympic",
		Usage: "Olympic network: pre-configured pre-release test network",
	}
	TestNetFlag = cli.BoolFlag{
		Name:  "testnet",
		Usage: "Morden network: pre-configured test network with modified starting nonces (replay protection)",
	}
	DevModeFlag = cli.BoolFlag{
		Name:  "dev",
		Usage: "Developer mode: pre-configured private network with several debugging flags",
	}
	GenesisFileFlag = cli.StringFlag{
		Name:  "genesis",
		Usage: "Insert/overwrite the genesis block (JSON format)",
	}
	IdentityFlag = cli.StringFlag{
		Name:  "identity",
		Usage: "Custom node name",
	}
	NatspecEnabledFlag = cli.BoolFlag{
		Name:  "natspec",
		Usage: "Enable NatSpec confirmation notice",
	}
	DocRootFlag = DirectoryFlag{
		Name:  "docroot",
		Usage: "Document Root for HTTPClient file scheme",
		Value: DirectoryString{common.HomeDir()},
	}
	CacheFlag = cli.IntFlag{
		Name:  "cache",
		Usage: "Megabytes of memory allocated to internal caching (min 16MB / database forced)",
		Value: 0,
	}
	BlockchainVersionFlag = cli.IntFlag{
		Name:  "blockchainversion",
		Usage: "Blockchain version (integer)",
		Value: core.BlockChainVersion,
	}
	FastSyncFlag = cli.BoolFlag{
		Name:  "fast",
		Usage: "Enable fast syncing through state downloads",
	}
	LightKDFFlag = cli.BoolFlag{
		Name:  "lightkdf",
		Usage: "Reduce key-derivation RAM & CPU usage at some expense of KDF strength",
	}
	// Miner settings
	// TODO: refactor CPU vs GPU mining flags
	MiningEnabledFlag = cli.BoolFlag{
		Name:  "mine",
		Usage: "Enable mining",
	}
	MinerThreadsFlag = cli.IntFlag{
		Name:  "minerthreads",
		Usage: "Number of CPU threads to use for mining",
		Value: runtime.NumCPU(),
	}
	MiningGPUFlag = cli.StringFlag{
		Name:  "minergpus",
		Usage: "List of GPUs to use for mining (e.g. '0,1' will use the first two GPUs found)",
	}
	AutoDAGFlag = cli.BoolFlag{
		Name:  "autodag",
		Usage: "Enable automatic DAG pregeneration",
	}
	EtherbaseFlag = cli.StringFlag{
		Name:  "etherbase",
		Usage: "Public address for block mining rewards (default = first account created)",
		Value: "0",
	}
	GasPriceFlag = cli.StringFlag{
		Name:  "gasprice",
		Usage: "Minimal gas price to accept for mining a transactions",
		Value: new(big.Int).Mul(big.NewInt(50), common.Shannon).String(),
	}
	ExtraDataFlag = cli.StringFlag{
		Name:  "extradata",
		Usage: "Block extra data set by the miner (default = client version)",
	}
	// Account settings
	UnlockedAccountFlag = cli.StringFlag{
		Name:  "unlock",
		Usage: "Unlock an account (may be creation index) until this program exits (prompts for password)",
		Value: "",
	}
	PasswordFileFlag = cli.StringFlag{
		Name:  "password",
		Usage: "Password file to use with options/subcommands needing a pass phrase",
		Value: "",
	}

	// vm flags
	VMDebugFlag = cli.BoolFlag{
		Name:  "vmdebug",
		Usage: "Virtual Machine debug output",
	}
	VMForceJitFlag = cli.BoolFlag{
		Name:  "forcejit",
		Usage: "Force the JIT VM to take precedence",
	}
	VMJitCacheFlag = cli.IntFlag{
		Name:  "jitcache",
		Usage: "Amount of cached JIT VM programs",
		Value: 64,
	}
	VMEnableJitFlag = cli.BoolFlag{
		Name:  "jitvm",
		Usage: "Enable the JIT VM",
	}

	// logging and debug settings
	VerbosityFlag = cli.IntFlag{
		Name:  "verbosity",
		Usage: "Logging verbosity: 0-6 (0=silent, 1=error, 2=warn, 3=info, 4=core, 5=debug, 6=debug detail)",
		Value: int(logger.InfoLevel),
	}
	LogFileFlag = cli.StringFlag{
		Name:  "logfile",
		Usage: "Log output file within the data dir (default = no log file generated)",
		Value: "",
	}
	LogVModuleFlag = cli.GenericFlag{
		Name:  "vmodule",
		Usage: "Per-module verbosity: comma-separated list of <module>=<level>, where <module> is file literal or a glog pattern",
		Value: glog.GetVModule(),
	}
	BacktraceAtFlag = cli.GenericFlag{
		Name:  "backtrace",
		Usage: "Request a stack trace at a specific logging statement (e.g. \"block.go:271\")",
		Value: glog.GetTraceLocation(),
	}
	PProfEanbledFlag = cli.BoolFlag{
		Name:  "pprof",
		Usage: "Enable the profiling server on localhost",
	}
	PProfPortFlag = cli.IntFlag{
		Name:  "pprofport",
		Usage: "Profile server listening port",
		Value: 6060,
	}
	MetricsEnabledFlag = cli.BoolFlag{
		Name:  metrics.MetricsEnabledFlag,
		Usage: "Enable metrics collection and reporting",
	}

	// RPC settings
	RPCEnabledFlag = cli.BoolFlag{
		Name:  "rpc",
		Usage: "Enable the HTTP-RPC server",
	}
	RPCListenAddrFlag = cli.StringFlag{
		Name:  "rpcaddr",
		Usage: "HTTP-RPC server listening interface",
		Value: "127.0.0.1",
	}
	RPCPortFlag = cli.IntFlag{
		Name:  "rpcport",
		Usage: "HTTP-RPC server listening port",
		Value: 8545,
	}
	RPCCORSDomainFlag = cli.StringFlag{
		Name:  "rpccorsdomain",
		Usage: "Domains from which to accept cross origin requests (browser enforced)",
		Value: "",
	}
	RpcApiFlag = cli.StringFlag{
		Name:  "rpcapi",
		Usage: "API's offered over the HTTP-RPC interface",
		Value: comms.DefaultHttpRpcApis,
	}
	IPCDisabledFlag = cli.BoolFlag{
		Name:  "ipcdisable",
		Usage: "Disable the IPC-RPC server",
	}
	IPCApiFlag = cli.StringFlag{
		Name:  "ipcapi",
		Usage: "API's offered over the IPC-RPC interface",
		Value: comms.DefaultIpcApis,
	}
	IPCPathFlag = DirectoryFlag{
		Name:  "ipcpath",
		Usage: "Filename for IPC socket/pipe",
		Value: DirectoryString{common.DefaultIpcPath()},
	}
	ExecFlag = cli.StringFlag{
		Name:  "exec",
		Usage: "Execute JavaScript statement (only in combination with console/attach)",
	}
	// Network Settings
	MaxPeersFlag = cli.IntFlag{
		Name:  "maxpeers",
		Usage: "Maximum number of network peers (network disabled if set to 0)",
		Value: 25,
	}
	MaxPendingPeersFlag = cli.IntFlag{
		Name:  "maxpendpeers",
		Usage: "Maximum number of pending connection attempts (defaults used if set to 0)",
		Value: 0,
	}
	ListenPortFlag = cli.IntFlag{
		Name:  "port",
		Usage: "Network listening port",
		Value: 30303,
	}
	BootnodesFlag = cli.StringFlag{
		Name:  "bootnodes",
		Usage: "Space-separated enode URLs for P2P discovery bootstrap",
		Value: "",
	}
	NodeKeyFileFlag = cli.StringFlag{
		Name:  "nodekey",
		Usage: "P2P node key file",
	}
	NodeKeyHexFlag = cli.StringFlag{
		Name:  "nodekeyhex",
		Usage: "P2P node key as hex (for testing)",
	}
	NATFlag = cli.StringFlag{
		Name:  "nat",
		Usage: "NAT port mapping mechanism (any|none|upnp|pmp|extip:<IP>)",
		Value: "any",
	}
	NoDiscoverFlag = cli.BoolFlag{
		Name:  "nodiscover",
		Usage: "Disables the peer discovery mechanism (manual peer addition)",
	}
	WhisperEnabledFlag = cli.BoolFlag{
		Name:  "shh",
		Usage: "Enable Whisper",
	}
	// ATM the url is left to the user and deployment to
	JSpathFlag = cli.StringFlag{
		Name:  "jspath",
		Usage: "JavaSript root path for `loadScript` and document root for `admin.httpGet`",
		Value: ".",
	}
	SolcPathFlag = cli.StringFlag{
		Name:  "solc",
		Usage: "Solidity compiler command to be used",
		Value: "solc",
	}

	// Gas price oracle settings
	GpoMinGasPriceFlag = cli.StringFlag{
		Name:  "gpomin",
		Usage: "Minimum suggested gas price",
		Value: new(big.Int).Mul(big.NewInt(50), common.Shannon).String(),
	}
	GpoMaxGasPriceFlag = cli.StringFlag{
		Name:  "gpomax",
		Usage: "Maximum suggested gas price",
		Value: new(big.Int).Mul(big.NewInt(500), common.Shannon).String(),
	}
	GpoFullBlockRatioFlag = cli.IntFlag{
		Name:  "gpofull",
		Usage: "Full block threshold for gas price calculation (%)",
		Value: 80,
	}
	GpobaseStepDownFlag = cli.IntFlag{
		Name:  "gpobasedown",
		Usage: "Suggested gas price base step down ratio (1/1000)",
		Value: 10,
	}
	GpobaseStepUpFlag = cli.IntFlag{
		Name:  "gpobaseup",
		Usage: "Suggested gas price base step up ratio (1/1000)",
		Value: 100,
	}
	GpobaseCorrectionFactorFlag = cli.IntFlag{
		Name:  "gpobasecf",
		Usage: "Suggested gas price base correction factor (%)",
		Value: 110,
	}
)

// MakeNAT creates a port mapper from set command line flags.
func MakeNAT(ctx *cli.Context) nat.Interface {
	natif, err := nat.Parse(ctx.GlobalString(NATFlag.Name))
	if err != nil {
		Fatalf("Option %s: %v", NATFlag.Name, err)
	}
	return natif
}

// MakeNodeKey creates a node key from set command line flags.
func MakeNodeKey(ctx *cli.Context) (key *ecdsa.PrivateKey) {
	hex, file := ctx.GlobalString(NodeKeyHexFlag.Name), ctx.GlobalString(NodeKeyFileFlag.Name)
	var err error
	switch {
	case file != "" && hex != "":
		Fatalf("Options %q and %q are mutually exclusive", NodeKeyFileFlag.Name, NodeKeyHexFlag.Name)
	case file != "":
		if key, err = crypto.LoadECDSA(file); err != nil {
			Fatalf("Option %q: %v", NodeKeyFileFlag.Name, err)
		}
	case hex != "":
		if key, err = crypto.HexToECDSA(hex); err != nil {
			Fatalf("Option %q: %v", NodeKeyHexFlag.Name, err)
		}
	}
	return key
}

// MakeEthConfig creates cjminercn options from set command line flags.
func MakeEthConfig(clientID, version string, ctx *cli.Context) *eth.Config {
	customName := ctx.GlobalString(IdentityFlag.Name)
	if len(customName) > 0 {
		clientID += "/" + customName
	}
	am := MakeAccountManager(ctx)
	etherbase, err := ParamToAddress(ctx.GlobalString(EtherbaseFlag.Name), am)
	if err != nil {
		glog.V(logger.Error).Infoln("WARNING: No etherbase set and no accounts found as default")
	}
	// Assemble the entire eth configuration and return
	cfg := &eth.Config{
		Name:                    common.MakeName(clientID, version),
		DataDir:                 MustDataDir(ctx),
		GenesisFile:             ctx.GlobalString(GenesisFileFlag.Name),
		FastSync:                ctx.GlobalBool(FastSyncFlag.Name),
		BlockChainVersion:       ctx.GlobalInt(BlockchainVersionFlag.Name),
		DatabaseCache:           ctx.GlobalInt(CacheFlag.Name),
		SkipBcVersionCheck:      false,
		NetworkId:               ctx.GlobalInt(NetworkIdFlag.Name),
		LogFile:                 ctx.GlobalString(LogFileFlag.Name),
		Verbosity:               ctx.GlobalInt(VerbosityFlag.Name),
		Etherbase:               common.HexToAddress(etherbase),
		MinerThreads:            ctx.GlobalInt(MinerThreadsFlag.Name),
		AccountManager:          am,
		VmDebug:                 ctx.GlobalBool(VMDebugFlag.Name),
		MaxPeers:                ctx.GlobalInt(MaxPeersFlag.Name),
		MaxPendingPeers:         ctx.GlobalInt(MaxPendingPeersFlag.Name),
		Port:                    ctx.GlobalString(ListenPortFlag.Name),
		Olympic:                 ctx.GlobalBool(OlympicFlag.Name),
		NAT:                     MakeNAT(ctx),
		NatSpec:                 ctx.GlobalBool(NatspecEnabledFlag.Name),
		DocRoot:                 ctx.GlobalString(DocRootFlag.Name),
		Discovery:               !ctx.GlobalBool(NoDiscoverFlag.Name),
		NodeKey:                 MakeNodeKey(ctx),
		Shh:                     ctx.GlobalBool(WhisperEnabledFlag.Name),
		Dial:                    true,
		BootNodes:               ctx.GlobalString(BootnodesFlag.Name),
		GasPrice:                common.String2Big(ctx.GlobalString(GasPriceFlag.Name)),
		GpoMinGasPrice:          common.String2Big(ctx.GlobalString(GpoMinGasPriceFlag.Name)),
		GpoMaxGasPrice:          common.String2Big(ctx.GlobalString(GpoMaxGasPriceFlag.Name)),
		GpoFullBlockRatio:       ctx.GlobalInt(GpoFullBlockRatioFlag.Name),
		GpobaseStepDown:         ctx.GlobalInt(GpobaseStepDownFlag.Name),
		GpobaseStepUp:           ctx.GlobalInt(GpobaseStepUpFlag.Name),
		GpobaseCorrectionFactor: ctx.GlobalInt(GpobaseCorrectionFactorFlag.Name),
		SolcPath:                ctx.GlobalString(SolcPathFlag.Name),
		AutoDAG:                 ctx.GlobalBool(AutoDAGFlag.Name) || ctx.GlobalBool(MiningEnabledFlag.Name),
	}

	if ctx.GlobalBool(DevModeFlag.Name) && ctx.GlobalBool(TestNetFlag.Name) {
		glog.Fatalf("%s and %s are mutually exclusive\n", DevModeFlag.Name, TestNetFlag.Name)
	}

	if ctx.GlobalBool(TestNetFlag.Name) {
		// testnet is always stored in the testnet folder
		cfg.DataDir += "/testnet"
		cfg.NetworkId = 2
		cfg.TestNet = true
	}

	if ctx.GlobalBool(VMEnableJitFlag.Name) {
		cfg.Name += "/JIT"
	}
	if ctx.GlobalBool(DevModeFlag.Name) {
		if !ctx.GlobalIsSet(VMDebugFlag.Name) {
			cfg.VmDebug = true
		}
		if !ctx.GlobalIsSet(MaxPeersFlag.Name) {
			cfg.MaxPeers = 0
		}
		if !ctx.GlobalIsSet(GasPriceFlag.Name) {
			cfg.GasPrice = new(big.Int)
		}
		if !ctx.GlobalIsSet(ListenPortFlag.Name) {
			cfg.Port = "0" // auto port
		}
		if !ctx.GlobalIsSet(WhisperEnabledFlag.Name) {
			cfg.Shh = true
		}
		if !ctx.GlobalIsSet(DataDirFlag.Name) {
			cfg.DataDir = os.TempDir() + "/cjminercn_dev_mode"
		}
		cfg.PowTest = true
		cfg.DevMode = true

		glog.V(logger.Info).Infoln("dev mode enabled")
	}
	return cfg
}

// SetupLogger configures glog from the logging-related command line flags.
func SetupLogger(ctx *cli.Context) {
	glog.SetV(ctx.GlobalInt(VerbosityFlag.Name))
	glog.CopyStandardLogTo("INFO")
	glog.SetToStderr(true)
	glog.SetLogDir(ctx.GlobalString(LogFileFlag.Name))
}

// SetupNetwork configures the system for either the main net or some test network.
func SetupNetwork(ctx *cli.Context) {
	switch {
	case ctx.GlobalBool(OlympicFlag.Name):
		params.DurationLimit = big.NewInt(8)
		params.GenesisGasLimit = big.NewInt(3141592)
		params.MinGasLimit = big.NewInt(125000)
		params.MaximumExtraDataSize = big.NewInt(1024)
		NetworkIdFlag.Value = 0
		core.BlockReward = big.NewInt(1.5e+18)
		core.ExpDiffPeriod = big.NewInt(math.MaxInt64)
	}
}

// SetupVM configured the VM package's global settings
func SetupVM(ctx *cli.Context) {
	vm.EnableJit = ctx.GlobalBool(VMEnableJitFlag.Name)
	vm.ForceJit = ctx.GlobalBool(VMForceJitFlag.Name)
	vm.SetJITCacheSize(ctx.GlobalInt(VMJitCacheFlag.Name))
}

// MakeChain creates a chain manager from set command line flags.
func MakeChain(ctx *cli.Context) (chain *core.BlockChain, chainDb ethdb.Database) {
	datadir := MustDataDir(ctx)
	cache := ctx.GlobalInt(CacheFlag.Name)

	var err error
	if chainDb, err = ethdb.NewLDBDatabase(filepath.Join(datadir, "chaindata"), cache); err != nil {
		Fatalf("Could not open database: %v", err)
	}
	if ctx.GlobalBool(OlympicFlag.Name) {
		_, err := core.WriteTestNetGenesisBlock(chainDb, 42)
		if err != nil {
			glog.Fatalln(err)
		}
	}

	eventMux := new(event.TypeMux)
	pow := ethash.New()
	//genesis := core.GenesisBlock(uint64(ctx.GlobalInt(GenesisNonceFlag.Name)), blockDB)
	chain, err = core.NewBlockChain(chainDb, pow, eventMux)
	if err != nil {
		Fatalf("Could not start chainmanager: %v", err)
	}

	return chain, chainDb
}

// MakeChain creates an account manager from set command line flags.
func MakeAccountManager(ctx *cli.Context) *accounts.Manager {
	dataDir := MustDataDir(ctx)
	if ctx.GlobalBool(TestNetFlag.Name) {
		dataDir += "/testnet"
	}
	scryptN := crypto.StandardScryptN
	scryptP := crypto.StandardScryptP
	if ctx.GlobalBool(LightKDFFlag.Name) {
		scryptN = crypto.LightScryptN
		scryptP = crypto.LightScryptP
	}
	ks := crypto.NewKeyStorePassphrase(filepath.Join(dataDir, "keystore"), scryptN, scryptP)
	return accounts.NewManager(ks)
}

// MustDataDir retrieves the currently requested data directory, terminating if
// none (or the empty string) is specified.
func MustDataDir(ctx *cli.Context) string {
	if path := ctx.GlobalString(DataDirFlag.Name); path != "" {
		return path
	}
	Fatalf("Cannot determine default data directory, please set manually (--datadir)")
	return ""
}

func IpcSocketPath(ctx *cli.Context) (ipcpath string) {
	if runtime.GOOS == "windows" {
		ipcpath = common.DefaultIpcPath()
		if ctx.GlobalIsSet(IPCPathFlag.Name) {
			ipcpath = ctx.GlobalString(IPCPathFlag.Name)
		}
	} else {
		ipcpath = common.DefaultIpcPath()
		if ctx.GlobalIsSet(DataDirFlag.Name) {
			ipcpath = filepath.Join(ctx.GlobalString(DataDirFlag.Name), "geth.ipc")
		}
		if ctx.GlobalIsSet(IPCPathFlag.Name) {
			ipcpath = ctx.GlobalString(IPCPathFlag.Name)
		}
	}

	return
}

func StartIPC(eth *eth.cjminercn, ctx *cli.Context) error {
	config := comms.IpcConfig{
		Endpoint: IpcSocketPath(ctx),
	}

	initializer := func(conn net.Conn) (comms.Stopper, shared.cjminercnApi, error) {
		fe := useragent.NewRemoteFrontend(conn, eth.AccountManager())
		xeth := xeth.New(eth, fe)
		apis, err := api.ParseApiString(ctx.GlobalString(IPCApiFlag.Name), codec.JSON, xeth, eth)
		if err != nil {
			return nil, nil, err
		}
		return xeth, api.Merge(apis...), nil
	}

	return comms.StartIpc(config, codec.JSON, initializer)
}

func StartRPC(eth *eth.cjminercn, ctx *cli.Context) error {
	config := comms.HttpConfig{
		ListenAddress: ctx.GlobalString(RPCListenAddrFlag.Name),
		ListenPort:    uint(ctx.GlobalInt(RPCPortFlag.Name)),
		CorsDomain:    ctx.GlobalString(RPCCORSDomainFlag.Name),
	}

	xeth := xeth.New(eth, nil)
	codec := codec.JSON

	apis, err := api.ParseApiString(ctx.GlobalString(RpcApiFlag.Name), codec, xeth, eth)
	if err != nil {
		return err
	}

	return comms.StartHttp(config, codec, api.Merge(apis...))
}

func StartPProf(ctx *cli.Context) {
	address := fmt.Sprintf("localhost:%d", ctx.GlobalInt(PProfPortFlag.Name))
	go func() {
		log.Println(http.ListenAndServe(address, nil))
	}()
}

func ParamToAddress(addr string, am *accounts.Manager) (addrHex string, err error) {
	if !((len(addr) == 40) || (len(addr) == 42)) { // with or without 0x
		index, err := strconv.Atoi(addr)
		if err != nil {
			Fatalf("Invalid account address '%s'", addr)
		}

		addrHex, err = am.AddressByIndex(index)
		if err != nil {
			return "", err
		}
	} else {
		addrHex = addr
	}
	return
}
