package handlers

import (
    "os"
    "context"
    "encoding/json"
    "net/http"
    "gopoc/model"
    "gopoc/db"
    "log"
	"time"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "golang.org/x/crypto/bcrypt"
	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
    UserID string `json:"userId"`
    jwt.StandardClaims
}

type UserResponse struct {
    ID string `json:"id"`
    Username string `json:"username"`
    Name string `json:"name"`
    ProfilePicUrl string `json:"profilePicUrl"`
    CreatedAt time.Time `json:"createdAt"`
    UpdatedAt *time.Time `json:"updatedAt"`
    Token string `json:"token"`
}

type UserListResponse struct {
    ID string `json:"id"`
    Username string `json:"username"`
    Name string `json:"name"`
    ProfilePicUrl string `json:"profilePicUrl"`
}

func createUserListResponse(users []model.User) []UserListResponse {
    var responseList []UserListResponse
    for _, u := range users {
        responseList = append(responseList, UserListResponse{
            ID: u.ID.Hex(), 
            Username: u.Username,
            Name: u.Name,
            ProfilePicUrl: u.ProfilePicUrl,
        })
    }
    return responseList
}

func createUserResponse(user model.User) UserResponse {
    Token, err := createToken(user.ID.Hex())
    if err != nil {
        log.Fatal(err)
    }
    return UserResponse{
        ID: user.ID.Hex(),
        Username: user.Username,
        Name: user.Name,
        ProfilePicUrl: user.ProfilePicUrl,
        CreatedAt: user.CreatedAt,
        UpdatedAt: user.UpdatedAt,
        Token: Token,
    }
}

func createToken(userID string) (string, error) {
    expirationTime := time.Now().Add(24 * time.Hour)
    claims := &Claims{
        UserID: userID,
        StandardClaims: jwt.StandardClaims{
            ExpiresAt: expirationTime.Unix(),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

    tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
    if err != nil {
        return "", err
    }

    return tokenString, nil
}

func ParseToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        return []byte(os.Getenv("JWT_SECRET")), nil
    })
    if err != nil {
        return nil, err
    }

    claims, ok := token.Claims.(*Claims)
    if !ok {
        return nil, err
    }

    return claims, nil
}

func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), 4)
    if err != nil {
        return "", err
    }
    return string(bytes), nil
}

func CheckPasswordHash(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
    var user model.User
    if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    collection := db.Client.Database("chatDB").Collection("users")
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    filter := bson.M{"username": user.Username}
    if err := collection.FindOne(ctx, filter).Decode(&user); err == nil {
        http.Error(w, "User already exists with same username", http.StatusConflict)
        return
    }

    user.ID = primitive.NewObjectID()
    user.CreatedAt = time.Now()
    hash, err := HashPassword(user.Password)
    if err != nil {
        log.Fatal(err)
    }
    user.Password = hash

    _, err = collection.InsertOne(ctx, user)
    if err != nil {
        log.Printf("Failed to create user: %v", err)
        http.Error(w, "Failed to create user", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(createUserResponse(user))
}


func GetUser(w http.ResponseWriter, r *http.Request) {
    var user model.User
    if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    Username := user.Username
    Password := user.Password

    collection := db.Client.Database("chatDB").Collection("users")
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    filter := bson.M{"username": Username}
    if err := collection.FindOne(ctx, filter).Decode(&user); err != nil {
        http.Error(w, "User not found", http.StatusNotFound)
        return
    }

    if !CheckPasswordHash(Password, user.Password) {
        http.Error(w, "Invalid password", http.StatusUnauthorized)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(createUserResponse(user))
}

func ListUsers(w http.ResponseWriter, r *http.Request) {
    collection := db.Client.Database("chatDB").Collection("users")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    var users []model.User
    cursor, err := collection.Find(ctx, bson.M{})
    if err != nil {
        http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
        return
    }
    defer cursor.Close(ctx)
    if err = cursor.All(ctx, &users); err != nil {
        http.Error(w, "Failed to parse users", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(createUserListResponse(users))
}