package accounts

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/RichardKnop/example-api/models"
	"github.com/RichardKnop/example-api/oauth/roles"
	"github.com/RichardKnop/example-api/util/response"
	"github.com/gorilla/mux"
)

var (
	// ErrGetUserPermission ...
	ErrGetUserPermission = errors.New("Need permission to get user")
)

// Handles requests to get a user
// GET /v1/users/{id:[0-9]+}
func (s *Service) getUserHandler(w http.ResponseWriter, r *http.Request) {
	// Get the authenticated user from the request context
	authenticatedUser, err := GetAuthenticatedUser(r)
	if err != nil {
		response.UnauthorizedError(w, err.Error())
		return
	}

	// Get the id from request URI and type assert it
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["id"])
	if err != nil {
		response.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Fetch the user we want to update
	user, err := s.FindUserByID(uint(userID))
	if err != nil {
		response.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Check permissions
	if err = checkGetUserPermissions(authenticatedUser, user); err != nil {
		response.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	// Create response
	userResponse, err := NewUserResponse(user)
	if err != nil {
		response.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Write JSON response
	response.WriteJSON(w, userResponse, http.StatusOK)
}

func checkGetUserPermissions(authenticatedUser, user *models.User) error {
	// Superusers can get any users
	if authenticatedUser.OauthUser.RoleID.String == roles.Superuser {
		return nil
	}

	// Users can get their own account
	if authenticatedUser.ID == user.ID {
		return nil
	}

	// The user doesn't have the permission
	return ErrGetUserPermission
}
