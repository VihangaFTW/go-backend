package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	store := NewStore(testDB)

	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	n := 5
	amount := int64(10)

	//? channels to retrieve the result and errors from separate goroutines into the main thread
	errors := make(chan error)
	results := make(chan TransferTxResult)

	//? run n concurrent transfer transactions
	for range n {
		go func() {
			result, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount:        amount,
			})

			errors <- err
			results <- result

		}()
	}
	//? ensure each transaction executes exactly once
	quotientMap := make(map[int]bool)

	// check all results in channel
	//! Why not just check the final account details?
	//* If we only checked the final balance, we know the end state was correct, but wouldn't know if the transactions were executed properly along the way. Database might have race conditions that occasionally work but could fail under different timing conditions.

	for range n {
		err := <-errors
		require.NoError(t, err)

		result := <-results
		require.NotEmpty(t, result)

		//* check transfer
		transfer := result.Transfer
		require.NotEmpty(t, transfer)
		require.Equal(t, transfer.FromAccountID, account1.ID)
		require.Equal(t, transfer.ToAccountID, account2.ID)
		require.Equal(t, transfer.Amount, amount)

		//* check auto initializing fields too
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)

		//* check if transfer record is actually created in the database
		_, err = store.GetTranfer(context.Background(), transfer.ID)

		require.NoError(t, err)

		//* check entries
		fromEntry := result.FromEntry
		require.NotEmpty(t, fromEntry)
		require.Equal(t, fromEntry.AccountID, account1.ID)
		require.Equal(t, fromEntry.Amount, -amount)
		require.NotZero(t, fromEntry.ID)

		//* check if entry record is actually created in the database
		_, err = store.GetEntry(context.Background(), fromEntry.ID)
		require.NoError(t, err)

		ToEntry := result.ToEntry
		require.NotEmpty(t, ToEntry)
		require.Equal(t, ToEntry.AccountID, account2.ID)
		require.Equal(t, ToEntry.Amount, amount)
		require.NotZero(t, ToEntry.ID)

		//* check if entry record is actually created in the database
		_, err = store.GetEntry(context.Background(), ToEntry.ID)
		require.NoError(t, err)

		//* check sender account
		fromAccount := result.FromAccount
		require.NotEmpty(t, fromAccount)
		require.Equal(t, fromAccount.ID, account1.ID)

		//* check receiver account
		toAccount := result.ToAccount
		require.NotEmpty(t, toAccount)
		require.Equal(t, toAccount.ID, account2.ID)

		//* check account balances
		//? compare current balance compared to intial balance before transaction for both accounts
		moneyRemoved := account1.Balance - fromAccount.Balance
		moneyAdded := toAccount.Balance - account2.Balance
		// the total amount lost by sender so far should be gained by receiver
		require.Equal(t, moneyRemoved, moneyAdded)
		// extra check: cannot lose money less than the amount per transaction
		require.True(t, moneyRemoved >= amount)
		//? IMP: ensure that were no fatal race conditions after the 5 concurrent operations that lead to an out of sync data manipulation
		//* after each goroutine, the total amount deducted multiples: 1* amount, 2* amount, 3*amount
		//* hence, the final balance after a goroutine should be a multiple of the amount deducted
		require.True(t, moneyRemoved%amount == 0)

		quotient := int(moneyRemoved / amount)
		require.True(t, quotient >= 1 && quotient <= n)
		require.NotContains(t, quotientMap, quotient)
		quotientMap[quotient] = true
	}

	//? check the final updated balances
	updatedAccount1, err := testQueries.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)
	updatedAccount2, err := testQueries.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)

	require.Equal(t, account1.Balance-(int64(n)*int64(amount)), updatedAccount1.Balance)
	require.Equal(t, account2.Balance+(int64(n)*int64(amount)), updatedAccount2.Balance)
}
