ALTER TABLE todos ADD COLUMN user_id UUID NOT NULL;

ALTER TABLE todos ADD CONSTRAINT fk_todos_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;


--These are SQL schema changes that connect your todos table to a users table—basically enforcing “each todo belongs to a user.”
/*
ALTER TABLE todos
→ Modify the existing todos table
ADD COLUMN user_id
→ Add a new column named user_id
UUID
→ Data type is a UUID (e.g. 550e8400-e29b-41d4-a716-446655440000)
This matches typical users.id types in modern apps
NOT NULL
→ Every todo must have a user_id
No todo can exist without being tied to a user

ADD CONSTRAINT fk_todos_user
Adds a named constraint
Name: fk_todos_user
fk = foreign key
naming helps when debugging or removing later
FOREIGN KEY (user_id)
Declares that user_id in todos is a foreign key
Meaning: it must match a value in another table
REFERENCES users(id)
Points to:
table: users
column: id

So now:

every todos.user_id must exist in users.id

ON DELETE CASCADE

This is the most important behavior:

If a user is deleted:
→ all their todos are automatically deleted
*/