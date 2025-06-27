package auth

import (
    "errors"
    "time"
    "github.com/google/uuid"
    //"github.com/golang-jwt/jwt"
    "github.com/golang-jwt/jwt/v5"
    "golang.org/x/crypto/bcrypt"
)

// CH6 L01
// Hash the password using the bcrypt.GenerateFromPassword function.
// Bcrypt is a secure hash function that is intended for use with passwords.
// https://gowebexamples.com/password-hashing/
func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
    return string(bytes), err
}

// CH6 L01
// Use the bcrypt.CompareHashAndPassword function to compare the password
// that the user entered in the HTTP request with the password that is stored in the database.
func CheckPasswordHash(password, hash string) error {
    return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// CH6 L06
func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
    // Create claims
    current_utc_time := time.Now().UTC()
    claims := jwt.RegisteredClaims{
        Issuer: "chirpy",
        IssuedAt: jwt.NewNumericDate(current_utc_time),
        ExpiresAt: jwt.NewNumericDate(current_utc_time.Add(expiresIn)),
        Subject: userID.String(),
    }

    // Create token
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

    // Sign token and return it
    return token.SignedString([]byte(tokenSecret))
}

// CH6 L06
func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
    token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{},
        func(token *jwt.Token) (interface{}, error) {
	        return []byte(tokenSecret), nil
        })

    if err != nil {
        return uuid.UUID{}, err
    }

    if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok {
        // Now you can access claims.Subject
        // Convert claims.Subject (string) back to uuid.UUID
        // Return the UUID
        return uuid.Parse(claims.Subject)
    } else {
        // Handle the case where claims aren't the expected type
        return uuid.UUID{}, errors.New("invalid claims")
    }
}