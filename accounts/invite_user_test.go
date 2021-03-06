package accounts_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/RichardKnop/example-api/accounts"
	"github.com/RichardKnop/example-api/models"
	"github.com/RichardKnop/example-api/oauth/roles"
	"github.com/RichardKnop/example-api/test-util"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func (suite *AccountsTestSuite) TestInviteUserRequiresUserAuthentication() {
	testutil.TestPostErrorExpectedResponse(
		suite.T(),
		suite.router,
		"http://1.2.3.4/v1/invitations",
		"invite_user",
		nil,
		"", // no access token
		accounts.ErrUserAuthenticationRequired.Error(),
		http.StatusUnauthorized,
		suite.assertMockExpectations,
	)
}

func (suite *AccountsTestSuite) TestInviteUser() {
	// Prepare a request
	payload, err := json.Marshal(&accounts.InvitationRequest{
		Email: "john@reese",
	})
	assert.NoError(suite.T(), err, "JSON marshalling failed")
	r, err := http.NewRequest(
		"POST",
		"http://1.2.3.4/v1/invitations",
		bytes.NewBuffer(payload),
	)
	assert.NoError(suite.T(), err, "Request setup should not get an error")
	r.Header.Set("Authorization", "Bearer test_user_token")

	suite.service.WaitForNotifications(1)
	// Mock invitation email
	suite.mockInvitationEmail()

	// Check the routing
	match := new(mux.RouteMatch)
	suite.router.Match(r, match)
	if assert.NotNil(suite.T(), match.Route) {
		assert.Equal(suite.T(), "invite_user", match.Route.GetName())
	}

	// Count before
	var (
		countBefore            int
		invitationsCountBefore int
	)
	suite.db.Model(new(models.User)).Count(&countBefore)
	suite.db.Model(new(models.Invitation)).Count(&invitationsCountBefore)

	// And serve the request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, r)

	// Count after
	var (
		countAfter            int
		invitationsCountAfter int
	)
	suite.db.Model(new(models.User)).Count(&countAfter)
	suite.db.Model(new(models.Invitation)).Count(&invitationsCountAfter)
	assert.Equal(suite.T(), countBefore+1, countAfter)
	assert.Equal(suite.T(), invitationsCountBefore+1, invitationsCountAfter)

	// Fetch the created invitation
	invitation := new(models.Invitation)
	assert.False(suite.T(), models.InvitationPreload(suite.db).
		Last(invitation).RecordNotFound())

	// And correct data was saved
	assert.Equal(suite.T(), invitation.InvitedUser.ID, invitation.InvitedUser.OauthUser.MetaUserID)
	assert.Equal(suite.T(), "john@reese", invitation.InvitedUser.OauthUser.Username)
	assert.False(suite.T(), invitation.InvitedUser.OauthUser.Password.Valid)
	assert.Equal(suite.T(), roles.User, invitation.InvitedUser.OauthUser.RoleID.String)
	assert.Equal(suite.T(), "test@user", invitation.InvitedByUser.OauthUser.Username)

	// Check the response
	expected, err := accounts.NewInvitationResponse(invitation)
	assert.NoError(suite.T(), err, "Failed to create expected response object")
	testutil.TestResponseObject(suite.T(), w, expected, 201)

	// block until email goroutine has finished
	assert.True(suite.T(), <-suite.service.GetNotifications(), "The email goroutine should have run")

	// Check that the mock object expectations were met
	suite.assertMockExpectations()
}
