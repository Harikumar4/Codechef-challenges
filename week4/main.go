package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type User struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username string             `bson:"username" json:"username"`
	Password string             `bson:"password" json:"password"`
	Role     string             `bson:"role" json:"role"`
}

type Event struct {
	ID          primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Title       string               `bson:"title" json:"title"`
	Description string               `bson:"description" json:"description"`
	CreatorID   primitive.ObjectID   `bson:"creator_id" json:"creator_id"`
	Attendees   []primitive.ObjectID `bson:"attendees" json:"attendees"`
	Date        time.Time            `bson:"date" json:"date"`
}

type Attendance struct {
	ID      primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID  primitive.ObjectID `bson:"user_id" json:"user_id"`
	EventID primitive.ObjectID `bson:"event_id" json:"event_id"`
}

var jwtKey = []byte("secret_key")

var userCol *mongo.Collection
var eventCol *mongo.Collection
var attendanceCol *mongo.Collection

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	db := client.Database("eventdb")
	userCol = db.Collection("users")
	eventCol = db.Collection("events")
	attendanceCol = db.Collection("attendances")
	r := gin.Default()
	r.POST("/register", register)
	r.POST("/login", login)
	auth := r.Group("/")
	auth.Use(authMiddleware)
	auth.POST("/events", createEvent)
	auth.GET("/events", listEvents)
	auth.POST("/events/:id/attend", attendEvent)
	r.Run(":8080")
}

type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func register(c *gin.Context) {
	var u User
	if err := c.ShouldBindJSON(&u); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}
	if u.Role != "creator" && u.Role != "attendee" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role"})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var exists User
	err := userCol.FindOne(ctx, bson.M{"username": u.Username}).Decode(&exists)
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user exists"})
		return
	}
	res, err := userCol.InsertOne(ctx, u)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": res.InsertedID})
}

func login(c *gin.Context) {
	var req User
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var u User
	err := userCol.FindOne(ctx, bson.M{"username": req.Username, "password": req.Password}).Decode(&u)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	exp := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: u.ID.Hex(),
		Role:   u.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

func authMiddleware(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		return
	}
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}
	c.Set("user_id", claims.UserID)
	c.Set("role", claims.Role)
	c.Next()
}

func createEvent(c *gin.Context) {
	role := c.GetString("role")
	if role != "creator" {
		c.JSON(http.StatusForbidden, gin.H{"error": "not allowed"})
		return
	}
	var e Event
	if err := c.ShouldBindJSON(&e); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}
	userID := c.GetString("user_id")
	creatorID, _ := primitive.ObjectIDFromHex(userID)
	e.CreatorID = creatorID
	e.Attendees = []primitive.ObjectID{}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res, err := eventCol.InsertOne(ctx, e)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": res.InsertedID})
}

func listEvents(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cur, err := eventCol.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}
	var events []Event
	if err := cur.All(ctx, &events); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}
	c.JSON(http.StatusOK, events)
}

func attendEvent(c *gin.Context) {
	role := c.GetString("role")
	if role != "attendee" {
		c.JSON(http.StatusForbidden, gin.H{"error": "not allowed"})
		return
	}
	userID := c.GetString("user_id")
	uid, _ := primitive.ObjectIDFromHex(userID)
	eid := c.Param("id")
	eventID, err := primitive.ObjectIDFromHex(eid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = eventCol.UpdateOne(ctx, bson.M{"_id": eventID}, bson.M{"$addToSet": bson.M{"attendees": uid}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}
	attendanceCol.InsertOne(ctx, Attendance{UserID: uid, EventID: eventID})
	c.JSON(http.StatusOK, gin.H{"status": "attending"})
}
