package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j/dbtype"
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

	result, err := session.Run(database.Neo4jCtx,
		"MATCH (u:User {username: $username}) RETURN u",
		map[string]interface{}{
			"username": username,
		})

	if err != nil {
		log.Fatal(err)
	}
	if result.Next(database.Neo4jCtx) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"User Card Already Exists"}`))
		return
	}

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

	_, err = session.ExecuteWrite(database.Neo4jCtx, func(transaction neo4j.ManagedTransaction) (interface{}, error) {
		// Check if there is a pending friend request
		result, err := transaction.Run(database.Neo4jCtx, "MATCH (u:User {username: $friend_username})-[friendRequest:SENT_FRIEND_REQUEST]-(n:FriendRequest{status: 'pending'})-[dd:TO_USER]->(f:User {username: $username}) RETURN n.status",
			map[string]interface{}{
				"username":        username,
				"friend_username": friendRequest.UserName,
			})
		if err != nil {
			return nil, err
		}

		if result.Next(database.Neo4jCtx) {
			_, err := transaction.Run(database.Neo4jCtx,
				"MATCH (u:User)-[s:SENT_FRIEND_REQUEST]-(fr:FriendRequest)-[dd:TO_USER]->(uu:User) WHERE u.username = $friend_username AND uu.username = $username DELETE fr, dd, s",
				map[string]interface{}{
					"username":        username,
					"friend_username": friendRequest.UserName,
				})
			if err != nil {
				return nil, err
			}

			_, err = transaction.Run(database.Neo4jCtx,
				"MATCH (u:User {username: $username}), (f:User {username: $friend_username}) "+
					"MERGE (u)-[:FRIENDS_WITH]->(f) "+
					"MERGE (f)-[:FRIENDS_WITH]->(u)",
				map[string]interface{}{
					"username":        username,
					"friend_username": friendRequest.UserName,
				})
			if err != nil {
				return nil, err
			}

		} else {
			_, err := transaction.Run(database.Neo4jCtx,
				"MATCH (u:User {username: $username}), (f:User {username: $friend_username}) "+
					"MERGE (u)-[:SENT_FRIEND_REQUEST]->(request:FriendRequest {status: 'pending', sender: $sender})-[:TO_USER]->(f) RETURN request",
				map[string]interface{}{
					"username":        username,
					"friend_username": friendRequest.UserName,
					"sender":          username, // Add the sender property here
				})
			if err != nil {
				return nil, err
			}
		}

		return nil, nil
	})

	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"Failed to add friend"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"Friend request processed successfully"}`))
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	var username string
	n, exists := CheckLogin(w, r)
	if !exists {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Not logged in"))
		return
	}
	var requestBody map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		log.Println(err)
		return
	}
	usernameFromBody, exists := requestBody["username"].(string)
	if !exists || usernameFromBody == "" {
		username = n
	} else {
		username = usernameFromBody
	}

	session := database.Neo4jDriver.NewSession(database.Neo4jCtx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(database.Neo4jCtx)

	result, err := session.Run(database.Neo4jCtx,
		"MATCH (u:User {username: $username})-[:HAS_CARD]->(c:Card) RETURN u, c",
		map[string]interface{}{
			"username": username,
		})
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"Failed to retrieve user data"}`))
		return
	}

	record, err := result.Single(database.Neo4jCtx)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"User not found"}`))
		return
	}

	userNode, ok := record.Values[0].(dbtype.Node)
	if !ok {
		log.Println("Failed to convert to Node")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"Failed to convert to Node"}`))
		return
	}

	cardNode, ok := record.Values[1].(dbtype.Node)
	if !ok {
		log.Println("Failed to convert to Node")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"Failed to convert to Node"}`))
		return
	}

	userProps := userNode.Props
	cardProps := cardNode.Props

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]interface{}{
		"user": userProps,
		"card": cardProps,
	}
	json.NewEncoder(w).Encode(response)
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	username, exists := CheckLogin(w, r)
	if !exists {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Not logged in"))
		return
	}

	// Parse JSON request body
	var updateRequest map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&updateRequest)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"Invalid request body"}`))
		return
	}

	cardPropsToUpdate, ok := updateRequest["card"].(map[string]interface{})
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"Invalid card properties in request"}`))
		return
	}

	session := database.Neo4jDriver.NewSession(database.Neo4jCtx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(database.Neo4jCtx)

	_, err = session.ExecuteWrite(database.Neo4jCtx, func(transaction neo4j.ManagedTransaction) (any, error) {
		_, err := transaction.Run(database.Neo4jCtx,
			"MATCH (u:User {username: $username})-[:HAS_CARD]->(c:Card) SET c += $cardPropsToUpdate RETURN u, c",
			map[string]any{
				"username":          username,
				"cardPropsToUpdate": cardPropsToUpdate,
			})
		if err != nil {
			return nil, err
		}

		return nil, nil
	})

	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"Failed to update card properties"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"Card properties updated successfully"}`))
}
