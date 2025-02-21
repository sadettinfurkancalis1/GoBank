package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

/*
Aşağıdaki açıklamada, bu test kodunun ne yaptığını adım adım anlayabilirsiniz:

Store ve Hesapların Oluşturulması:

store := NewStore(testDB) ifadesi, test veritabanı (testDB) kullanılarak yeni bir Store nesnesi oluşturur.
account1 := CreateRandomAccount(t) ve account2 := CreateRandomAccount(t) ile iki adet rastgele hesap (örneğin, para transferi için gönderici ve alıcı) oluşturuluyor.
Test Parametreleri ve Kanal Tanımları:

n := 5 ifadesi, 5 adet eşzamanlı (concurrent) para transferi işlemi yapılacağını belirtir.
amount := int64(10) ile her işlemde transfer edilecek miktar 10 birim olarak ayarlanıyor.
errs := make(chan error) ve results := make(chan TransferTxResult) ile hata ve sonuçların toplanacağı kanallar (channels) oluşturuluyor.
Python’daki queue.Queue veya benzeri yapılarla karşılaştırılabilir. Kanallar, goroutine’ler arasında veri iletişimi için kullanılır.
Eşzamanlı (Concurrent) Transfer İşlemleri:

for i := 0; i < n; i++ { ... } döngüsü içinde her iterasyonda yeni bir goroutine oluşturuluyor.
Go’daki go func() { ... }() ifadesi, fonksiyonu yeni bir thread benzeri yapı (goroutine) içinde eşzamanlı çalıştırır.
Her goroutine, store.TransferTx metodunu çağırarak para transfer işlemini gerçekleştirir.
Burada context.Background() kullanılarak context oluşturulur; bu, işlemin iptali veya zaman aşımı gibi durumları yönetmek için gereklidir.
TransferTxParams ile transferin detayları (gönderen, alıcı, miktar) aktarılır.
İşlem tamamlandığında, eğer hata varsa errs <- err ile, işlem sonucu varsa da results <- result ile ilgili kanal üzerinden gönderilir.
Özetle, bu test kodu, aynı anda 5 tane para transfer işlemini eşzamanlı olarak gerçekleştirir. Her bir transfer işlemi sonucunda hata veya işlem sonuçları kanallara gönderilir. Bu yapı, Python'da concurrent.futures veya asyncio kullanılarak benzer şekilde eşzamanlı testler yazmaya benzer; önemli olan, aynı anda birçok işlemin doğru şekilde yürütülmesini kontrol etmektir.
*/
func TestTransferTx(t *testing.T){
	store := NewStore(testDB)

	account1 := CreateRandomAccount(t)
	account2 := CreateRandomAccount(t)
	fmt.Println("Before:", account1.Balance, account2.Balance)

	// run n concurrent transfer transactions
	n := 5
	amount := int64(10)

	// TODO: learn
	// its like a queue
	errs := make(chan error)
	results := make(chan TransferTxResult)

	for i := 0; i < n; i++ {
		go func() {
			result, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount: 	  amount,
			})

			errs <- err
			results <- result
		}()
	}

	// check results
	existed := make(map[int]bool)
	for i := 0; i < n; i++ {
		// kanalından sırayla bir adet hata değeri alır. Kanallar FIFO (first-in, first-out) mantığıyla çalışır.
		err := <-errs
		require.NoError(t, err)

		result := <- results
		require.NotEmpty(t, result)

		// check transfer
		transfer := result.Transfer
		require.NotEmpty(t, transfer)
		require.Equal(t, account1.ID, transfer.FromAccountID)
		require.Equal(t, account2.ID, transfer.ToAccountID)
		require.Equal(t, amount, transfer.Amount)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)
		
		_, err = store.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)

		// check entries
		fromEntry := result.FromEntry
		require.NotEmpty(t, fromEntry)
		require.Equal(t, account1.ID, fromEntry.AccountID)
		require.Equal(t, -amount, fromEntry.Amount)
		require.NotZero(t, fromEntry.ID)
		require.NotZero(t, fromEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), fromEntry.ID)
		require.NoError(t, err)


		toEntry := result.ToEntry
		require.NotEmpty(t, toEntry)
		require.Equal(t, account2.ID, toEntry.AccountID)
		require.Equal(t, amount, toEntry.Amount)
		require.NotZero(t, toEntry.ID)
		require.NotZero(t, toEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), toEntry.ID)
		require.NoError(t, err)


		// check accounts
		fromAccount := result.FromAccount
		require.NotEmpty(t, fromAccount)
		require.Equal(t, account1.ID, fromAccount.ID)

		toAccount := result.ToAccount
		require.NotEmpty(t, toAccount)
		require.Equal(t, account2.ID, toAccount.ID)


		// check account balances
		fmt.Println("After:", fromAccount.Balance, toAccount.Balance)
		diff1 := account1.Balance - fromAccount.Balance
		diff2 := toAccount.Balance - account2.Balance
		require.Equal(t, diff1, diff2)
		require.True(t, diff1 > 0)
		require.True(t, diff1 % amount == 0) // 1 * amount, 2 * amount, 3 * amount, ...

		k := int(diff1 / amount)
		require.True(t, k >= 1 && k <= n)
		require.NotContains(t, existed, k)
		existed[k] = true
	}	

	// check final updated balances
	updateAccount1, err := testQueries.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)

	updateAccount2, err := testQueries.GetAccount(context.Background(), account2.ID)
	require.NoError()

	require.Equal(t, account1.Balance - int64(n) * amount, updateAccount1.Balance)
	require.Equal(t, account2.Balance + int64(n) * amount, updateAccount2.Balance)
	
	fmt.Println("After:", updateAccount1.Balance, updateAccount2.Balance)
}