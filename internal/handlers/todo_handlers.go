// takes care of Business Logic, identifies request types,Acts as middle layer between client ↔ database
package handlers

import (
	"net/http"
	"strconv"
	"todo_api/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CreateTodoInput struct { //This defines what the API expects from the client
	Title     string `json:"title" binding:"required"` //json:"title" → maps JSON → Go struct ,binding:"required" → Title MUST exist (validation)
	Completed bool   `json:"completed"`
}

type UpdateTodoInput struct {
	Title *string `json:"title"`
	//&true --> set completed as --> true
	//&false --> set completed as --> false
	//nil --> set completed as -> not provided
	Completed *bool `json:"completed"`
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

func GetAllTodosHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		todos, err := repository.GetAllTodos(pool)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, todos)
	}
}

func GetTodoByIDHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		idstr := c.Param("id") //- Gets `"id"` from the URL, Example route:```go GET /todos/5

		id, err := strconv.Atoi(idstr) // Converts string → integer

		if err != nil { //- Checks if conversion failed
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid todo ID"})
			return
		}

		todo, err := repository.GetTodoByID(pool, id) //Calls your repository function, Fetches todo from database

		if err != nil { //- Checks if DB query failed
			if err == pgx.ErrNoRows { //Special case: no record found, This means: “Query worked, but no matching ID exists”
				c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}) //- Handles other errors (DB down, query issue, etc.)
			return
		}

		c.JSON(http.StatusOK, todo)
	}
}

func UpdateToDoHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		idstr := c.Param("id")

		id, err := strconv.Atoi(idstr)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid todo ID"})
		}

		var input UpdateTodoInput

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if input.Title == nil && input.Completed == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "At least one field must be provided"})
			return
		}

		existing, err := repository.GetTodoByID(pool, id)

		if err != nil {
			if err == pgx.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		title := existing.Title
		if input.Title != nil {
			title = *input.Title
		}

		completed := existing.Completed

		if input.Completed != nil {
			completed = *input.Completed
		}

		todo, err := repository.UpdateToDo(pool, id, title, completed)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, todo)
	}
}

func DeleteTodoHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")

		id, err := strconv.Atoi(idStr)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid todo ID"})
		}

		err = repository.DeleteTodo(pool, id)

		if err != nil {
			if err.Error() == "todo with id "+idStr+" not found" {
				c.JSON(http.StatusNotFound, gin.H{"error": "todo not found"})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		c.JSON(http.StatusOK, gin.H{"message": "Deleted Successfully"})
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

/*
Simple flow of Get Single Todo:
Get ID from URL
Convert to int
If invalid → 400
Query DB
If not found → 404
If error → 500
If success → 200 + data
*/
