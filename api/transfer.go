package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	db "github.com/VihangaFTW/Go-Backend/db/sqlc"
	"github.com/VihangaFTW/Go-Backend/token"
	"github.com/gin-gonic/gin"
)

// transferRequest defines the expected JSON structure for transfer creation requests
// This struct is used to validate and bind incoming HTTP request data
type transferRequest struct {
	FromAccountID int64  `json:"from_account_id" binding:"required,min=1"` // Source account ID (must be positive)
	ToAccountID   int64  `json:"to_account_id" binding:"required,min=1"`   // Destination account ID (must be positive)
	Amount        int64  `json:"amount" binding:"required,gt=0"`           // Transfer amount (must be greater than 0)
	Currency      string `json:"currency" binding:"required,currency"`     // Currency code (validated by custom currency validator)
}

// createTransfer handles POST /transfers requests to create money transfers between accounts
// This endpoint requires authentication (protected by authMiddleware)
// It validates both accounts exist, have matching currencies, and creates the transfer transaction
func (server *Server) createTransfer(ctx *gin.Context) {
	var req transferRequest

	// Parse and validate the JSON request body against the transferRequest struct
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	fromAccount, valid := server.validAccount(ctx, req.FromAccountID, req.Currency)

	if !valid {
		return
	}

	// Extract the authenticated user's payload from the authorization middleware
	// This contains the username and other claims from the validated JWT token
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	// Authorization check: Verify that the authenticated user owns the source account
	// This prevents users from transferring money from accounts they don't own
	// Only the account owner can initiate transfers from their account
	if authPayload.Username != fromAccount.Owner {
		err := errors.New("from account does not belong to authenticated user")
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	_, valid = server.validAccount(ctx, req.ToAccountID, req.Currency)

	if !valid {
		return
	}
	// Prepare parameters for the database transfer transaction
	arg := db.CreateTranferParams{
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
	}

	// Execute the transfer transaction in the database
	// This will create transfer, entry, and account balance update records atomically
	transfer, err := server.store.CreateTranfer(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Return the created transfer details
	ctx.JSON(http.StatusOK, transfer)
}

// validAccount is a helper function that validates an account for transfer operations
// It checks if the account exists and has the expected currency
// Parameters:
//   - ctx: Gin context for sending error responses
//   - accountID: ID of the account to validate
//   - currency: Expected currency code for the account
//
// Returns: true if account is valid, false if invalid (with error response sent)
func (server *Server) validAccount(ctx *gin.Context, accountID int64, currency string) (db.Account, bool) {
	// Attempt to retrieve the account from the database
	account, err := server.store.GetAccount(ctx, accountID)
	if err != nil {
		// Handle case where account doesn't exist
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return account, false
		}

		// Handle other database errors (connection issues, etc.)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return account, false
	}

	// Verify that the account's currency matches the expected currency
	// This prevents transfers between accounts with different currencies
	if account.Currency != currency {
		err := fmt.Errorf("account [%d] currency mismatch: %s vs %s", accountID, account.Currency, currency)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return account, false
	}

	// Account exists and has the correct currency
	return account, true
}
