package db

import (
	"context"
	"database/sql"
	"testing"

	"github.com/VihangaFTW/Go-Backend/util"
	"github.com/stretchr/testify/require"
)

// createRandomAccount is a test helper function that creates a random account in the database
// This function is used by multiple test cases to set up test data
// Returns: A newly created Account with random but valid data
func createRandomAccount(t *testing.T) Account {
	//* handle owner_account foregin key constraint after adding users table
	// First create a random user since accounts must have a valid owner (foreign key constraint)
	user := createRandomUser(t)

	//? test input - prepare parameters for account creation
	arg := CreateAccountParams{
		Owner:    user.Username,         // Use the username from the created user
		Balance:  util.RandomMoney(),    // Generate random balance amount
		Currency: util.RandomCurrency(), // Generate random currency (USD, EUR, CAD, etc.)
	}

	// Call the database function to create the account
	account, err := testQueries.CreateAccount(context.Background(), arg)
	//? automatically fails the test case if there error != nil
	require.NoError(t, err)
	// function should return an Account object as defined by our sql rules
	require.NotEmpty(t, account)

	// Verify that the data written to database matches our input parameters
	require.Equal(t, arg.Owner, account.Owner)
	require.Equal(t, arg.Balance, account.Balance)
	require.Equal(t, arg.Currency, account.Currency)

	// Verify that auto-generated fields are properly set by the database
	require.NotZero(t, account.ID)        // ID should be auto-generated and non-zero
	require.NotZero(t, account.CreatedAt) // CreatedAt timestamp should be auto-generated

	return account
}

// TestCreateAccount tests the basic account creation functionality
// This test verifies that we can successfully create an account with valid data
func TestCreateAccount(t *testing.T) {
	// Simply call createRandomAccount which already contains all the necessary assertions
	createRandomAccount(t)
}

// TestGetAccount tests the account retrieval functionality
// This test verifies that we can retrieve an account by ID and get the same data back
func TestGetAccount(t *testing.T) {
	// Create a test account first
	account1 := createRandomAccount(t)

	// Retrieve the account from database using its ID
	account2, err := testQueries.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, account2)

	// Verify that all fields from the retrieved account match the original account
	require.Equal(t, account1.ID, account2.ID)
	require.Equal(t, account1.Balance, account2.Balance)
	require.Equal(t, account1.Currency, account2.Currency)
	require.Equal(t, account1.Owner, account2.Owner)
	require.Equal(t, account1.CreatedAt, account2.CreatedAt)
}

// TestUpdateAccount tests the account balance update functionality
// This test verifies that we can update an account's balance and the change persists
func TestUpdateAccount(t *testing.T) {
	// Create a test account to update
	account1 := createRandomAccount(t)

	// Prepare update parameters with new random balance
	arg := UpdateAccountParams{
		ID:      account1.ID,        // Keep the same account ID
		Balance: util.RandomMoney(), // Generate new random balance
	}

	// Perform the update operation
	account2, err := testQueries.UpdateAccount(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, account2)

	// Verify that non-balance fields remain unchanged
	require.Equal(t, account1.ID, account2.ID)
	require.Equal(t, account1.Currency, account2.Currency)
	require.Equal(t, account1.Owner, account2.Owner)
	require.Equal(t, account1.CreatedAt, account2.CreatedAt)

	// Verify that the balance was updated to the new value
	require.Equal(t, arg.Balance, account2.Balance)
}

// TestDeleteAccount tests the account deletion functionality
// This test verifies that we can delete an account and it's no longer retrievable
func TestDeleteAccount(t *testing.T) {
	// Create a test account to delete
	account := createRandomAccount(t)

	// Delete the account
	err := testQueries.DeleteAccount(context.Background(), account.ID)
	require.NoError(t, err)

	// Try to retrieve the deleted account - this should fail
	account2, err := testQueries.GetAccount(context.Background(), account.ID)
	require.Error(t, err)                             // We expect an error because the account no longer exists
	require.EqualError(t, err, sql.ErrNoRows.Error()) // Specifically, it should be a "no rows" error
	require.Empty(t, account2)                        // The returned account should be empty
}

// TestListAccounts tests the account listing functionality with pagination
// This test verifies that we can list accounts for a specific owner with limit and offset
func TestListAccounts(t *testing.T) {
	var lastAccount Account
	// Create 10 random accounts for testing
	// We keep track of the last account to use its owner for filtering
	for range 10 {
		lastAccount = createRandomAccount(t)
	}

	// Set up parameters for listing accounts
	arg := ListAccountsParams{
		Owner:  lastAccount.Owner, // Filter by the owner of the last created account
		Offset: 0,                 // Start from the beginning (no offset)
		Limit:  5,                 // Limit results to 5 accounts
	}

	// Retrieve the list of accounts
	accounts, err := testQueries.ListAccounts(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, accounts)

	// Verify that all returned accounts belong to the specified owner
	for _, account := range accounts {
		require.NotEmpty(t, account)
		require.Equal(t, lastAccount.Owner, account.Owner)
	}
}
