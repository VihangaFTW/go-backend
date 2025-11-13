package db

import "context"

type CreateUserTxParams struct {
	CreateUserParams
	// AfterCreateUser callback executes post-creation logic (e.g., send verify email task scheduling)
	// within the same transaction, ensuring atomicity.
	AfterCreateUser func(user User) error
}

type CreateUserTxResult struct {
	User User
}

// CreateUserTx creates a user and executes post-creation logic (e.g., email verification task scheduling)
// within a single database transaction. If any step fails, the entire transaction is rolled back.
func (store *SQLStore) CreateUserTx(ctx context.Context, arg CreateUserTxParams) (CreateUserTxResult, error) {

	var result CreateUserTxResult

	err := store.execTx(ctx, func(q *Queries) error {

		//* keep track of latest error
		var err error

		// Create user within transaction.
		result.User, err = q.CreateUser(ctx, arg.CreateUserParams)

		if err != nil {
			return err
		}

		// Execute post-creation hook (e.g., schedule verify email task) within same transaction.
		return arg.AfterCreateUser(result.User)

	})

	return result, err
}
