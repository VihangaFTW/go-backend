package db

import (
	"context"
	"testing"

	"github.com/VihangaFTW/Go-Backend/util"
	"github.com/stretchr/testify/require"
)

func createRandomTransfer(t *testing.T) {

	// create two account ids to not violate foreign key constraints
	sender_acc := createRandomAccount(t)
	receiver_acc := createRandomAccount(t)

	arg := CreateTranferParams{
		FromAccountID: sender_acc.ID,
		ToAccountID:   receiver_acc.ID,
		Amount:        util.RandomAmount(),
	}

	transfer, err := testQueries.CreateTranfer(context.Background(),arg)

	require.NoError(t, err)
	require.NotEmpty(t, transfer)

	// check for account ids
	require.Equal(t, transfer.FromAccountID, sender_acc.ID)
	require.Equal(t, transfer.ToAccountID, receiver_acc.ID)

	// check for amount
	require.Equal(t, transfer.Amount, arg.Amount)

	// check for autogen transfer id
	require.NotEmpty(t, transfer.ID)
}


func TestCreateTransfer(t *testing.T)  {
	createRandomTransfer(t)
}

func TestGetTransfer(t *testing.T){
	
}
