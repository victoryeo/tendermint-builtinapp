package main

import (
	types "github.com/tendermint/tendermint/abci/types"
	"bytes"
	"github.com/dgraph-io/badger"
)

type KVStoreApplication struct {
	db		*badger.DB
	currentBatch 	*badger.Txn
}

var _ types.Application = (*KVStoreApplication)(nil)

func NewKVStoreApplication(db *badger.DB) *KVStoreApplication {
	return &KVStoreApplication{
		db: db,
	}
}

func (app *KVStoreApplication) Info(req types.RequestInfo) types.ResponseInfo {
	return types.ResponseInfo{}
}

func (app *KVStoreApplication) DeliverTx(req types.RequestDeliverTx) types.ResponseDeliverTx {
	code := app.isValid(req.Tx)
	if code != 0 {
		return types.ResponseDeliverTx{Code: code}
	}

	parts := bytes.Split(req.Tx, []byte("="))
	key, value := parts[0], parts[1]

	err := app.currentBatch.Set(key, value)
	if err != nil {
		panic(err)
	}

	return types.ResponseDeliverTx{Code: 0}
}

func (app *KVStoreApplication) CheckTx(req types.RequestCheckTx) types.ResponseCheckTx {
	code := app.isValid(req.Tx)
	return types.ResponseCheckTx{Code: code, GasWanted: 1}
}

func (app *KVStoreApplication) Commit() types.ResponseCommit {
	app.currentBatch.Commit()
	return types.ResponseCommit{Data: []byte{}}
}

func (app *KVStoreApplication) Query(reqQuery types.RequestQuery) (resQuery types.ResponseQuery) {
	resQuery.Key = reqQuery.Data
	err := app.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(reqQuery.Data)
		if err != nil && err != badger.ErrKeyNotFound {
			return err
		}
		if err == badger.ErrKeyNotFound {
			resQuery.Log = "does not exist"
		} else {
			return item.Value(func(val []byte) error {
				resQuery.Log = "exists"
				resQuery.Value = val
				return nil
			})
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	return
}

func (app *KVStoreApplication) isValid(tx []byte) (code uint32) {
	// check format
	parts := bytes.Split(tx, []byte("="))
	if len(parts) != 2 {
		return 1
	}

	key, value := parts[0], parts[1]

	// check if the same key=value already exists
	err := app.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil && err != badger.ErrKeyNotFound {
			return err
		}
		if err == nil {
			return item.Value(func(val []byte) error {
				if bytes.Equal(val, value) {
					code = 2
				}
				return nil
			})
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	return code
}

func (KVStoreApplication) InitChain(req types.RequestInitChain) types.ResponseInitChain {
	return types.ResponseInitChain{}
}

func (app *KVStoreApplication) ApplySnapshotChunk(req types.RequestApplySnapshotChunk) types.ResponseApplySnapshotChunk {
	return types.ResponseApplySnapshotChunk{}
}

func (app *KVStoreApplication) LoadSnapshotChunk(req types.RequestLoadSnapshotChunk) types.ResponseLoadSnapshotChunk {
	return types.ResponseLoadSnapshotChunk{}
}

func (app *KVStoreApplication) OfferSnapshot(req types.RequestOfferSnapshot) types.ResponseOfferSnapshot {
	return types.ResponseOfferSnapshot{}
}

func (app *KVStoreApplication) SetOption(req types.RequestSetOption) types.ResponseSetOption {
        return types.ResponseSetOption{
                Code: 0,
                Log:  "No options are supported yet",
        }
}

func (app *KVStoreApplication) BeginBlock(req types.RequestBeginBlock) types.ResponseBeginBlock {
	app.currentBatch = app.db.NewTransaction(true)
	return types.ResponseBeginBlock{}
}

func (KVStoreApplication) EndBlock(req types.RequestEndBlock) types.ResponseEndBlock {
	return types.ResponseEndBlock{}
}

func (KVStoreApplication) ListSnapshots(types.RequestListSnapshots) types.ResponseListSnapshots {
	return types.ResponseListSnapshots{}
}

