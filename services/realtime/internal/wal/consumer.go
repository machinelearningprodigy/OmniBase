package wal

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pglogrepl"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgproto3"
	"go.uber.org/zap"
)

// ChangeEvent represents a database change captured from WAL
type ChangeEvent struct {
	Schema    string         `json:"schema"`
	Table     string         `json:"table"`
	EventType string         `json:"eventType"` // "INSERT" | "UPDATE" | "DELETE"
	OldRecord map[string]any `json:"old_record,omitempty"`
	NewRecord map[string]any `json:"record,omitempty"`
	Errors    []string       `json:"errors,omitempty"`
}

// Consumer listens to Postgres WAL logical replication and emits ChangeEvents
type Consumer struct {
	conn    *pgconn.PgConn
	slot    string
	pubName string
	log     *zap.Logger
	events  chan ChangeEvent
}

// NewConsumer creates a WAL consumer that connects to Postgres replication
func NewConsumer(dsn, slot, pubName string, log *zap.Logger) (*Consumer, error) {
	// Connect in replication mode
	replicationDSN := dsn + " replication=database"
	conn, err := pgconn.Connect(context.Background(), replicationDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to connect for replication: %w", err)
	}

	return &Consumer{
		conn:    conn,
		slot:    slot,
		pubName: pubName,
		log:     log,
		events:  make(chan ChangeEvent, 1000), // Buffered channel for backpressure
	}, nil
}

// Start begins consuming WAL and emitting events on the Events channel
func (c *Consumer) Start(ctx context.Context) error {
	// Create replication slot if it doesn't exist
	_, err := pglogrepl.CreateReplicationSlot(ctx, c.conn, c.slot, "pgoutput",
		pglogrepl.CreateReplicationSlotOptions{Temporary: false, SnapshotAction: "NOEXPORT_SNAPSHOT"})
	if err != nil {
		// Slot likely already exists — that's fine
		c.log.Debug("replication slot may already exist", zap.String("slot", c.slot), zap.Error(err))
	}

	// Start logical replication
	err = pglogrepl.StartReplication(ctx, c.conn, c.slot, 0,
		pglogrepl.StartReplicationOptions{
			PluginArgs: []string{
				"proto_version '1'",
				fmt.Sprintf("publication_names '%s'", c.pubName),
			},
		})
	if err != nil {
		return fmt.Errorf("failed to start replication: %w", err)
	}

	c.log.Info("WAL consumer started",
		zap.String("slot", c.slot),
		zap.String("publication", c.pubName),
	)

	go c.consumeLoop(ctx)
	return nil
}

// Events returns the channel of database change events
func (c *Consumer) Events() <-chan ChangeEvent {
	return c.events
}

func (c *Consumer) consumeLoop(ctx context.Context) {
	defer close(c.events)
	defer c.conn.Close(ctx)

	relations := map[uint32]*pglogrepl.RelationMessage{}

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		rawMsg, err := c.conn.ReceiveMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.log.Error("WAL receive error", zap.Error(err))
			return
		}

		if errMsg, ok := rawMsg.(*pgproto3.ErrorResponse); ok {
			c.log.Error("WAL error from server", zap.String("msg", errMsg.Message))
			continue
		}

		msg, ok := rawMsg.(*pgproto3.CopyData)
		if !ok {
			continue
		}

		if msg.Data[0] == pglogrepl.PrimaryKeepaliveMessageByteID {
			// Send standby status update to prevent WAL accumulation
			continue
		}

		if msg.Data[0] != pglogrepl.XLogDataByteID {
			continue
		}

		xld, err := pglogrepl.ParseXLogData(msg.Data[1:])
		if err != nil {
			c.log.Error("failed to parse XLogData", zap.Error(err))
			continue
		}

		logicalMsg, err := pglogrepl.Parse(xld.WALData)
		if err != nil {
			c.log.Error("failed to parse logical msg", zap.Error(err))
			continue
		}

		switch m := logicalMsg.(type) {
		case *pglogrepl.RelationMessage:
			relations[m.RelationID] = m

		case *pglogrepl.InsertMessage:
			rel, ok := relations[m.RelationID]
			if !ok {
				continue
			}
			event := ChangeEvent{
				Schema:    rel.Namespace,
				Table:     rel.RelationName,
				EventType: "INSERT",
				NewRecord: decodeRow(rel, m.Tuple),
			}
			c.events <- event

		case *pglogrepl.UpdateMessage:
			rel, ok := relations[m.RelationID]
			if !ok {
				continue
			}
			event := ChangeEvent{
				Schema:    rel.Namespace,
				Table:     rel.RelationName,
				EventType: "UPDATE",
				NewRecord: decodeRow(rel, m.NewTuple),
			}
			if m.OldTuple != nil {
				event.OldRecord = decodeRow(rel, m.OldTuple)
			}
			c.events <- event

		case *pglogrepl.DeleteMessage:
			rel, ok := relations[m.RelationID]
			if !ok {
				continue
			}
			event := ChangeEvent{
				Schema:    rel.Namespace,
				Table:     rel.RelationName,
				EventType: "DELETE",
				OldRecord: decodeRow(rel, m.OldTuple),
			}
			c.events <- event
		}
	}
}

// decodeRow converts a pglogrepl tuple to a map
func decodeRow(rel *pglogrepl.RelationMessage, tuple *pglogrepl.TupleData) map[string]any {
	if tuple == nil {
		return nil
	}
	row := make(map[string]any, len(rel.Columns))
	for i, col := range tuple.Columns {
		if i >= len(rel.Columns) {
			break
		}
		colName := rel.Columns[i].Name
		switch col.DataType {
		case 'n': // null
			row[colName] = nil
		case 'u': // unchanged toast
			row[colName] = "__unchanged_toast__"
		case 't': // text
			row[colName] = string(col.Data)
		}
	}
	return row
}

// MarshalJSON serializes a ChangeEvent for WebSocket delivery
func (e ChangeEvent) MarshalJSON() ([]byte, error) {
	type Alias ChangeEvent
	return json.Marshal(Alias(e))
}
