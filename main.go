package main

import (
	"http-server/config"
	"http-server/db"
	"http-server/lib"
	jwt "http-server/middleware"
	"http-server/routes"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/go-chi/jwtauth/v5"
	"github.com/joho/godotenv"
)

func init() {
	// Load environment variables from .env file
	environmentError := godotenv.Load(".env")

	if environmentError != nil {
		log.Println(environmentError)
		log.Fatal("Error loading environment variables")
		return
	}

	// Connect to MongoDB
	databaseError := db.Connect()

	if databaseError != nil {
		log.Fatal("Error connecting to database")
		return
	}

	// Create AWS session
	session, awsSessionError := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
		// Credentials: credentials.NewStaticCredentials(config.GetConfig("AWS_SES_USERNAME"), config.GetConfig("AWS_SES_SECRET"), ""),
	})

	if awsSessionError != nil {
		log.Fatal("Error creating AWS session")
		return
	}

	// Set global AWS session variable
	lib.AwsSession = session
	lib.SesSession = ses.New(lib.AwsSession)

	secret := []byte(config.GetConfig("JWT_SECRET"))
	jwt.TokenAuth = jwtauth.New("HS256", secret, nil)
}

func main() {

	router := chi.NewRouter()

	router.Use(cors.Handler(cors.Options{
		AllowedMethods: []string{"GET", "POST"},
	}))

	router.Use(middleware.StripSlashes)
	router.Use(middleware.Heartbeat("/ping"))
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.NotFound(Err404)
	router.MethodNotAllowed(Err405)

	router.Group(func(router chi.Router) {
		// Auth routes all require rate limiting
		router.Use(httprate.Limit(
			5,             // Requests
			1*time.Minute, // Per Duration
			httprate.WithKeyFuncs(httprate.KeyByIP, httprate.KeyByEndpoint),
		))

		router.Post("/auth/login", routes.Login)
		router.Post("/auth/register", routes.Register)
		router.Post("/auth/register/verify", routes.RegisterVerify)

		router.Group(func(router chi.Router) {
			// Auth modification routes require authentication
			router.Use(jwtauth.Verifier(jwt.TokenAuth))
			router.Use(jwtauth.Authenticator)

			router.Post("/auth/register/profile", routes.RegisterProfile)
		})
	})

	router.Group(func(router chi.Router) {
		// r.use(middleware)
	})

	log.Println("Server started on port 8080")

	http.ListenAndServe("localhost:8080", router)

	log.Println("Server shutting down")
}

func Err404(res http.ResponseWriter, req *http.Request) {
	route := req.URL.Path

	res.WriteHeader(404)
	res.Write([]byte("Could not get " + route))
}

func Err405(res http.ResponseWriter, req *http.Request) {
	res.WriteHeader(405)
	res.Write([]byte("Not Allowed"))
}
