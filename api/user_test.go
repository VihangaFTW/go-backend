package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	mockdb "github.com/VihangaFTW/Go-Backend/db/mock"
	db "github.com/VihangaFTW/Go-Backend/db/sqlc"
	"github.com/VihangaFTW/Go-Backend/util"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// eqCreateUserParamsMatcher is a custom matcher for validating whether a given password matches
// its hashed equivalent stored in the db.
// When testing user creation, the password gets
// hashed before being stored, so we can't directly compare the expected parameters with the actual
// parameters because the hashed password will be different each time (due to bcrypt's random salt).
type eqCreateUserParamsMatcher struct {
	arg      db.CreateUserParams
	password string
}

// Mathches returns whether x is a match. This method must be implemented to satisfy the Matcher interface.
func (e eqCreateUserParamsMatcher) Matches(x any) bool {
	//? make sure the input is a CreateUserParams type
	actual, ok := x.(db.CreateUserParams)

	// input is not expected type: match failed
	if !ok {
		return false
	}

	//! use bcrypt to verify the password matches its hash instead of directly comparing two unidentical hashes
	if err := util.CheckPassword(e.password, actual.HashedPassword); err != nil {
		return false
	}

	// sets the expected struct's HashedPassword field to match the actual hashed password
	e.arg.HashedPassword = actual.HashedPassword
	// compare all other fields (username, email, full name)
	return reflect.DeepEqual(e.arg, actual)
}

// String describes what the Matcher matches. This method must be implemented to satisfy the Matcher interface.
func (e eqCreateUserParamsMatcher) String() string {
	return fmt.Sprintf("matches arg %v and password %v", e.arg, e.password)
}

// wrapper function that returns a gomock.Matcher
func EqCreateUserParams(expected db.CreateUserParams, password string) gomock.Matcher {
	return eqCreateUserParamsMatcher{expected, password}
}

// Test Setup
//
// # EqCreateUserParams(arg, password) → Creates matcher → Stored in gomock
//
// # API Call
//
// Handler calls store.CreateUser(ctx, actualParams)
//
//	↓
//
// Gomock intercepts and checks matchers
//
//	↓
//
// matcher.Matches(actualParams) gets called automatically
//
//	↓
//
// Your custom logic validates password hash
//
//	↓
//
// Returns true/false to determine if expectation matches
func TestCreateUserAPI(t *testing.T) {
	// 1. create a random user
	user, password := randomUser(t)

	// 2. prepare testcases

	// testCase defines the shape of each testcase object
	type testCase struct {
		name string
		//* mocks the shape of the http request payload (as JSON data)
		body          gin.H
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}

	testcases := []testCase{
		{
			name: "OK",
			body: gin.H{
				"username":  user.Username,
				"password":  password,
				"full_name": user.FullName,
				"email":     user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {

				// build the input object
				arg := db.CreateUserParams{
					Username: user.Username,
					FullName: user.FullName,
					Email:    user.Email,
				}
				//? EqCreateUserParams creates a Matcher object and stores it in gomock's expectations.
				//? When the API handler calls the store's CreateUser function, gomock intercepts this function call.
				//? Gomock runs the provided Matcher's Matches function and determines whether the expectation matches
				//? depending on the results of this function.
				store.EXPECT().CreateUser(gomock.Any(), EqCreateUserParams(arg, password)).Times(1).Return(user, nil)
								
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchUser(t, recorder.Body, user)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"username":  user.Username,
				"password":  password,
				"full_name": user.FullName,
				"email":     user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "DuplicateUsername",
			body: gin.H{
				"username":  user.Username,
				"password":  password,
				"full_name": user.FullName,
				"email":     user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, &pq.Error{Code: "23505"}) // code for unique_violation postgres error
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name: "InvalidUsername",
			body: gin.H{
				"username":  "invalid-user#1",
				"password":  password,
				"full_name": user.FullName,
				"email":     user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidEmail",
			body: gin.H{
				"username":  user.Username,
				"password":  password,
				"full_name": user.FullName,
				"email":     "invalid-email",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "TooShortPassword",
			body: gin.H{
				"username":  user.Username,
				"password":  "123",
				"full_name": user.FullName,
				"email":     user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	// 3. run all testcases
	for i := range testcases {
		tc := testcases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			//? check the request object and convert it to a JSON object to pass to the endpoint
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/users"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})

	}
}

// randomUser generates a db.User struct for testing the CreateUser API
func randomUser(t *testing.T) (user db.User, password string) {
	password = util.RandomString(6)

	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)

	user = db.User{
		Username:       util.RandomOwner(),
		Email:          util.RandomEmail(),
		FullName:       util.RandomOwner(),
		HashedPassword: hashedPassword,
	}
	return
}

func requireBodyMatchUser(t *testing.T, body *bytes.Buffer, user db.User) {

	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotUser db.User
	err = json.Unmarshal(data, &gotUser)

	require.NoError(t, err)
	require.Equal(t, user.FullName, gotUser.FullName)
	require.Equal(t, user.Email, gotUser.Email)
	require.Equal(t, user.Username, gotUser.Username)
	//? request should not return the password hash as we removed that field in the createUser api handler
	require.Empty(t, gotUser.HashedPassword)
}
