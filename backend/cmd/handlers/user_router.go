package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/petr-discover/cmd/database"
)

type FriendRequest struct {
	UserName string `json:"username"`
}

type UserCardRequest struct {
	FirstName        string `json:"first_name"`
	LastName         string `json:"last_name"`
	UserProfileImage string `json:"user_profile_image"`
}

func UserCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func CreateUserCard(w http.ResponseWriter, r *http.Request) {
	username, exists := CheckLogin(w, r)
	if !exists {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Not logged in"))
		return
	}
	var userCard UserCardRequest
	err := json.NewDecoder(r.Body).Decode(&userCard)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	session := database.Neo4jDriver.NewSession(database.Neo4jCtx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(database.Neo4jCtx)

	_, err = session.ExecuteWrite(database.Neo4jCtx, func(transaction neo4j.ManagedTransaction) (any, error) {
		_, err := transaction.Run(database.Neo4jCtx,
			"CREATE (u:User {username: $username})-[:HAS_CARD]->(c:Card {first_name: $first_name, last_name: $last_name, user_profile_image: $user_profile_image}) RETURN u, c",
			map[string]any{
				"username":           username,
				"first_name":         userCard.FirstName,        // Replace with actual first_name from request/body
				"last_name":          userCard.LastName,         // Replace with actual last_name from request/body
				"user_profile_image": userCard.UserProfileImage, // Replace with actual URL from request/body
			})
		if err != nil {
			return nil, err
		}

		return nil, nil
	})
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"Failed to create User and Card nodes"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"User and Card nodes created successfully"}`))
}

func AddFriend(w http.ResponseWriter, r *http.Request) {
	username, exists := CheckLogin(w, r)
	if !exists {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Not logged in"))
		return
	}

	var friendRequest FriendRequest
	err := json.NewDecoder(r.Body).Decode(&friendRequest)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	session := database.Neo4jDriver.NewSession(database.Neo4jCtx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(database.Neo4jCtx)

	_, err = session.ExecuteWrite(database.Neo4jCtx, func(transaction neo4j.ManagedTransaction) (any, error) {
		_, err := transaction.Run(database.Neo4jCtx,
			"MATCH (u:User {username: $username}), (f:User {username: $friend_username}) "+
				"MERGE (u)-[:SENT_FRIEND_REQUEST]->(request:FriendRequest)-[:TO_USER]->(f) RETURN request",
			map[string]any{
				"username":        username,
				"friend_username": friendRequest.UserName,
			})
		if err != nil {
			return nil, err
		}

		return nil, nil
	})
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"Failed to create FRIENDS_WITH relationship"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"FRIENDS_WITH relationship created successfully"}`))
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	username, exists := CheckLogin(w, r)
	if exists {
		w.Write([]byte(username))
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Not logged in"))
	}
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	username, exists := CheckLogin(w, r)
	if exists {
		w.Write([]byte(username))
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Not logged in"))
	}
}
