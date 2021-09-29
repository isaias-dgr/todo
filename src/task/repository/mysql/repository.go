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
	query := `ORDER BY created_at DESC LIMIT ? OFFSET ?`
	filter := []interface{}{f.Limit, f.Offset}
	return m.fetch(ctx, query, filter)
}

func (m *taskRepository) fetch(ctx context.Context, stmt string, filters []interface{}) (ts *domain.Tasks, err error) {
	var tasks []domain.Task
	query := `SELECT * FROM task ` + stmt
	rows, err := m.Conn.Query(query, filters...)
	if err != nil {
		m.l.Error(err.Error())
		return nil, errors.New("not_found")
	}
	defer rows.Close()

	for rows.Next() {
		task := domain.Task{}
		err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.CreatedAt, &task.UpdatedAt)
		if err != nil {
			m.l.Error(err.Error())
			return nil, errors.New("not_found")
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		m.l.Error(err.Error())
		return nil, errors.New("not_found")
	}

	total, err := m.count(ctx)
	if err != nil {
		m.l.Error(err.Error())
		return nil, errors.New("not_found")
	}
	return domain.NewTasks(tasks, total), nil
}

func (m *taskRepository) count(ctx context.Context) (total int, err error) {
	query := `SELECT count(*) FROM task`
	rows, err := m.Conn.Query(query)
	if err != nil {
		m.l.Error(err.Error())
		return 0, errors.New("not_found")
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&total)
		if err != nil {
			m.l.Error(err.Error())
			return 0, errors.New("not_found")
		}
	}
	if err := rows.Err(); err != nil {
		m.l.Error(err.Error())
		return 0, errors.New("not_found")
	}
	return total, nil
}

func (m *taskRepository) GetByID(ctx context.Context, id string) (t *domain.Task, err error) {
	raw_uuid, err := uuid.Parse(id)
	if err != nil {
		m.l.Error(err.Error())
		return nil, errors.New("not_found")
	}

	binary_uuid, err := raw_uuid.MarshalBinary()
	if err != nil {
		m.l.Error(err.Error())
		return nil, errors.New("not_found")
	}
	tasks, err := m.fetch(ctx, `WHERE id=? `, []interface{}{binary_uuid})
	if err != nil {
		m.l.Error(err.Error())
		return nil, errors.New("not_found")
	}
	if len(tasks.Data) == 0 {
		m.l.Error(err.Error())
		return nil, errors.New("not_found")
	}
	return &tasks.Data[0], nil
}

func (m *taskRepository) Update(ctx context.Context, id string, ta *domain.Task) (err error) {
	raw_uuid, err := uuid.Parse(id)
	if err != nil {
		m.l.Error(err.Error())
		return errors.New("not_found")
	}

	binary_uuid, err := raw_uuid.MarshalBinary()
	if err != nil {
		m.l.Error(err.Error())
		return errors.New("not_found")
	}

	updated_at := time.Now()
	ta.ID = raw_uuid
	ta.UpdatedAt = &updated_at

	query := `UPDATE task set title=?, description=?, updated_at=? WHERE ID = ?`
	stmt, err := m.Conn.PrepareContext(ctx, query)
	if err != nil {
		m.l.Error(err.Error())
		return errors.New("not_found")
	}

	res, err := stmt.ExecContext(ctx, ta.Title, ta.Description, ta.UpdatedAt, binary_uuid)
	if err != nil {
		m.l.Error(err.Error())
		return errors.New("not_found")
	}
	affect, err := res.RowsAffected()
	if err != nil {
		m.l.Error(err.Error())
		return errors.New("not_found")
	}
	if affect != 1 {
		m.l.Errorf("Weird  Behavior. Total Affected: %d", affect)
		return errors.New("conflict")
	}
	return
}

func (m *taskRepository) Insert(ctx context.Context, ta *domain.Task) (err error) {
	created_at := time.Now()
	ta.ID = uuid.New()
	ta.CreatedAt = &created_at
	ta.UpdatedAt = ta.CreatedAt
	binary_uuid, err := ta.ID.MarshalBinary()
	if err != nil {
		return
	}

	query := `INSERT task SET 
		id=?,
		title=?, 
		description=?,
		created_at=?,
		updated_at=?`

	stmt, err := m.Conn.PrepareContext(ctx, query)
	if err != nil {
		m.l.Error(err.Error())
		return
	}

	_, err = stmt.ExecContext(ctx,
		binary_uuid, ta.Title, ta.Description, ta.CreatedAt, ta.UpdatedAt)
	if err != nil {
		m.l.Error(err.Error())
		return
	}
	return
}

func (m *taskRepository) Delete(ctx context.Context, id string) (err error) {
	query := "DELETE FROM task WHERE id=?"

	raw_uuid, err := uuid.Parse(id)
	if err != nil {
		m.l.Error(err.Error())
		return errors.New("not_found")
	}

	binary_uuid, err := raw_uuid.MarshalBinary()
	if err != nil {
		m.l.Error(err.Error())
		return errors.New("not_found")
	}

	stmt, err := m.Conn.PrepareContext(ctx, query)
	if err != nil {
		m.l.Error(err.Error())
		return errors.New("not_found")
	}

	res, err := stmt.ExecContext(ctx, binary_uuid)
	if err != nil {
		m.l.Error(err.Error())
		return errors.New("not_found")
	}

	rowsAfected, err := res.RowsAffected()
	if err != nil {
		m.l.Error(err.Error())
		return errors.New("server_error")
	}

	if rowsAfected != 1 {
		m.l.Errorf("Weird  Behavior. Total Affected: %d", rowsAfected)
		return errors.New("conflict")
	}
	return
}
