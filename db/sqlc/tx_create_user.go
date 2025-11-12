package db

import "context"

type CreateUserTxParams struct {
	CreateUserParams
	AfterCreateUser func(user User) error
}

type CreateUserTxResult struct {
	User User
}

func (store *SQLStore) CreateUserTx(ctx context.Context, arg CreateUserTxParams) (CreateUserTxResult, error) {

	var result CreateUserTxResult

	err := store.execTx(ctx, func(q *Queries) error {

		//* keep track of latest error
		var err error

		result.User, err = q.CreateUser(ctx, arg.CreateUserParams)

		if err != nil {
			return err
		}

		return arg.AfterCreateUser(result.User)

	})

	return result, err
}
