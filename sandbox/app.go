package main

import (
	"context"
	"encoding/base64"
	"fmt"
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

func (app *SandboxApp) FinalizeBlock(_ context.Context, req *types.RequestFinalizeBlock) (*types.ResponseFinalizeBlock, error) {
	respTxs := make([]*types.ExecTxResult, len(req.Txs))
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
