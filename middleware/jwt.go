package middleware

import (
	"github.com/go-chi/jwtauth/v5"
	"go.mongodb.org/mongo-driver/bson"
)

var TokenAuth *jwtauth.JWTAuth

func CreateToken(user bson.M) (string, error) {
	_, tokenString, _ := TokenAuth.Encode(map[string]interface{}{
		"id":              user["_id"],
		"email":           user["email"],
		"firstName":       user["firstName"],
		"lastName":        user["lastName"],
		"onboardingStep":  user["onboardingStep"],
		"verified":        user["verified"],
		"termsAccepted":   user["termsAccepted"],
		"termsAcceptedAt": user["termsAcceptedAt"],
		"createdAt":       user["createdAt"],
	})

	return tokenString, nil
}
