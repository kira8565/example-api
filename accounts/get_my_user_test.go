package accounts

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/RichardKnop/jsonhal"
	"github.com/RichardKnop/recall/accounts/roles"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func (suite *AccountsTestSuite) TestGetMyUser() {
	// Prepare a request
	r, err := http.NewRequest("GET", "http://1.2.3.4/v1/accounts/users/me", nil)
	if err != nil {
		log.Fatal(err)
	}
	r.Header.Set("Authorization", "Bearer test_user_token")

	// Check the routing
	match := new(mux.RouteMatch)
	suite.router.Match(r, match)
	if assert.NotNil(suite.T(), match.Route) {
		assert.Equal(suite.T(), "get_my_user", match.Route.GetName())
	}

	// And serve the request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, r)

	// Check the status code
	if !assert.Equal(suite.T(), 200, w.Code) {
		log.Print(w.Body.String())
	}

	// Check the response body
	expected := &UserResponse{
		Hal: jsonhal.Hal{
			Links: map[string]*jsonhal.Link{
				"self": &jsonhal.Link{
					Href: fmt.Sprintf("/v1/accounts/users/%d", suite.users[1].ID),
				},
			},
		},
		ID:        suite.users[1].ID,
		Email:     "test@user",
		FirstName: "test_first_name_2",
		LastName:  "test_last_name_2",
		Role:      roles.User,
		CreatedAt: suite.users[1].CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: suite.users[1].UpdatedAt.UTC().Format(time.RFC3339),
	}
	expectedJSON, err := json.Marshal(expected)
	if err != nil {
		log.Fatal(err)
	}
	assert.Equal(
		suite.T(),
		string(expectedJSON),
		strings.TrimRight(w.Body.String(), "\n"), // trim the trailing \n
	)
}
