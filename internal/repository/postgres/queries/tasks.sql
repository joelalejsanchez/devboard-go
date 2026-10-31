-- name:CreateTask :one
INSERT INTO tasks(project_id,title,description,status,assignee_id,created_by)
VALUES($1,$2,$3,$4,$5,$6)
RETURNING *;

-- name:GetTaskByID :one
select * from tasks
where id = $1
limit 1;

-- name: ListTaskByProject :many
SELECT * from tasks
WHERE project_id = $1
ORDER BY created_at ASC;

-- name: UpdateTask :one
UPDATE tasks
SET
    title = COALESCE($2, title),
    description = COALESCE($3, description),
    status = COALESCE($4, status),
    assignee_id = $5,
    updated_at = NOW()
WHERE id = $1
RETURNING *;
