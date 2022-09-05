package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"http-server/db"
	"http-server/lib"
	"http-server/middleware"
	"net/http"
	"time"

	"github.com/go-chi/jwtauth/v5"
	"github.com/thanhpk/randstr"
	"go.mongodb.org/mongo-driver/bson"
)

type RegisterRequest struct {
	Email string `json:"email"`
}

/**
 * POST - /auth/register
 *
 * Request Body:
 * type RegisterRequest struct {
 * 	Email string `json:"email"`
 * }
 */
func Register(res http.ResponseWriter, req *http.Request) {
	var registerRequest RegisterRequest

	// Decode the request body into the RegisterRequest struct
	json.NewDecoder(req.Body).Decode(&registerRequest)

	// Make sure an email is provided
	if registerRequest.Email == "" {
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte("Missing field(s): email"))
		return
	}

	var result bson.M
	// If there is no error, the user already exists
	if err := db.Users.FindOne(context.TODO(), bson.D{{Key: "email", Value: registerRequest.Email}}).Decode(&result); err == nil {
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte("Email already exists"))
		return
	}

	// Create a new user
	userDocument := bson.D{
		{Key: "email", Value: registerRequest.Email},
		{Key: "verified", Value: false},
		{Key: "firstName", Value: nil},
		{Key: "lastName", Value: nil},
		{Key: "onboardingStep", Value: 1},
		{Key: "termsAccepted", Value: false},
		{Key: "termsAcceptedAt", Value: nil},
		{Key: "createdAt", Value: time.Now()},
	}

	verificationCode := randstr.String(6)

	// Create new verification code
	verificationCodeDocument := bson.D{
		{Key: "email", Value: registerRequest.Email},
		{Key: "code", Value: verificationCode},
	}

	_, err := db.Users.InsertOne(context.TODO(), userDocument)

	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte("Error creating user"))
		return
	}

	_, err = db.VerificationCodes.InsertOne(context.TODO(), verificationCodeDocument)

	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte("Error creating verification code"))
		return
	}

	// Send an email with a verification code.
	err = lib.SendMail(registerRequest.Email, "Verify your email", fmt.Sprintf("Your verification code is %s", verificationCode))

	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte("Error sending verification email"))
		return
	}

	res.WriteHeader(http.StatusOK)
	res.Write([]byte("Check email for verification code"))
}

type RegisterVerifyRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

/**
 * POST - /auth/register/verify
 *
 * Request Body:
 * type RegisterVerifyRequest struct {
 *	 Email string `json:"email"`
 *	 Code  string `json:"code"`
 * }
 */
func RegisterVerify(res http.ResponseWriter, req *http.Request) {
	var registerVerifyRequest RegisterVerifyRequest

	// Decode the request body into the RegisterRequest struct
	json.NewDecoder(req.Body).Decode(&registerVerifyRequest)

	// Make sure an email is provided
	if registerVerifyRequest.Email == "" {
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte("Missing field(s): email"))
		return
	}

	// Make sure a code is provided
	if registerVerifyRequest.Code == "" {
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte("Missing field(s): code"))
		return
	}

	// Find and delete the verification code
	deleteResult, err := db.VerificationCodes.DeleteOne(
		context.TODO(),
		bson.D{
			{Key: "email", Value: registerVerifyRequest.Email},
			{Key: "code", Value: registerVerifyRequest.Code},
		},
	)

	// If there is an error, the code could not be found so cannot be deleted
	if err != nil || deleteResult.DeletedCount == 0 {
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte("Invalid code"))
		return
	}

	// Update the user to be verified, and move on to the next onboarding step
	filter := bson.D{{Key: "email", Value: registerVerifyRequest.Email}}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "verified", Value: true}, {Key: "onboardingStep", Value: 2}}}}

	_, err = db.Users.UpdateOne(context.TODO(), filter, update)

	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte("Error updating user"))
		return
	}

	// Get the updated user
	var user bson.M
	err = db.Users.FindOne(context.TODO(), bson.D{{Key: "email", Value: registerVerifyRequest.Email}}).Decode(&user)

	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte("Error getting user"))
		return
	}

	// Create a new JWT
	token, err := middleware.CreateToken(user)

	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte("Error creating token"))
		return
	}

	// Send the JWT back to the client
	res.WriteHeader(http.StatusOK)
	res.Write([]byte(token))
}

