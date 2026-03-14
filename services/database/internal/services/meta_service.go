package services

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type MetaService struct {
	db  *pgxpool.Pool
	log *zap.Logger
}

func (s *MetaService) DB() *pgxpool.Pool {
	return s.db
}

func NewMetaService(db *pgxpool.Pool, log *zap.Logger) *MetaService {
	service := &MetaService{
		db:  db,
		log: log,
	}
	if err := service.ensureSystemSchema(context.Background()); err != nil {
		log.Warn("failed to initialize system schema", zap.Error(err))
	}
	if err := service.ensureDashboardAccess(context.Background()); err != nil {
		log.Warn("failed to initialize dashboard grants", zap.Error(err))
	}
	return service
}

type TableMeta struct {
	Name     string `json:"name"`
	Schema   string `json:"schema"`
	RowCount int64  `json:"row_count"`
	Size     string `json:"size"`
	HasRLS   bool   `json:"has_rls"`
}

type CreateTableRequest struct {
	Name    string             `json:"name"`
	Schema  string             `json:"schema"`
	Columns []ColumnDefinition `json:"columns"`
}

type ColumnDefinition struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	IsNullable bool   `json:"is_nullable"`
	IsPrimary  bool   `json:"is_primary"`
	Default    string `json:"default,omitempty"`
}

type FunctionMeta struct {
	Name       string `json:"name"`
	Schema     string `json:"schema"`
	ReturnType string `json:"return_type"`
	Language   string `json:"language"`
	Arguments  string `json:"arguments"`
	Definition string `json:"definition"`
}

type PolicyMeta struct {
	Name      string   `json:"name"`
	Schema    string   `json:"schema"`
	Table     string   `json:"table"`
	Action    string   `json:"action"` // ALL, SELECT, INSERT, etc.
	Roles     []string `json:"roles"`
	Qualifier string   `json:"qualifier"` // USING / WITH CHECK
}

type MigrationMeta struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Query      string `json:"query"`
	ExecutedAt string `json:"executed_at"`
}

// GetTables returns a list of tables and their metadata in the specified schema
func (s *MetaService) GetTables(ctx context.Context, schema string) ([]TableMeta, error) {
	query := `
		SELECT 
			c.relname as name,
			n.nspname as schema,
			COALESCE(s.n_live_tup, 0) as row_count,
			pg_size_pretty(pg_total_relation_size(c.oid)) as size,
			c.relrowsecurity as has_rls
		FROM pg_class c
		JOIN pg_namespace n ON n.oid = c.relnamespace
		LEFT JOIN pg_stat_user_tables s ON s.relid = c.oid
		WHERE n.nspname = $1 AND c.relkind = 'r'
		ORDER BY c.relname;
	`
	rows, err := s.db.Query(ctx, query, schema)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []TableMeta
	for rows.Next() {
		var t TableMeta
		if err := rows.Scan(&t.Name, &t.Schema, &t.RowCount, &t.Size, &t.HasRLS); err != nil {
			return nil, err
		}
		tables = append(tables, t)
	}
	return tables, nil
}

