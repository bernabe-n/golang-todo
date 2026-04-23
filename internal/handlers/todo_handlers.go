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
		idstr := c.Param("id") //Gets the id from the URL (e.g., /todos/5 → "5")

		id, err := strconv.Atoi(idstr) //Converts the string ID to an integer

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid todo ID"}) //If conversion fails (e.g., /todos/abc)
			return
		}

		var input UpdateTodoInput //Declares a struct to hold incoming JSON data

		if err := c.ShouldBindJSON(&input); err != nil { //Reads JSON body and maps it into input
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}) //If JSON is invalid → return 400 and stop execution
			return
		}
		if input.Title == nil && input.Completed == nil { //Checks if both fields are missing
			c.JSON(http.StatusBadRequest, gin.H{"error": "At least one field must be provided"}) //Returns error if no fields provided
			return
		}

		existing, err := repository.GetTodoByID(pool, id) //Fetches the current todo from database

		if err != nil {
			if err == pgx.ErrNoRows { //If no record found
				c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"}) //Return 404 Not Found
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}) //Otherwise → server error (500)
			return
		}

		title := existing.Title //Default: keep old title
		if input.Title != nil {
			title = *input.Title //If user provided a new title → use it, *input.Title dereferences pointer
		}

		completed := existing.Completed //Default: keep old status

		if input.Completed != nil {
			completed = *input.Completed //If user provided new value → update it
		}

		todo, err := repository.UpdateToDo(pool, id, title, completed) //Calls repository function to update the record

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}) //If update fails → return 500
			return
		}

		c.JSON(http.StatusOK, todo) //Sends updated todo back to client (200 OK)
	}
}

func DeleteTodoHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) { //Returns an anonymous function, c is the Gin context (handles request + response)
		idStr := c.Param("id") //Gets the id from the URL, Example: /todos/10 → "10"

		id, err := strconv.Atoi(idStr) //Converts idStr (string) into an integer

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid todo ID"}) //If conversion fails (e.g., /todos/abc), Sends 400 Bad Request
			return
		}

		err = repository.DeleteTodo(pool, id) //Calls your repository function to delete the todo from the database using the given id

		if err != nil { //Checks if something went wrong during deletion
			if err.Error() == "todo with id "+idStr+" not found" { //Compares error message as a string, If it matches → means the todo doesn’t exist
				c.JSON(http.StatusNotFound, gin.H{"error": "todo not found"})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}) //Any other error → return 500 Internal Server Error
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Deleted Successfully"}) //If no errors → return 200 OK
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
