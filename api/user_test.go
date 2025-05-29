package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	mockdb "github.com/VihangaFTW/Go-Backend/db/mock"
	db "github.com/VihangaFTW/Go-Backend/db/sqlc"
	"github.com/VihangaFTW/Go-Backend/db/util"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// eqCreateUserParamsMatcher is a custom matcher for validating whether a given password matches
// its hashed equivalent stored in the db.
// when testing user creation, the password gets
// hashed before being stored, so we can't directly compare the expected parameters with the actual
// parameters because the hashed password will be different each time (due to bcrypt's random salt).
type eqCreateUserParamsMatcher struct {
	arg      db.CreateUserParams
	password string
}

// Mathches returns whether x is a match.
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

// String describes what the Matcher matches.
func (e eqCreateUserParamsMatcher) String() string {
	return fmt.Sprintf("matches arg %v and password %v", e.arg, e.password)
}

func EqCreateUserParams(expected db.CreateUserParams, password string) gomock.Matcher {
	return eqCreateUserParamsMatcher{expected, password}
}

func TestCreateUserAPI(t *testing.T) {
	// 1. create a random user
	user, password := randomUser(t)

	// 2. prepare testcases

	// testCase defines the shape of each testcase object
	type testCase struct {
		name          string
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
				arg := db.CreateUserParams{
					Username: user.Username,
					FullName: user.FullName,
					Email:    user.Email,
				}

				store.EXPECT().CreateUser(gomock.Any(), arg).Times(1).Return(user, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchUser(t, recorder.Body, user)
			},
		},
	}


	// 3. run all testcases
	for 
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
