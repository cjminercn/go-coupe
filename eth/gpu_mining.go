// Copyright 2014 The go-coupe Authors
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

// +build opencl

package eth

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/cjminercn/ethash"
	"github.com/cjminercn/go-coupe/common"
	"github.com/cjminercn/go-coupe/core/types"
	"github.com/cjminercn/go-coupe/logger"
	"github.com/cjminercn/go-coupe/logger/glog"
	"github.com/cjminercn/go-coupe/miner"
)

func (s *cjminercn) StartMining(threads int, gpus string) error {
	eb, err := s.Etherbase()
	if err != nil {
		err = fmt.Errorf("Cannot start mining without etherbase address: %v", err)
		glog.V(logger.Error).Infoln(err)
		return err
	}

	// GPU mining
	if gpus != "" {
		var ids []int
		for _, s := range strings.Split(gpus, ",") {
			i, err := strconv.Atoi(s)
			if err != nil {
				return fmt.Errorf("Invalid GPU id(s): %v", err)
			}
			if i < 0 {
				return fmt.Errorf("Invalid GPU id: %v", i)
			}
			ids = append(ids, i)
		}

		// TODO: re-creating miner is a bit ugly
		cl := ethash.NewCL(ids)
		s.miner = miner.New(s, s.EventMux(), cl)
		go s.miner.Start(eb, len(ids))
		return nil
	}

	// CPU mining
	go s.miner.Start(eb, threads)
	return nil
}

func GPUBench(gpuid uint64) {
	e := ethash.NewCL([]int{int(gpuid)})

	var h common.Hash
	bogoHeader := &types.Header{
		ParentHash: h,
		Number:     big.NewInt(int64(42)),
		Difficulty: big.NewInt(int64(999999999999999)),
	}
	bogoBlock := types.NewBlock(bogoHeader, nil, nil, nil)

	err := ethash.InitCL(bogoBlock.NumberU64(), e)
	if err != nil {
		fmt.Println("OpenCL init error: ", err)
		return
	}

	stopChan := make(chan struct{})
	reportHashRate := func() {
		for {
			time.Sleep(3 * time.Second)
			fmt.Printf("hashes/s : %v\n", e.GetHashrate())
		}
	}
	fmt.Printf("Starting benchmark (%v seconds)\n", 60)
	go reportHashRate()
	go e.Search(bogoBlock, stopChan, 0)
	time.Sleep(60 * time.Second)
	fmt.Println("OK.")
}

func PrintOpenCLDevices() {
	ethash.PrintDevices()
}
