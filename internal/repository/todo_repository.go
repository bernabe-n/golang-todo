// MAKE SQL QUERIES, INTERACT WITH DATABASE
package repository

import (
	"context"
	"time"
	"todo_api/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateTodo(pool *pgxpool.Pool, title string, completed bool) (*models.Todo, error) {
	var ctx context.Context
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second) //Creates a new context with 5-second timeout, If DB takes too long → it cancels automatically
	defer cancel()                                                         // <----Prevents memory leaks / hanging resources
	//schedule cancel to run when CreateTodo returns,for context resources to be released

	var query string = `
			INSERT INTO todos (title, completed)
			VALUES ($1, $2)
			RETURNING id, title, completed, created_at, updated_at
	`

	var todo models.Todo

	var err error = pool.QueryRow(ctx, query, title, completed).Scan( //execute the query string, Scan --->return from postgres data and assign to todo struct
		&todo.ID,
		&todo.Title,
		&todo.Completed,
		&todo.CreatedAt,
		&todo.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &todo, nil
}

func GetAllTodos(pool *pgxpool.Pool) ([]models.Todo, error) { //accepts a database connection pool (used to query PostgreSQL)returns a slice (list) of Todo
	var ctx context.Context       //→ context (used to control request lifetime)
	var cancel context.CancelFunc //→ function to cancel the context

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second) //Creates a context with 5-second timeout
	defer cancel()                                                         //Ensures cancel() runs when function finishes

	var query string = `
		SELECT id, title, completed, created_at, updated_at
		FROM todos
		ORDER BY created_at DESC
	`
	var rows, err = pool.Query(ctx, query) //Executes SQL using: pool → database connection,ctx → with timeout

	if err != nil {
		return nil, err
	}

	defer rows.Close() //Ensures database resources are freed after function ends

	var todos []models.Todo = []models.Todo{} //Creates empty slice to store todos

	for rows.Next() { //Loops over each row returned from database
		var todo models.Todo //Temporary variable to store one row

		err = rows.Scan(
			&todo.ID,
			&todo.Title,
			&todo.Completed,
			&todo.CreatedAt,
			&todo.UpdatedAt,
		) //Copies column values into struct fields, Order must match SQL query

		if err != nil {
			return nil, err
		}

		todos = append(todos, todo) //Adds current todo to list
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return todos, nil
}

func GetTodoByID(pool *pgxpool.Pool, id int) (*models.Todo, error) { //*models.Todo` → pointer to a Todo (can be `nil`)
	var ctx context.Context
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var query string = `
		SELECT id, title, completed, created_at, updated_at
		FROM todos
		WHERE id = $1
	`

	var todo models.Todo

	var err error = pool.QueryRow(ctx, query, id).Scan(
		&todo.ID,
		&todo.Title,
		&todo.Completed,
		&todo.CreatedAt,
		&todo.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &todo, nil
}

func UpdateToDo(pool *pgxpool.Pool, id int, title string, completed bool) (*models.Todo, error) {
	var ctx context.Context
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var query string = `
		UPDATE todos
		SET title = $1, completed = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3
		RETURNING id, title, completed, created_at, updated_at
	`
	var todo models.Todo

	var err error = pool.QueryRow(ctx, query, title, completed, id).Scan(
		&todo.ID,
		&todo.Title,
		&todo.Completed,
		&todo.CreatedAt,
		&todo.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &todo, nil
}

/*Summary → CreateTodo()
  → SQL INSERT
  → PostgreSQL
  → returns row
  → Scan into struct
  → return Todo

  	context → controls timeout
	query → SQL command
	QueryRow → executes query
	Scan → maps DB → struct
	defer cancel() → cleanup
	return → send result back to handler

	What happens in database

	A new row is inserted into todos table:

	id: 1
	title: "Buy milk"
	completed: false
	created_at: 2026-04-13 21:10:00
	updated_at: 2026-04-13 21:10:00

	Your function returns:

	&models.Todo{
		ID:        1,
		Title:     "Buy milk",
		Completed: false,
		CreatedAt: time.Date(2026, 4, 13, 21, 10, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 4, 13, 21, 10, 0, 0, time.UTC),
	}
*/

/*Summary ->

This function:

Creates a 5-second timeout
Runs a SELECT query
Loops through results
Converts rows → Go structs
Stores them in a slice
Returns the list
*/
