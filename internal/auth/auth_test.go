package auth

import (
    "testing"
    "time"
    "github.com/google/uuid"
	_ "github.com/golang-jwt/jwt/v5"
)

func TestMakeJWT(t *testing.T) {
	// Set up your test data
    userID := uuid.New()
	tokenSecret := "la gallina diu que no"
	expiresIn := time.Duration(1 * time.Hour)

	// Call the function you're testing
	jwtString, err := MakeJWT(userID, tokenSecret, expiresIn)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
		return
	}

	// Check if the result is what you expected
	// Check that we got a token
	if jwtString == "" {
		t.Errorf("Expected a JWT token, got empty string")
		return
	}

	// Validate the token we just created
	validatedUserID, err := ValidateJWT(jwtString, tokenSecret)
	if err != nil {
		t.Errorf("Expected no error validating JWT, got %v", err)
		return
	}

	// Check that the user ID matches
	if validatedUserID != userID {
		t.Errorf("Expected user ID %v, got %v", userID, validatedUserID)
	}
}

// TODO
func TestHeaderHasBearer(t *testing.T) {
}