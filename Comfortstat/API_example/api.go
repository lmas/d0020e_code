package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// serving as a template
type todo struct {
	ID        string `json:"id"`
	Item      string `json:"title"`
	Completed bool   `json:"completed"`
}

var todos = []todo{
	{ID: "1", Item: "Clean room", Completed: false},
	{ID: "2", Item: "Read Book", Completed: false},
	{ID: "3", Item: "Record Video", Completed: false},
}

func getTodos(context *gin.Context) {
	context.IndentedJSON(http.StatusOK, todos)
}

/*
func main() {
	router := gin.Default() // Create a server(router)
	router.GET("/todos", getTodos)
	router.Run("localhost:9090") // App should be running on port 9090 (the path)

}
*/