func (s *MetaService) GetTableColumns(ctx context.Context, schema, table string) ([]ColumnDefinition, error) {
	query := `
		SELECT
			c.column_name,
			c.udt_name,
			c.is_nullable = 'YES' AS is_nullable,
			COALESCE(tc.constraint_type = 'PRIMARY KEY', false) AS is_primary,
			COALESCE(c.column_default, '') AS column_default
		FROM information_schema.columns c
		LEFT JOIN information_schema.key_column_usage kcu
			ON c.table_schema = kcu.table_schema
			AND c.table_name = kcu.table_name
			AND c.column_name = kcu.column_name
		LEFT JOIN information_schema.table_constraints tc
			ON kcu.constraint_name = tc.constraint_name
			AND kcu.table_schema = tc.table_schema
			AND kcu.table_name = tc.table_name
		WHERE c.table_schema = $1 AND c.table_name = $2
		ORDER BY c.ordinal_position;
	`

	rows, err := s.db.Query(ctx, query, schema, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []ColumnDefinition
	for rows.Next() {
		var col ColumnDefinition
		if err := rows.Scan(&col.Name, &col.Type, &col.IsNullable, &col.IsPrimary, &col.Default); err != nil {
			return nil, err
		}
		columns = append(columns, col)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return columns, nil
}

// RunQuery executes an arbitrary SQL string and returns the raw rows
// In OmniBase, this requires Service Role token (Admin only).
func (s *MetaService) RunQuery(ctx context.Context, sql string, migrationName string) ([]map[string]interface{}, error) {
	trimmed := strings.TrimSpace(sql)
	if trimmed == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}

	if queryReturnsRows(trimmed) {
		var results []map[string]interface{}
		err := pgx.BeginFunc(ctx, s.db, func(tx pgx.Tx) error {
			if _, err := tx.Exec(ctx, "SET LOCAL search_path TO public"); err != nil {
				return err
			}

			rows, err := tx.Query(ctx, sql)
			if err != nil {
				return err
			}
			defer rows.Close()

			fields := rows.FieldDescriptions()
			for rows.Next() {
				values, err := rows.Values()
				if err != nil {
					return err
				}

				rowMap := make(map[string]interface{})
				for i, field := range fields {
					rowMap[string(field.Name)] = values[i]
				}
				results = append(results, rowMap)
			}

			return rows.Err()
		})
		if err != nil {
			return nil, err
		}
		return results, nil
	}

	var tag pgconn.CommandTag
	err := pgx.BeginFunc(ctx, s.db, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, "SET LOCAL search_path TO public"); err != nil {
			return err
		}

		execTag, err := tx.Exec(ctx, sql)
		if err != nil {
			return err
		}
		tag = execTag
		return nil
	})
	if err != nil {
		return nil, err
	}
	if tableRef := extractCreatedTableRef(trimmed); tableRef != nil {
		if err := s.grantDashboardAccess(ctx, tableRef.Schema, tableRef.Table); err != nil {
			s.log.Warn("failed to grant table access after query", zap.Error(err), zap.String("schema", tableRef.Schema), zap.String("table", tableRef.Table))
		}
	}
	if migrationName != "" {
		if err := s.recordMigration(ctx, migrationName, sql); err != nil {
			s.log.Warn("failed to record migration", zap.Error(err))
		}
	}
	if isSchemaMutation(trimmed) {
		_ = s.ReloadSchemaCache(ctx)
	}

	return []map[string]interface{}{{
		"command":       tag.String(),
		"rows_affected": tag.RowsAffected(),
	}}, nil
}

// ResolveGraphQL calls the pg_graphql extension to resolve a query
func (s *MetaService) ResolveGraphQL(ctx context.Context, query string, variables map[string]interface{}, userID, role string) (interface{}, error) {
	var resultBytes []byte

	varsStr := "{}"
	if variables != nil {
		varsBytes, _ := json.Marshal(variables)
		varsStr = string(varsBytes)
	}

	// Use a transaction to set the role and claims for RLS
	err := pgx.BeginFunc(ctx, s.db, func(tx pgx.Tx) error {
		// Set the role
		if role != "" {
			if _, err := tx.Exec(ctx, fmt.Sprintf("SET LOCAL ROLE %s", pgx.Identifier{role}.Sanitize())); err != nil {
				return err
			}
		}

		// Set the claims (PostgREST style)
		if userID != "" || role != "" {
			claims := map[string]interface{}{
				"sub":  userID,
				"role": role,
			}
			claimsJSON, _ := json.Marshal(claims)
			if _, err := tx.Exec(ctx, "SELECT set_config('request.jwt.claims', $1, true)", string(claimsJSON)); err != nil {
				return err
			}
		}

		return tx.QueryRow(ctx, "SELECT graphql.resolve(query := $1, variables := $2)", query, varsStr).Scan(&resultBytes)
	})

	if err != nil {
		return nil, err
	}

	var parsed interface{}
	if err := json.Unmarshal(resultBytes, &parsed); err != nil {
		return nil, err
	}

	return parsed, nil
}

// CreateTable creates a new table in the specified schema
func (s *MetaService) CreateTable(ctx context.Context, req CreateTableRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return fmt.Errorf("table name is required")
	}
	if req.Schema == "" {
		req.Schema = "public"
	}
	if len(req.Columns) == 0 {
		return fmt.Errorf("at least one column is required")
	}

	sql := fmt.Sprintf("CREATE TABLE %s.%s (", pgx.Identifier{req.Schema}.Sanitize(), pgx.Identifier{req.Name}.Sanitize())
	var primaryKeys []string
	var columns []string

	for _, col := range req.Columns {
		if strings.TrimSpace(col.Name) == "" {
			continue
		}
		colDef := fmt.Sprintf("%s %s", pgx.Identifier{col.Name}.Sanitize(), col.Type)
		if !col.IsNullable {
			colDef += " NOT NULL"
		}
		if col.Default != "" {
			colDef += fmt.Sprintf(" DEFAULT %s", col.Default)
		}
		columns = append(columns, colDef)
		if col.IsPrimary {
			primaryKeys = append(primaryKeys, pgx.Identifier{col.Name}.Sanitize())
		}
	}
	if len(columns) == 0 {
		return fmt.Errorf("at least one valid column is required")
	}

	sql += strings.Join(columns, ", ")
	if len(primaryKeys) > 0 {
		sql += fmt.Sprintf(", PRIMARY KEY (%s)", strings.Join(primaryKeys, ", "))
	}
	sql += ");"

	_, resErr := s.db.Exec(ctx, sql)
	if resErr == nil {
		if err := s.grantDashboardAccess(ctx, req.Schema, req.Name); err != nil {
			s.log.Warn("failed to grant table access after table creation", zap.Error(err), zap.String("schema", req.Schema), zap.String("table", req.Name))
		}
		_ = s.recordMigration(ctx, "create_table_"+req.Name, sql)
		s.ReloadSchemaCache(ctx)
	}
	return resErr
}

