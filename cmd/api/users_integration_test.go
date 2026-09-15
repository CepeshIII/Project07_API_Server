package main

// func TestGetUser(t *testing.T) {
// 	app := newTestApplication(t)
// 	mux := app.mount()

// 	user := store.User{
// 		ID:       1,
// 		Username: "username",
// 		Email:    "email",
// 		IsActive: false,
// 		RoleID:   0,
// 	}

// 	userWithRole := &store.UserWithRole{
// 		User: user,
// 	}

// 	app.store.Users.CreateAndInvite(context.Background(), userWithRole, "token", time.Duration(time.Hour))

// 	t.Run("should not allow unauthenticated requests", func(t *testing.T) {
// 		// check for 401 code
// 		req, err := http.NewRequest(http.MethodGet, "/v1/users/1", nil)
// 		// req.AddCookie(&cookie)

// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		rr := executeRequest(req, mux)

// 		checkResponseCode(t, http.StatusUnauthorized, rr.Code)
// 	})

// 	t.Run("should allow authenticated requests", func(t *testing.T) {
// 		// check for 200 code
// 		req, err := http.NewRequest(http.MethodGet, "/v1/users/1", nil)
// 		if err != nil {
// 			t.Fatal(err)
// 		}

// 		// create new Access Token
// 		accesstoken, accessTokenExp, err := app.generateAccessToken(1, "user")
// 		if err != nil {
// 			t.Fatal(err)
// 			return
// 		}

// 		cookie := http.Cookie{
// 			Name:     accessCookieName,
// 			Value:    accesstoken,
// 			Path:     "/",
// 			HttpOnly: true,
// 			SameSite: http.SameSiteLaxMode,
// 			Expires:  accessTokenExp,
// 		}
// 		req.AddCookie(&cookie)

// 		rr := executeRequest(req, mux)
// 		checkResponseCode(t, http.StatusOK, rr.Code)
// 	})

// 	t.Run("should return 404 when user does not exist", func(t *testing.T) {
// 		// check for 404 code
// 		req, err := http.NewRequest(http.MethodGet, "/v1/users/2", nil)
// 		if err != nil {
// 			t.Fatal(err)
// 		}

// 		// create new Access Token
// 		accesstoken, accessTokenExp, err := app.generateAccessToken(2, "user")
// 		if err != nil {
// 			t.Fatal(err)
// 			return
// 		}

// 		cookie := http.Cookie{
// 			Name:     accessCookieName,
// 			Value:    accesstoken,
// 			Path:     "/",
// 			HttpOnly: true,
// 			SameSite: http.SameSiteLaxMode,
// 			Expires:  accessTokenExp,
// 		}
// 		req.AddCookie(&cookie)

// 		rr := executeRequest(req, mux)
// 		checkResponseCode(t, http.StatusNotFound, rr.Code)
// 	})

// 	t.Run("should return 200 when user exist", func(t *testing.T) {
// 		// check for 200 code
// 		req, err := http.NewRequest(http.MethodGet, "/v1/users/1", nil)
// 		if err != nil {
// 			t.Fatal(err)
// 		}

// 		// create new Access Token
// 		accesstoken, accessTokenExp, err := app.generateAccessToken(1, "user")
// 		if err != nil {
// 			t.Fatal(err)
// 			return
// 		}

// 		cookie := http.Cookie{
// 			Name:     accessCookieName,
// 			Value:    accesstoken,
// 			Path:     "/",
// 			HttpOnly: true,
// 			SameSite: http.SameSiteLaxMode,
// 			Expires:  accessTokenExp,
// 		}
// 		req.AddCookie(&cookie)

// 		rr := executeRequest(req, mux)
// 		checkResponseCode(t, http.StatusOK, rr.Code)
// 	})

// }
