DROP INDEX IF EXISTS idx_workspace_members_workspace_id;
DROP INDEX IF EXISTS idx_workspace_members_user_id;
DROP TABLE IF EXISTS workspace_members;

DROP INDEX IF EXISTS idx_workspaces_slug;
DROP INDEX IF EXISTS idx_workspaces_created_by;
DROP TABLE IF EXISTS workspaces;