// CreateProject creates a new database schema for a project
func (s *MetaService) CreateProject(ctx context.Context, name string) error {
	// 1. Create schema
	schemaIdent := pgx.Identifier{name}.Sanitize()
	_, err := s.db.Exec(ctx, fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schemaIdent))
	if err != nil {
		return err
	}

	// 2. Grant usage to public/anon (Phase 1 simplicity)
	// In production, we'd create dedicated roles per project
	_, err = s.db.Exec(ctx, fmt.Sprintf("GRANT USAGE ON SCHEMA %s TO authenticated, anon", schemaIdent))
	if err != nil {
		return err
	}

	_, resErr := s.db.Exec(ctx, fmt.Sprintf("ALTER DEFAULT PRIVILEGES IN SCHEMA %s GRANT ALL ON TABLES TO authenticated, anon", schemaIdent))
	if resErr == nil {
		_ = s.recordMigration(ctx, "create_project_"+name, "CREATE SCHEMA "+schemaIdent)
		s.ReloadSchemaCache(ctx)
	}
	return resErr
}

// SetRLSEnabled enabled or disables RLS on a table
func (s *MetaService) SetRLSEnabled(ctx context.Context, schema, table string, enabled bool) error {
	action := "ENABLE"
	if !enabled {
		action = "DISABLE"
	}
	_, err := s.db.Exec(ctx, fmt.Sprintf("ALTER TABLE %s.%s %s ROW LEVEL SECURITY", pgx.Identifier{schema}.Sanitize(), pgx.Identifier{table}.Sanitize(), action))
	return err
}

func (s *MetaService) DeleteTable(ctx context.Context, schema, table string) error {
	if strings.TrimSpace(schema) == "" {
		schema = "public"
	}
	if strings.TrimSpace(table) == "" {
		return fmt.Errorf("table is required")
	}

	_, err := s.db.Exec(ctx, fmt.Sprintf("DROP TABLE %s.%s", pgx.Identifier{schema}.Sanitize(), pgx.Identifier{table}.Sanitize()))
	if err != nil {
		return err
	}

	return s.ReloadSchemaCache(ctx)
}

func (s *MetaService) ListMigrations(ctx context.Context) ([]MigrationMeta, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, name, query, executed_at::text
		FROM omnibase.migrations
		ORDER BY id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var migrations []MigrationMeta
	for rows.Next() {
		var m MigrationMeta
		if err := rows.Scan(&m.ID, &m.Name, &m.Query, &m.ExecutedAt); err != nil {
			return nil, err
		}
		migrations = append(migrations, m)
	}
	return migrations, nil
}

func (s *MetaService) GetFunctions(ctx context.Context, schema string) ([]FunctionMeta, error) {
	query := `
		SELECT 
			p.proname as name,
			n.nspname as schema,
			pg_get_function_result(p.oid) as return_type,
			l.lanname as language,
			pg_get_function_arguments(p.oid) as arguments,
			pg_get_functiondef(p.oid) as definition
		FROM pg_proc p
		JOIN pg_namespace n ON n.oid = p.pronamespace
		JOIN pg_language l ON l.oid = p.prolang
		WHERE n.nspname = $1
		  AND l.lanname NOT IN ('internal', 'c')
		  AND NOT EXISTS (
			SELECT 1 FROM pg_depend d
			WHERE d.objid = p.oid AND d.deptype = 'e'
		  )
		ORDER BY p.proname;
	`
	rows, err := s.db.Query(ctx, query, schema)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fns []FunctionMeta
	for rows.Next() {
		var f FunctionMeta
		if err := rows.Scan(&f.Name, &f.Schema, &f.ReturnType, &f.Language, &f.Arguments, &f.Definition); err != nil {
			return nil, err
		}
		fns = append(fns, f)
	}
	return fns, nil
}

