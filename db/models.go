package db

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID              primitive.ObjectID `bson:"_id"`
	Email           string             `bson:"email"`
	FirstName       string             `bson:"firstName"`
	LastName        string             `bson:"lastName"`
	OnboardingStep  int                `bson:"onboardingStep"`
	EmailVerified   bool               `bson:"verified"`
	TermsAccepted   bool               `bson:"termsAccepted"`
	TermsAcceptedAt primitive.DateTime `bson:"termsAcceptedAt"`
	CreatedAt       string             `bson:"createdAt"`
}

type VerificationCode struct {
	ID    primitive.ObjectID `bson:"_id"`
	Email string             `bson:"email"`
	Code  string             `bson:"code"`
}

type App struct {
	ID          primitive.ObjectID `bson:"_id"`
	Name        string             `bson:"name"`
	Description string             `bson:"description"`
	Screens     []string           `bson:"screens"`
	CreatedAt   string             `bson:"createdAt"`
}
