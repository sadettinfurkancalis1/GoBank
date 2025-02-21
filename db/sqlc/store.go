package db

import (
	"context"
	"database/sql"
	"fmt"
)

// Store provides all functions to execute db querieds

type Store struct {
	/*
		Store Yapısı (Store struct):
		Bu yapı, veritabanıyla ilgili işlemleri yapmak için kullanılan araçları bir arada tutuyor. İki ana parçası var:

		Queries: Veritabanına yapılacak sorguları içeren bir araç seti.
		db: Asıl veritabanı bağlantısı.
	*/
	*Queries
	db *sql.DB
}

// Newstore creates a new store
func NewStore(db *sql.DB) *Store {
	/*
		NewStore Fonksiyonu:
		Bu fonksiyon, verilen veritabanı bağlantısı (db) ile yeni bir Store nesnesi oluşturur. Yani, veritabanı işlemleri için gerekli araçları bir araya getirir.
	*/
	return &Store{
		db:      db,
		Queries: New(db),
	}
}

// ExecTx executes a function within a database transaction
// "store *Store", execTx fonksiyonunun hangi nesne üzerinde çalışacağını belirtir;
// context Bu parametre, fonksiyonlar arası iptal sinyalleri, zaman aşımı bilgisi ve istekle ilişkili değerlerin taşınmasını sağlar.
// fn sayesinde, gerçekleştirmek istediğiniz veritabanı işlemlerini tek bir transaction içinde topluyor, bu da veri tutarlılığını sağlamaya yardımcı oluyor.
// execTx fonksiyonunun yalnızca bir Store nesnesi üzerinden çağrılabileceği anlamına gelir.
func (store *Store) execTx(ctx context.Context, fn func(*Queries) error) error {
	/*
		ExecTx Fonksiyonu:
		Bu fonksiyon, veritabanında işlem (transaction) yapmayı sağlar. İşlemin nasıl çalıştığını şu şekilde özetleyebiliriz:

		Öncelikle veritabanında bir işlem (transaction) başlatılır.
		Kullanıcıdan gelen belirli bir işlemi (fonksiyon) çalıştırır.
		Eğer bu işlem sırasında hata oluşursa, yapılan tüm değişiklikler geri alınır (rollback), böylece veritabanı önceki haline döner.
		Her şey sorunsuz giderse, yapılan işlemler kaydedilir (commit).
	*/
	tx, err := store.db.BeginTx(ctx, nil)

	if err != nil {
		return err
	}

	// Create a new Queries instance using the active transaction (tx).
	q := New(tx)
	// Execute the provided function (fn) with the Queries instance.
	// This allows all database operations within fn to run inside the transaction.
	err = fn(q)

	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx error: %v, rb error: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}

// TransferTxParams contains the input parameters of the transfer transaction
type TransferTxParams struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

// TransferTxResult is the result of the transfer transaction
type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}

// TransferTx performs a money transfer from one account to the other
// It creates a transfer record, add account entries, and update accounts' balance within a single database transaction
func (store *Store) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult

	/*
		Python Wrapper
		with db_connection.begin() as tx:
			transfer = create_transfer(tx, from_account_id, to_account_id, amount)
			from_entry = create_entry(tx, from_account_id, -amount)
			to_entry = create_entry(tx, to_account_id, amount)
			# TODO: update accounts' balance
		# hata olduğu durumda 'with' bloğu otomatik rollback yapar
	*/
	err := store.execTx(ctx, func(q *Queries) error {
		// anonim fonksiyon (closure)
		var err error

		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: arg.FromAccountID,
			ToAccountID:   arg.ToAccountID,
			Amount:        arg.Amount,
		})

		if err != nil {
			return err
		}

		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount:    -arg.Amount,
		})

		if err != nil {
			return err
		}

		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount,
		})
		if err != nil {
			return err
		}

		//  update accounts' balance

		account1, err := q.GetAccount(ctx, arg.FromAccountID)
		if err != nil {
			return err
		}

		result.FromAccount, err = q.UpdateAccount(ctx, UpdateAccountParams{
			ID:      account1.ID,

		return nil
	})

	return result, err

}
