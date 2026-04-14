// takes care of Business Logic, identifies request types,Acts as middle layer between client ↔ database
package handlers

import (
	"net/http"
	"todo_api/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CreateTodoInput struct { //This defines what the API expects from the client
	Title     string `json:"title" binding:"required"` //json:"title" → maps JSON → Go struct ,binding:"required" → Title MUST exist (validation)
	Completed bool   `json:"completed"`
}

func CreatedTodoHandler(pool *pgxpool.Pool) gin.HandlerFunc { //gin.HandlerFunc :reads request (c.ShouldBindJSON),processes logic,sends response (c.JSON)
	return func(c *gin.Context) {
		var input CreateTodoInput //Create empty struct to store incoming JSON

		if err := c.ShouldBindJSON(&input); err != nil { //makes sure that title exist and has correct type, Converts request body into Go struct
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		todo, err := repository.CreateTodo(pool, input.Title, input.Completed) //This sends data to database layer

		//Communicate to repository
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, todo) //sends the todo back to the user
	}
}

/*Client (Postman)
     │
     ▼
POST /todos
     │
     ▼
Gin Handler (CreatedTodoHandler)
     │
     ├── Bind JSON → input struct
     │
     ├── Call repository (DB layer)
     │
     └── Return JSON response
     ▼
Client receives result*/

/*FLOW:
SERVER STARTS
    │
    ▼
CreatedTodoHandler(pool)
    │
    ▼
RETURNS this function:
    func(c *gin.Context)
    │
    ▼
Gin stores it as route handler
    │
    ▼
POST /todos happens
    │
    ▼
Gin executes the returned function
*/
