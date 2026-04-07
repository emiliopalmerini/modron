# ADR-001: Custom Notion MCP Server

**Status**: Accepted  
**Date**: 2026-04-07

## Context

The official Notion MCP requires a multi-step workflow (fetch schema, create filtered view, query view) that agents don't follow deterministically. This leads to inconsistent results: sometimes semantic search is used instead of filters, schema fetching is skipped, or wrong property names are guessed.

We need a custom MCP server that exposes per-database tools with typed parameters, so the agent makes a single deterministic call and gets structured results.

## Decision

Build a Go MCP server using `github.com/mark3labs/mcp-go` with hexagonal slice architecture. Each Notion database is a vertical slice owning its domain types, ports, adapters, and MCP tool registration.

### Architecture

```
notion-mcp/
  cmd/
    notion-mcp/
      main.go                    # wiring: creates client, injects into slices, ServeStdio
  internal/
    projects/
      domain.go                  # Project struct, ProjectFilter
      ports.go                   # Repository interface
      notion.go                  # Notion API adapter
      mcp.go                     # MCP tool handlers
    tasks/
      domain.go
      ports.go
      notion.go
      mcp.go
    events/
      domain.go
      ports.go
      notion.go
      mcp.go
    blackhole/
      domain.go
      ports.go
      notion.go
      mcp.go
    shared/
      notion/
        client.go                # HTTP client, auth, request helpers
        filter.go                # Notion API filter builder
        properties.go            # Property value parsing/building
        pagination.go            # Cursor-based pagination
  go.mod
```

### Database Schemas (from Notion API)

#### Projects (`30a7cc4a-ef13-81fd-8b36-fe632b889b70`)

| Property     | Type          | Values                                                 |
| ------------ | ------------- | ------------------------------------------------------ |
| Project name | title         |                                                        |
| Status       | status        | Planning, In Progress, Paused, Backlog, Done, Canceled |
| Tag          | select        | Content, Dev, Marketing, Community, Business, Work     |
| Dates        | date (range)  | start + end                                            |
| Launch date  | date (single) |                                                        |
| Summary      | text          |                                                        |
| Owner        | person        |                                                        |
| Tasks        | relation      | -> Tasks DB                                            |
| Is Blocking  | relation      | -> Projects DB                                         |
| Blocked By   | relation      | -> Projects DB                                         |
| Completion   | rollup        | read-only                                              |

#### Tasks (`30a7cc4a-ef13-8151-8d39-d44ff1118cf7`)

| Property    | Type          | Values                                   |
| ----------- | ------------- | ---------------------------------------- |
| Task name   | title         |                                          |
| Status      | status        | Not Started, In progress, Done, Archived |
| Due         | date (single) |                                          |
| Assignee    | person        |                                          |
| Project     | relation      | -> Projects DB                           |
| Sub-tasks   | relation      | -> Tasks DB                              |
| Parent-task | relation      | -> Tasks DB                              |
| Is Blocking | relation      | -> Tasks DB                              |
| Blocked By  | relation      | -> Tasks DB                              |

#### Events (`30c7cc4a-ef13-80d8-81ec-ffe5d241d8fd`)

| Property | Type  | Values                 |
| -------- | ----- | ---------------------- |
| Name     | title |                        |
| Date     | date  | start (+ optional end) |

#### BlackholeDB (`3097cc4a-ef13-80de-b79b-ec8552eb7d7e`)

| Property  | Type         | Values                                                                                                                                                                                                                                                                                                                                                                                 |
| --------- | ------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Name      | title        |                                                                                                                                                                                                                                                                                                                                                                                        |
| Type      | select       | Idea, Reference, TBR                                                                                                                                                                                                                                                                                                                                                                   |
| Tags      | multi_select | Dev, CSS, AI, Compilers, Anthropic, Agentic, Research, Worldbuilding, D&D, Content, Metodologia, Gaming, Podcast, Travel, Other, Patreon, ASMR, Storie di Vapore, Vivaticket, Video, Done, Episodio singolo, Serie, Prodotto, Tool, Articolo, Template/Risorsa, Workflow, Post Patreon, Manga, Fantasy, Action, Sci-fi, Romcom, Slice of Life, Supernatural, Thriller, Mystery, Horror |
| Summary   | text         |                                                                                                                                                                                                                                                                                                                                                                                        |
| Processed | checkbox     |                                                                                                                                                                                                                                                                                                                                                                                        |
| URL       | url          | (Notion key: `userDefined:URL`)                                                                                                                                                                                                                                                                                                                                                        |

### MCP Tools

#### Query Tools (one per database)

Each query tool accepts optional filter parameters matching the database properties and returns a formatted list of results. All support `sort_by` (property name) and `sort_direction` (asc/desc) parameters, plus `page_size` (default 50, max 100).

| Tool              | Filter Parameters                                                                              |
| ----------------- | ---------------------------------------------------------------------------------------------- |
| `query_projects`  | `status`, `tag`, `name` (contains), `date_start_after`, `date_start_before`, `has_launch_date` |
| `query_tasks`     | `status`, `name` (contains), `due_before`, `due_after`, `project_id`                           |
| `query_events`    | `name` (contains), `date_after`, `date_before`                                                 |
| `query_blackhole` | `type`, `tags` (comma-separated, any match), `name` (contains), `processed`, `has_url`         |

#### Read Tool

| Tool       | Parameters           |
| ---------- | -------------------- |
| `get_page` | `page_id` (required) |

Returns full page properties as formatted text.

#### Write Tools

| Tool                     | Required Parameters               | Optional Parameters                                                   |
| ------------------------ | --------------------------------- | --------------------------------------------------------------------- |
| `create_project`         | `name`                            | `status`, `tag`, `summary`, `dates_start`, `dates_end`, `launch_date` |
| `create_task`            | `name`, `due`                     | `status`, `project_id`, `parent_task_id`                              |
| `create_event`           | `name`, `date_start`              | `date_end`, `is_datetime`                                             |
| `create_blackhole_entry` | `name`, `type`                    | `summary`, `tags`, `url`, `processed`                                 |
| `update_page`            | `page_id` + at least one property | All writable properties for the page's database                       |

### Inputs

- **NOTION_TOKEN**: env var, Notion internal integration token (managed by sops-nix)
- **Database IDs**: hardcoded constants (they belong to the user and don't change)

### Outputs

- Query tools return markdown-formatted text listing matched pages with their properties
- Write tools return the created/updated page ID and URL
- Errors return `mcp.NewToolResultError` with a descriptive message

### Edge Cases

- **Empty filters**: return all pages (up to page_size), sorted by last edited descending
- **Invalid select/status values**: return error listing valid options
- **Pagination**: if `has_more` is true, append a note with the cursor for the next page; accept `start_cursor` parameter on query tools
- **Rate limiting**: Notion API rate limit is 3 req/s; use simple retry with backoff on 429
- **Relation filters**: `project_id` and `parent_task_id` accept a Notion page UUID; the adapter converts to the relation filter format
- **Date parameters**: accept ISO-8601 date strings (`2026-04-07`) or datetime strings (`2026-04-07T09:00:00`)

### Error Conditions

- Missing NOTION_TOKEN env var: fatal at startup
- Notion API errors (400, 401, 404, 429, 500): mapped to descriptive tool errors
- Invalid parameter values: validated before API call, return error with valid options

## Consequences

- Agents make a single tool call instead of 3-4 steps; deterministic behavior
- Database schema changes require updating the corresponding slice's domain types
- Replaces the official Notion MCP for this workspace (or runs alongside it)
- The shared Notion HTTP client can be reused if more databases are added later
