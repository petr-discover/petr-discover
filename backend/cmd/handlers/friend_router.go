package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/petr-discover/cmd/database"
)

type UserNode struct {
	UserName string `json:"username"`
}

func FriendCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func GetGraph(w http.ResponseWriter, r *http.Request) {
	_, exists := CheckLogin(w, r)
	if !exists {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Not logged in"))
		return
	}
	session := database.Neo4jDriver.NewSession(database.Neo4jCtx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(database.Neo4jCtx)
	result, err := session.Run(database.Neo4jCtx, "MATCH (u:User) RETURN u", map[string]interface{}{})
	if err != nil {
		log.Fatal(err)
	}
	var users []UserNode
	for result.Next(database.Neo4jCtx) {
		record := result.Record()
		userNode, b := record.Get("u")

		if b {
			user := UserNode{
				UserName: userNode.(neo4j.Node).Props["username"].(string),
			}
			users = append(users, user)
		}
	}
	jsonResponse := map[string][]UserNode{"users": users}
	responseBytes, err := json.Marshal(jsonResponse)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"Failed to marshal JSON response"}`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(responseBytes)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"Failed to write JSON response"}`))
		return
	}
}

func GetPendingFriend(w http.ResponseWriter, r *http.Request) {
	username, exists := CheckLogin(w, r)
	if !exists {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Not logged in"))
		return
	}
	session := database.Neo4jDriver.NewSession(database.Neo4jCtx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(database.Neo4jCtx)

	result, err := session.ExecuteRead(database.Neo4jCtx, func(transaction neo4j.ManagedTransaction) (any, error) {
		result, err := transaction.Run(database.Neo4jCtx,
			"MATCH (u:User)<-[:TO_USER]-(request:FriendRequest{status: 'pending'}) "+
				"WHERE u.username = $username "+
				"RETURN DISTINCT request.sender "+
				"LIMIT 25",
			map[string]interface{}{
				"username": username,
			})
		fmt.Println("this is result")
		fmt.Println(result)
		if err != nil {
			log.Println(err)
			http.Error(w, `{"message":"Failed to retrieve pending friend requests"}`, http.StatusInternalServerError)
		}
		var friendUsernames []string
		for result.Next(database.Neo4jCtx) {
			record := result.Record()
			fmt.Println(record)
			friendUsername, b := record.Get("request.sender")
			if b {
				friendUsernames = append(friendUsernames, friendUsername.(string))
			}
		}
		fmt.Println(friendUsernames)
		return friendUsernames, nil
	})
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"Failed to retrieve pending friend requests"}`))
		return
	}
	fmt.Println(result)
	jsonResponse := map[string][]string{"pending_friends": result.([]string)}
	responseBytes, err := json.Marshal(jsonResponse)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"Failed to marshal JSON response"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")

	_, err = w.Write(responseBytes)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"Failed to write JSON response"}`))
		return
	}
	w.WriteHeader(http.StatusOK)
}

func DeleteFriend(w http.ResponseWriter, r *http.Request) {
	username, exists := CheckLogin(w, r)
	if !exists {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Not logged in"))
		return
	}

	var friendToRemove struct {
		FriendUsername string `json:"friend_username"`
	}
	err := json.NewDecoder(r.Body).Decode(&friendToRemove)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	session := database.Neo4jDriver.NewSession(database.Neo4jCtx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(database.Neo4jCtx)

	_, err = session.ExecuteWrite(database.Neo4jCtx, func(transaction neo4j.ManagedTransaction) (interface{}, error) {
		_, err := transaction.Run(database.Neo4jCtx,
			"MATCH (u:User {username: $username})-[r:FRIENDS_WITH]-(f:User {username: $friendUsername}) "+
				"DELETE r",
			map[string]interface{}{
				"username":       username,
				"friendUsername": friendToRemove.FriendUsername,
			})
		if err != nil {
			return nil, err
		}

		return nil, nil
	})

	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"Failed to remove friend relationship"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"Friend relationship removed successfully"}`))
}
