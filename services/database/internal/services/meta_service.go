package services

import (
	"context"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type MetaService struct {
	db  *pgx.Conn
	log *zap.Logger
}

func NewMetaService(db *pgx.Conn, log *zap.Logger) *MetaService {
	return &MetaService{
		db:  db,
		log: log,
	}
}

type TableMeta struct {
	Name     string `json:"name"`
	Schema   string `json:"schema"`
	RowCount int64  `json:"row_count"`
	Size     string `json:"size"`
	HasRLS   bool   `json:"has_rls"`
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

// RunQuery executes an arbitrary SQL string and returns the raw rows
// In OmniBase, this requires Service Role token (Admin only).
func (s *MetaService) RunQuery(ctx context.Context, sql string) ([]map[string]interface{}, error) {
	rows, err := s.db.Query(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	
	fields := rows.FieldDescriptions()

	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, err
		}
		
		rowMap := make(map[string]interface{})
		for i, field := range fields {
			rowMap[string(field.Name)] = values[i]
		}
		results = append(results, rowMap)
	}
	
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return results, nil
}