type RegisterProfileRequest struct {
	FirstName     string `json:"firstName"`
	LastName      string `json:"lastName"`
	TermsAccepted bool   `json:"termsAccepted"`
}

/**
 * POST - /auth/register/profile
 *
 * Request Body:
 * type RegisterProfileRequest struct {
 *	 FirstName     string `json:"firstName"`
 *	 LastName      string `json:"lastName"`
 *	 TermsAccepted bool   `json:"termsAccepted"`
 * }
 */

func RegisterProfile(res http.ResponseWriter, req *http.Request) {
	var registerProfileRequest RegisterProfileRequest

	// Decode the request body into the RegisterProfileRequest struct
	json.NewDecoder(req.Body).Decode(&registerProfileRequest)

	// Fetch the object from the JWT
	_, claims, _ := jwtauth.FromContext(req.Context())

	// If body has first and last name, then user is on step 2
	if registerProfileRequest.FirstName != "" && registerProfileRequest.LastName != "" {
		if int(claims["onboardingStep"].(float64)) != 2 {
			res.WriteHeader(http.StatusBadRequest)
			res.Write([]byte("Invalid onboarding step"))
			return
		}

		// Update the user to be verified, and move on to the next onboarding step
		filter := bson.D{{Key: "email", Value: claims["email"]}}
		update := bson.D{{Key: "$set", Value: bson.D{{Key: "firstName", Value: registerProfileRequest.FirstName}, {Key: "lastName", Value: registerProfileRequest.LastName}, {Key: "onboardingStep", Value: 3}}}}

		_, err := db.Users.UpdateOne(context.TODO(), filter, update)

		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			res.Write([]byte("Error updating user"))
			return
		}

		// Get the updated user
		var user bson.M
		err = db.Users.FindOne(context.TODO(), bson.D{{Key: "email", Value: claims["email"]}}).Decode(&user)

		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			res.Write([]byte("Error getting user"))
			return
		}

		// Create a new JWT
		token, err := middleware.CreateToken(user)

		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			res.Write([]byte("Error creating token"))
			return
		}

		// Send the JWT back to the client
		res.WriteHeader(http.StatusOK)
		res.Write([]byte(token))
		return

	} else if registerProfileRequest.TermsAccepted {
		if int(claims["onboardingStep"].(float64)) != 3 {
			res.WriteHeader(http.StatusBadRequest)
			res.Write([]byte("Invalid onboarding step"))
			return
		}

		// Update the user to be verified, and move on to the next onboarding step
		filter := bson.D{{Key: "email", Value: claims["email"]}}
		update := bson.D{{Key: "$set", Value: bson.D{{Key: "termsAccepted", Value: registerProfileRequest.TermsAccepted}, {Key: "onboardingStep", Value: 4}, {Key: "termsAcceptedAt", Value: time.Now()}}}}

		_, err := db.Users.UpdateOne(context.TODO(), filter, update)

		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			res.Write([]byte("Error updating user"))
			return
		}

		// Get the updated user
		var user bson.M
		err = db.Users.FindOne(context.TODO(), bson.D{{Key: "email", Value: claims["email"]}}).Decode(&user)

		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			res.Write([]byte("Error getting user"))
			return
		}

		// Create a new JWT
		token, err := middleware.CreateToken(user)

		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			res.Write([]byte("Error creating token"))
			return
		}

		// Send the JWT back to the client
		res.WriteHeader(http.StatusOK)
		res.Write([]byte(token))
		return
	} else {
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte("Missing field(s): firstName, lastName, termsAccepted"))
		return
	}
}

func Login(res http.ResponseWriter, req *http.Request) {

}