func (s *MetaService) GetPolicies(ctx context.Context, schema string) ([]PolicyMeta, error) {
	query := `
		SELECT 
			policyname as name,
			schemaname as schema,
			tablename as table,
			cmd as action,
			roles::text[] as roles,
			COALESCE(qual, with_check, '') as qualifier
		FROM pg_policies
		WHERE schemaname = $1
		ORDER BY tablename, policyname;
	`
	rows, err := s.db.Query(ctx, query, schema)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var policies []PolicyMeta
	for rows.Next() {
		var p PolicyMeta
		if err := rows.Scan(&p.Name, &p.Schema, &p.Table, &p.Action, &p.Roles, &p.Qualifier); err != nil {
			return nil, err
		}
		policies = append(policies, p)
	}
	return policies, nil
}

func (s *MetaService) GetSchemas(ctx context.Context) ([]string, error) {
	query := `
		SELECT nspname 
		FROM pg_namespace 
		WHERE nspname NOT IN ('pg_catalog', 'information_schema', 'storage', 'auth', 'realtime', 'omnibase')
		AND nspname NOT LIKE 'pg_toast%'
		ORDER BY nspname;
	`
	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schemas []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		schemas = append(schemas, name)
	}
	return schemas, nil
}

// ReloadSchemaCache notifies PostgREST to refresh its schema cache
func (s *MetaService) ReloadSchemaCache(ctx context.Context) error {
	_, err := s.db.Exec(ctx, "NOTIFY pgrst, 'reload schema'")
	return err
}

func (s *MetaService) ensureSystemSchema(ctx context.Context) error {
	_, err := s.db.Exec(ctx, `
		CREATE SCHEMA IF NOT EXISTS omnibase;
		CREATE TABLE IF NOT EXISTS omnibase.migrations (
			id BIGSERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			query TEXT NOT NULL,
			executed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`)
	return err
}

func (s *MetaService) recordMigration(ctx context.Context, name, sql string) error {
	if strings.TrimSpace(name) == "" {
		return nil
	}
	_, err := s.db.Exec(ctx, `
		INSERT INTO omnibase.migrations (name, query)
		VALUES ($1, $2)
	`, name, sql)
	return err
}

func (s *MetaService) grantDashboardAccess(ctx context.Context, schema, table string) error {
	if schema == "" {
		schema = "public"
	}

	schemaIdent := pgx.Identifier{schema}.Sanitize()
	tableIdent := pgx.Identifier{table}.Sanitize()

	statements := []string{
		fmt.Sprintf("GRANT USAGE ON SCHEMA %s TO authenticated, anon", schemaIdent),
		fmt.Sprintf("GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE %s.%s TO authenticated, anon", schemaIdent, tableIdent),
		fmt.Sprintf("GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA %s TO authenticated, anon", schemaIdent),
		fmt.Sprintf("ALTER DEFAULT PRIVILEGES IN SCHEMA %s GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO authenticated, anon", schemaIdent),
		fmt.Sprintf("ALTER DEFAULT PRIVILEGES IN SCHEMA %s GRANT USAGE, SELECT ON SEQUENCES TO authenticated, anon", schemaIdent),
	}

	for _, statement := range statements {
		if _, err := s.db.Exec(ctx, statement); err != nil {
			return err
		}
	}

	return nil
}

