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
}
