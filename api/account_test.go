package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	mockdb "github.com/VihangaFTW/Go-Backend/db/mock"
	db "github.com/VihangaFTW/Go-Backend/db/sqlc"
	"github.com/VihangaFTW/Go-Backend/db/util"
	"github.com/VihangaFTW/Go-Backend/token"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestGetAccountAPI(t *testing.T) {

	//?    Flow:
	//? 1. create a fake account.
	//? 2. configure fake db to send the appropriate response back when the api endpoint is called with the account id as parameter (setup an "expectation")
	//? 3. start the server and send the get request for the account
	//? 4. validate response content

	user, _ := randomUser(t)
	account := randomAccount(user.Username)

	// define mock test case structure
	type testCase struct {
		name          string
		accountID     int64
		buildStubs    func(store *mockdb.MockStore)
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}

	testcases := []testCase{
		{
			name:      "OK",
			accountID: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				//? define expectations
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(account, nil)
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationHeadTypeBearer, user.Username, time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check response code
				require.Equal(t, http.StatusOK, recorder.Code)
				// check response body
				requireBodyMatchAccount(t, recorder.Body, account)
			},
		},
		{
			name:      "NotFound",
			accountID: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				//? define expectation
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(db.Account{}, sql.ErrNoRows)
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationHeadTypeBearer, user.Username, time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check response code
				require.Equal(t, http.StatusNotFound, recorder.Code)
				// no account to check for content

			},
		},
		{
			name:      "InternalError",
			accountID: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				//? define expectation
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(db.Account{}, sql.ErrConnDone)
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationHeadTypeBearer, user.Username, time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check response code
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
				// no account to check for content

			},
		},
		{
			name:      "BadRequest",
			accountID: 0,
			buildStubs: func(store *mockdb.MockStore) {
				//? define expectation
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationHeadTypeBearer, user.Username, time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check response code
				require.Equal(t, http.StatusBadRequest, recorder.Code)
				// no account to check for content

			},
		},
		{
			name:      "UnauthorizedUser",
			accountID: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				//? define expectations
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(account, nil)
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationHeadTypeBearer, "unauthorized_user", time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check response code
				require.Equal(t, http.StatusUnauthorized, recorder.Code)

			},
		},
		{
			name:      "NoAuthorization",
			accountID: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				//? define expectations
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				//? no auth header in request
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check response code
				require.Equal(t, http.StatusUnauthorized, recorder.Code)

			},
		},
	}

	//? loop through the test cases
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)

			// verify that all the expectations were met at the end of the test
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)

			// build stub
			tc.buildStubs(store)

			// start test server and send request to mock db
			server := newTestServer(t, store)

			//! store the response from the server so we can validate its content later on
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/accounts/%d", tc.accountID)

			// send the request for the account
			request, err := http.NewRequest(http.MethodGet, url, nil)

			//make sure the server does not return an error
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			//? 1. router receives the get request
			//? 2. the getAccount handler function runs
			//? 3. calls store.GetAccount(ctx, 123)
			//? 4. mock responds by returning the fake account data as defined in the expectation
			//? 5. the handler responses by sending JSON response back
			//? 6. the recorder captures all response data (status, headers, body)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)

		})
	}

}

func randomAccount(owner string) db.Account {
	return db.Account{
		ID:       util.RandomAccountId(),
		Owner:    owner,
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}
}

func requireBodyMatchAccount(t *testing.T, body *bytes.Buffer, account db.Account) {

	data, err := io.ReadAll(body)
	require.NoError(t, err)
	var gotAccount db.Account

	err = json.Unmarshal(data, &gotAccount)
	require.NoError(t, err)

	require.Equal(t, account, gotAccount)
}
