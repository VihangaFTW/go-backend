package db

import (
	"context"
	"database/sql"
)

// VerifyEmailTxParams contains the parameters needed to verify an email address.
type VerifyEmailTxParams struct {
	EmailId    int64
	SecretCode string
}

// VerifyEmailTxResult contains the updated user and verify email records after successful verification.
type VerifyEmailTxResult struct {
	User        User
	VerifyEmail VerifyEmail
}

// VerifyEmailTx marks a verify email record as used and updates the user's email verification status
// within a single database transaction. If any step fails, the entire transaction is rolled back.
func (store *SQLStore) VerifyEmailTx(ctx context.Context, arg VerifyEmailTxParams) (VerifyEmailTxResult, error) {
	
	var result VerifyEmailTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		// Keep track of latest error.
		var err error

		// Mark the verify email record as used within transaction.
		result.VerifyEmail, err = q.UpdateVerifyEmail(ctx, UpdateVerifyEmailParams{
			ID:         arg.EmailId,
			SecretCode: arg.SecretCode,
		})

		if err != nil {
			return err
		}

		// Update user's email verification status within same transaction.
		result.User, err = q.UpdateUser(ctx, UpdateUserParams{
			Username: result.VerifyEmail.Username,
			IsEmailVerified: sql.NullBool{
				Bool:  true,
				Valid: true,
			},
		})

		return err
	})

	return result, err
}
