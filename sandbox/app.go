package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"sort"
	"strconv"

	"github.com/tendermint/tendermint/abci/types"
)

var _ types.Application = (*SandboxApp)(nil)

type SandboxApp struct {
	types.BaseApplication
}

func NewSandboxApp() *SandboxApp {
	return &SandboxApp{}
}

func (app *SandboxApp) CheckTx(_ context.Context, req *types.RequestCheckTx) (*types.ResponseCheckTx, error) {
	code, gas := app.estimateTx(req.Tx)
	fmt.Printf("CheckTx Gas = %d\n", gas)
	return &types.ResponseCheckTx{Code: code, GasWanted: gas}, nil
}

func (app *SandboxApp) PrepareProposal(_ context.Context, req *types.RequestPrepareProposal) (*types.ResponsePrepareProposal, error) {
	// Let's sort transactions in the most profitable way according to gas consumption
	txs := append(req.Txs[:0:0], req.Txs...)
	sort.Slice(txs, func(i, j int) bool {
		_, gasI := app.estimateTx(txs[i])
		_, gasJ := app.estimateTx(txs[j])
		return gasI > gasJ
	})

	trs := make([]*types.TxRecord, 0, len(txs))
	var totalBytes int64
	for _, tx := range txs {
		totalBytes += int64(len(tx))
		if totalBytes > req.MaxTxBytes {
			break
		}
		trs = append(trs, &types.TxRecord{
			Action: types.TxRecord_UNMODIFIED,
			Tx:     tx,
		})
	}
	return &types.ResponsePrepareProposal{TxRecords: trs}, nil
}

func (app *SandboxApp) FinalizeBlock(_ context.Context, req *types.RequestFinalizeBlock) (*types.ResponseFinalizeBlock, error) {
	respTxs := make([]*types.ExecTxResult, len(req.Txs))
	// At this stage the transactions are already sorted
	for i, tx := range req.Txs {
		respTxs[i] = app.handleTx(tx)
	}

	return &types.ResponseFinalizeBlock{TxResults: respTxs}, nil
}

func (app *SandboxApp) handleTx(tx []byte) *types.ExecTxResult {
	code, gas := app.estimateTx(tx)
	fmt.Printf("handleTx Gas = %d\n", gas)
	return &types.ExecTxResult{Code: code, GasWanted: gas, GasUsed: gas}
}

func (app *SandboxApp) estimateTx(tx []byte) (code uint32, gas int64) {
	decoded, err := base64.StdEncoding.DecodeString(string(tx[:]))
	if err != nil {
		return 1, 0
	}
	value, err := strconv.ParseInt(string(decoded[:]), 10, 64)
	if err != nil {
		return 2, 0
	}
	if value <= 0 {
		return 3, 0
	}
	return 0, value
}
