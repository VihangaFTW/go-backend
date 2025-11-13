package db

import (
	"context"
	"database/sql"
	"testing"

	"github.com/VihangaFTW/Go-Backend/util"
	"github.com/stretchr/testify/require"
)

func createRandomEntry(t *testing.T) Entry {

	// create a common account that the entry depends on
	 account := createRandomAccount(t)


	arg := CreateEntryParams{
		AccountID: account.ID,
		Amount:    util.RandomAmount(),
	}

	entry, err := testQueries.CreateEntry(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, entry)
	require.Equal(t, entry.AccountID, arg.AccountID)
	require.Equal(t, entry.Amount, arg.Amount)

	require.NotZero(t, entry.ID)

	return entry
}

func TestCreateEntry(t *testing.T) {
	createRandomEntry(t)
}

func TestGetEntry(t *testing.T) {
	entry1 := createRandomEntry(t)
	entry2, err := testQueries.GetEntry(context.Background(), entry1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, entry2)
	require.Equal(t, entry1.ID, entry2.ID)
	require.Equal(t, entry1.AccountID, entry2.AccountID)
	require.Equal(t, entry1.Amount, entry2.Amount)


	// test for entry that doesnt exist
	entry3, err2 := testQueries.GetEntry(context.Background(), 0)
	require.Error(t, err2)
	require.Empty(t, entry3)
	require.ErrorIs(t, err2, sql.ErrNoRows)
}

func TestListEntries(t *testing.T) {
	
	for range 10 {
		createRandomEntry(t)
	}
	// create an account
	account := createRandomAccount(t)

	// add 10 entries related to that account
	for range 10 {
		testQueries.CreateEntry(context.Background(), CreateEntryParams{
			AccountID: account.ID,
			Amount:    util.RandomAmount(),
		})
	}

	// we should get 10 entries, list the second page of 5 entries
	arg := ListEntriesParams{
		AccountID: account.ID,
		Limit:     5,
		Offset:    5,
	}

	entries, err := testQueries.ListEntries(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, entries)

	require.Len(t, entries, 5)

	// verify each entry belongs to the same account id
	for _, entry := range entries {
		require.Equal(t, entry.AccountID, account.ID)
	}
}
