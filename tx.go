package sq

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
)

type Tx struct {
	Core *sqlx.Tx
}
func newTx(tx *sqlx.Tx) *Tx {
	return &Tx{tx}
}
func (Tx) Commit() TxResult {
	return TxResult{
		isCommit: true,
	}
}
func (Tx) Rollback() TxResult {
	return TxResult{
		isCommit: false,
	}
}
func (Tx) RollbackWithError(err error) TxResult {
	return TxResult{
		isCommit: false,
		withError: err,
	}
}
type TxResult struct {
	isCommit bool
	withError error
}
// 给 TxResult 增加 Error 接口是为了避出现类似  tx.Rollback() 前面没有 return 的错误
func (result TxResult) Error() string {
	if result.withError != nil {
		return result.withError.Error()
	}
	return fmt.Sprintf("%+v", result)
}

func (db *Database) Transaction(ctx context.Context, handle func (tx *Tx) TxResult) (err error) {
	return db.TransactionOpts(ctx, handle, nil)
}
func (db *Database) TransactionOpts(ctx context.Context, handle func (tx *Tx) TxResult, opts *sql.TxOptions) (err error) {
	coreTx, err := db.Core.BeginTxx(ctx, opts) ; if err != nil {
		return
	}
	tx := newTx(coreTx)
	txResult := handle(tx)
	if txResult.isCommit {
		err = tx.Core.Commit() ; if err != nil {
			return
		}
	} else {
		err = tx.Core.Rollback() ; if err != nil {
			return
		}
		if txResult.withError != nil {
			return txResult.withError
		}
	}
	return
}