func (s *MetaService) ensureDashboardAccess(ctx context.Context) error {
	rows, err := s.db.Query(ctx, `
		SELECT nspname
		FROM pg_namespace
		WHERE nspname NOT IN ('pg_catalog', 'information_schema', 'omnibase')
		  AND nspname NOT LIKE 'pg_toast%'
		  AND nspname NOT LIKE 'pg_temp_%'
		ORDER BY nspname
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var schemas []string
	for rows.Next() {
		var schema string
		if err := rows.Scan(&schema); err != nil {
			return err
		}
		schemas = append(schemas, schema)
	}
	if rows.Err() != nil {
		return rows.Err()
	}

	for _, schema := range schemas {
		schemaIdent := pgx.Identifier{schema}.Sanitize()
		statements := []string{
			fmt.Sprintf("GRANT USAGE ON SCHEMA %s TO authenticated, anon", schemaIdent),
			fmt.Sprintf("GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA %s TO authenticated, anon", schemaIdent),
			fmt.Sprintf("GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA %s TO authenticated, anon", schemaIdent),
			fmt.Sprintf("ALTER DEFAULT PRIVILEGES IN SCHEMA %s GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO authenticated, anon", schemaIdent),
			fmt.Sprintf("ALTER DEFAULT PRIVILEGES IN SCHEMA %s GRANT USAGE, SELECT ON SEQUENCES TO authenticated, anon", schemaIdent),
		}

		for _, statement := range statements {
			if _, err := s.db.Exec(ctx, statement); err != nil {
				return err
			}
		}
	}

	return nil
}

type tableRef struct {
	Schema string
	Table  string
}

var createTablePattern = regexp.MustCompile(`(?is)^create\s+table\s+(if\s+not\s+exists\s+)?(?:"?([a-zA-Z_][\w$]*)"?\.)?"?([a-zA-Z_][\w$]*)"?`)

func extractCreatedTableRef(sql string) *tableRef {
	match := createTablePattern.FindStringSubmatch(strings.TrimSpace(sql))
	if len(match) == 0 {
		return nil
	}

	schema := match[2]
	if schema == "" {
		schema = "public"
	}

	return &tableRef{
		Schema: schema,
		Table:  match[3],
	}
}

func queryReturnsRows(sql string) bool {
	upper := strings.ToUpper(strings.TrimSpace(sql))
	for _, prefix := range []string{"SELECT", "WITH", "SHOW", "EXPLAIN"} {
		if strings.HasPrefix(upper, prefix) {
			return true
		}
	}
	return false
}

func isSchemaMutation(sql string) bool {
	upper := strings.ToUpper(strings.TrimSpace(sql))
	for _, prefix := range []string{"CREATE", "ALTER", "DROP", "COMMENT", "GRANT", "REVOKE"} {
		if strings.HasPrefix(upper, prefix) {
			return true
		}
	}
	return false
}
func (s *MetaService) GetTableDDL(ctx context.Context, schema, table string) (string, error) {
	// Simple DDL generation logic
	// In a real production app, this would be more complex (handling indexes, triggers, etc.)
	// For now, we'll reconstruct the basic CREATE TABLE statement from columns
	columns, err := s.GetTableColumns(ctx, schema, table)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("-- Schema Blueprint for %s.%s\n", schema, table))
	sb.WriteString(fmt.Sprintf("CREATE TABLE %s.%s (\n", schema, table))

	for i, col := range columns {
		sb.WriteString(fmt.Sprintf("  %s %s", col.Name, col.Type))
		if !col.IsNullable {
			sb.WriteString(" NOT NULL")
		}
		if col.Default != "" {
			sb.WriteString(fmt.Sprintf(" DEFAULT %s", col.Default))
		}
		if i < len(columns)-1 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}

	// Add Primary Key if it exists
	pkQuery := `
		SELECT
			a.attname
		FROM pg_index i
		JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = ANY(i.indkey)
		WHERE i.indrelid = $1::regclass
		AND i.indisprimary;
	`
	rows, err := s.db.Query(ctx, pkQuery, fmt.Sprintf("%s.%s", schema, table))
	if err == nil {
		var pkCols []string
		for rows.Next() {
			var colName string
			if err := rows.Scan(&colName); err == nil {
				pkCols = append(pkCols, colName)
			}
		}
		rows.Close()
		if len(pkCols) > 0 {
			sb.WriteString(fmt.Sprintf("\n  , CONSTRAINT %s_pkey PRIMARY KEY (%s)\n", table, strings.Join(pkCols, ", ")))
		}
	}

	sb.WriteString(");")
	return sb.String(), nil
}

type TableStats struct {
	RowCount    int64  `json:"row_count"`
	TableSize   string `json:"table_size"`
	IndexSize   string `json:"index_size"`
	TotalSize   string `json:"total_size"`
	Description string `json:"description"`
}

func (s *MetaService) GetTableStats(ctx context.Context, schema, table string) (TableStats, error) {
	var stats TableStats
	query := `
		SELECT
			(SELECT n_live_tup FROM pg_stat_user_tables WHERE schemaname = $1 AND relname = $2) as row_count,
			pg_size_pretty(pg_table_size($3)) as table_size,
			pg_size_pretty(pg_indexes_size($3)) as index_size,
			pg_size_pretty(pg_total_relation_size($3)) as total_size,
			COALESCE(obj_description($3::regclass, 'pg_class'), '') as description;
	`
	fullTable := fmt.Sprintf("%s.%s", schema, table)
	err := s.db.QueryRow(ctx, query, schema, table, fullTable).Scan(
		&stats.RowCount, &stats.TableSize, &stats.IndexSize, &stats.TotalSize, &stats.Description,
	)
	return stats, err
}
