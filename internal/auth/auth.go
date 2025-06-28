package auth

import (
    "errors"
    "net/http"
    "strings"
    "time"
    "github.com/google/uuid"
    //"github.com/golang-jwt/jwt"
    "github.com/golang-jwt/jwt/v5"
    "golang.org/x/crypto/bcrypt"
)

// CH6 L01 https://www.boot.dev/lessons/294e5c16-d1e8-4836-871c-dedc98581236
// CH6 L06 https://www.boot.dev/lessons/be93db0d-4c6d-49cf-b56d-ba22392eb160
// CH6 L07 https://www.boot.dev/lessons/0689e0d0-bdb1-4cc8-b577-f0dd0535ad00

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

// CH6 L07
// Bearer TOKEN_STRING
// This function should look for the Authorization header in the headers parameter
// and return the TOKEN_STRING if it exists (stripping off the Bearer prefix and whitespace).
// If the header doesn't exist, return an error.
func GetBearerToken(headers http.Header) (string, error) {
    authorization := headers.Get("Authorization")
    if authorization == "" {
        return "", errors.New("no authorization header")
    }

    sep := strings.Split(authorization, " ")
    if len(sep) != 2 {
        return "", errors.New("wrong format in authorization header")
    }
    
    if sep[0] != "Bearer" {
        return "", errors.New("wrong format in authorization header")
    }

    return sep[1], nil
}