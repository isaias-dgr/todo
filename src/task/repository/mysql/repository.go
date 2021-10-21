package mysql

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/isaias-dgr/todo/src/domain"
	"go.uber.org/zap"
)

type taskRepository struct {
	Conn *sql.DB
	l    *zap.SugaredLogger
}

func NewtaskRepository(Conn *sql.DB, logger *zap.SugaredLogger) domain.TaskRepository {
	return &taskRepository{
		Conn: Conn,
		l:    logger,
	}
}

func (m *taskRepository) Fetch(ctx context.Context, f *domain.Filter) (ts *domain.Tasks, err error) {
	query := `ORDER BY created_at ASC LIMIT ? OFFSET ?`
	filter := []interface{}{f.Limit, f.Offset}

	tasks, err := m.fetch(ctx, query, filter)
	if err != nil {
		m.l.Error(err.Error())
		return nil, err
	}

	total, err := m.count(ctx)
	if err != nil {
		m.l.Error(err.Error())
		return nil, err
	}
	return domain.NewTasks(tasks, total), nil
}

func (m *taskRepository) fetch(ctx context.Context, stmt string, filters []interface{}) (ts []*domain.Task, err error) {
	tasks := []*domain.Task{}
	query := `SELECT * FROM task ` + stmt
	rows, err := m.Conn.QueryContext(ctx, query, filters...)
	if err != nil {
		m.l.Error(err.Error())
		return nil, errors.New("query_context")
	}
	defer rows.Close()

	for rows.Next() {
		task := &domain.Task{}
		err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.CreatedAt, &task.UpdatedAt)
		if err != nil {
			m.l.Error(err.Error())
			return nil, errors.New("row_data_types")
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		m.l.Error(err.Error())
		return nil, errors.New("row_corrupt")
	}

	return tasks, nil
}

func (m *taskRepository) count(ctx context.Context) (total int, err error) {
	query := `SELECT count(*) FROM task`
	rows, err := m.Conn.Query(query)
	if err != nil {
		m.l.Error(err.Error())
		return 0, errors.New("query_context")
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&total)
		if err != nil {
			m.l.Error(err.Error())
			return 0, errors.New("row_data_types")
		}
	}
	if err := rows.Err(); err != nil {
		m.l.Error(err.Error())
		return 0, errors.New("row_corrupt")
	}
	return total, nil
}


func (m *taskRepository) GetByID(ctx context.Context, id string) (t *domain.Task, err error) {
	_, binary_uuid, err := m.parse(id)
	if err != nil{
		return nil, err
	}

	tasks, err := m.fetch(ctx, `WHERE id=? `, []interface{}{binary_uuid})
	if err != nil {
		m.l.Error(err.Error())
		return nil, err
	}
	if len(tasks) == 0 {
		m.l.Error("Not Found")
		return nil, errors.New("not_found")
	}
	return tasks[0], nil
}

func (m *taskRepository) Insert(ctx context.Context, ta *domain.Task) (err error) {
	created_at := time.Now()
	ta.ID = uuid.New()
	binary_uuid, err := ta.ID.MarshalBinary()
	if err != nil {
		return errors.New("uuid_generate")
	}
	ta.CreatedAt = &created_at
	ta.UpdatedAt = ta.CreatedAt
	
	query := `INSERT task SET 
		id=?,
		title=?, 
		description=?,
		created_at=?,
		updated_at=?`

	stmt, err := m.Conn.PrepareContext(ctx, query)
	if err != nil {
		m.l.Error(err.Error())
		return errors.New("query_prepare_ctx")
	}

	res, err := stmt.ExecContext(ctx,
		binary_uuid, ta.Title, ta.Description, ta.CreatedAt, ta.UpdatedAt)
	if err != nil {
		m.l.Error(err.Error())
		return errors.New("query_exec")
	}
	affect, err := res.RowsAffected()
	if err != nil {
		m.l.Error(err.Error())
		return errors.New("query_exec")
	}
	if affect != 1 {
		m.l.Errorf("Weird  Behavior. Total Affected: %d", affect)
		return errors.New("conflict_insert")
	}
	return
}

func (m *taskRepository) Update(ctx context.Context, id string, ta *domain.Task) (err error) {
	raw_uuid, binary_uuid, err := m.parse(id)
	if err != nil{
		return err
	}
	updated_at := time.Now()
	ta.ID = *raw_uuid
	ta.UpdatedAt = &updated_at

	query := `UPDATE task set title=?, description=?, updated_at=? WHERE ID = ?`
	stmt, err := m.Conn.PrepareContext(ctx, query)
	if err != nil {
		m.l.Error(err.Error())
		return errors.New("query_prepare_ctx")
	}

	res, err := stmt.ExecContext(ctx, ta.Title, ta.Description, ta.UpdatedAt, binary_uuid)
	if err != nil {
		m.l.Error(err.Error())
		return errors.New("query_exec")
	}
	affect, err := res.RowsAffected()
	if err != nil {
		m.l.Error(err.Error())
		return errors.New("not_found")
	}
	if affect != 1 {
		m.l.Errorf("Weird  Behavior. Total Affected: %d", affect)
		return errors.New("conflict_update")
	}
	return
}

func (m *taskRepository) Delete(ctx context.Context, id string) (err error) {
	query := "DELETE FROM task WHERE id=?"
	_, binary_uuid, err := m.parse(id)
	if err != nil{
		return err
	}
	
	stmt, err := m.Conn.PrepareContext(ctx, query)
	if err != nil {
		m.l.Error(err.Error())
		return errors.New("query_prepare_ctx")
	}

	res, err := stmt.ExecContext(ctx, binary_uuid)
	if err != nil {
		m.l.Error(err.Error())
		return errors.New("query_exec")
	}

	rowsAfected, err := res.RowsAffected()
	if err != nil {
		m.l.Error(err.Error())
		return errors.New("query_exec_delete")
	}

	if rowsAfected != 1 {
		m.l.Errorf("Weird  Behavior. Total Affected: %d", rowsAfected)
		return errors.New("conflict_delete")
	}
	return
}

func (m *taskRepository) parse(id string) (*uuid.UUID, []byte, error) {
	raw_uuid, err := uuid.Parse(id)
	if err != nil {
		m.l.Error(err.Error())
		return nil,nil, errors.New("uuid_format")
	}

	binary_uuid, err := raw_uuid.MarshalBinary()
	if err != nil {
		m.l.Error(err.Error())
		return nil, nil, errors.New("uuid_format")
	}

	return &raw_uuid, binary_uuid, err
}