package app

import (
	"database/sql"
	"dbmx/model"
	"encoding/json"
)

type Tabs struct {
	DB *sql.DB
}

func NewTabs(db *sql.DB) *Tabs {
	return &Tabs{
		DB: db,
	}
}

func (t *Tabs) AddTab(activeDBID, activeDB, activeDBColour string) (*model.Tab, error) {
	var active_db_id, active_db, active_db_colour *string
	if activeDBID != "" {
		active_db_id = &activeDBID
	}
	if activeDB != "" {
		active_db = &activeDB
	}
	if activeDBColour != "" {
		active_db_colour = &activeDBColour
	}

	// Insert a new active tab
	query := `INSERT INTO tabs (name, editor, output, is_active, active_db_id, active_db, active_db_colour) VALUES ('SQL Editor', '', '', true, ?, ?, ?);`
	result, err := t.DB.Exec(query, active_db_id, active_db, active_db_colour)
	if err != nil {
		return nil, err
	}
	insertedID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Write an update query to set is_active to false for all other tabs
	updateQuery := `UPDATE tabs SET is_active = false WHERE id != ?`
	_, err = t.DB.Exec(updateQuery, insertedID)
	if err != nil {
		return nil, err
	}

	return &model.Tab{
		ID:            insertedID,
		Name:          "SQL Editor",
		Editor:        "",
		Output:        "",
		IsActive:      true,
		ActiveDBID:    active_db_id,
		ActiveDB:      active_db,
		ActiveDBColor: active_db_colour,
	}, nil
}

func (t *Tabs) SetActiveTab(id int64) (*model.Tab, error) {
	// Write an update query to set is_active to false for all other tabs
	updateQuery := `UPDATE tabs SET is_active = false WHERE id != ?`
	_, err := t.DB.Exec(updateQuery, id)
	if err != nil {
		return nil, err
	}

	// Write an update query to set is_active to true for the given tab
	updateQuery = `UPDATE tabs SET is_active = true WHERE id = ? RETURNING *`

	var tab model.Tab
	err = t.DB.QueryRow(updateQuery, id).Scan(&tab.ID, &tab.Name, &tab.Editor, &tab.Output, &tab.IsActive, &tab.ActiveDBID, &tab.ActiveDB, &tab.ActiveDBColor)
	if err != nil {
		return nil, err
	}

	if tab.Output == "" {
		return &tab, nil
	}

	var output model.Output
	err = json.Unmarshal([]byte(tab.Output), &output)
	if err != nil {
		return nil, err
	}

	tab.Columns = output.Columns
	tab.Rows = output.Rows

	return &tab, nil
}

func (t *Tabs) GetAllTabs() ([]model.Tab, error) {
	// Query for all tabs
	query := `SELECT id, name, editor, output, is_active, active_db_id, active_db, active_db_colour FROM tabs`
	rows, err := t.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tabs []model.Tab
	for rows.Next() {
		var tab model.Tab
		err := rows.Scan(&tab.ID, &tab.Name, &tab.Editor, &tab.Output, &tab.IsActive, &tab.ActiveDBID, &tab.ActiveDB, &tab.ActiveDBColor)
		if err != nil {
			return nil, err
		}

		if tab.IsActive {
			var output model.Output
			err = json.Unmarshal([]byte(tab.Output), &output)
			if err != nil {
				return nil, err
			}

			tab.Columns = output.Columns
			tab.Rows = output.Rows
		}

		tabs = append(tabs, tab)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tabs, nil
}

func (t *Tabs) DeleteTab(id int64) (*model.Tab, error) {
	// Check if the tab is active
	query := `SELECT is_active FROM tabs WHERE id = ?`
	var isActive bool
	err := t.DB.QueryRow(query, id).Scan(&isActive)
	if err != nil {
		return nil, err
	}

	// Tab to be returned
	var tab model.Tab

	var isLastTab bool

	// If the tab is active, set the first tab as active where id != id and return that tab
	if isActive {
		// Check for the count of tabs to find if it's the last tab
		query = `SELECT COUNT(*) FROM tabs`
		var count int64
		err := t.DB.QueryRow(query).Scan(&count)
		if err != nil {
			return nil, err
		}

		// If it's not the last tab, set the first tab as active
		if count > 1 {
			query = `UPDATE tabs SET is_active = true WHERE id = (SELECT id FROM tabs WHERE id != ? LIMIT 1) RETURNING *`
			row := t.DB.QueryRow(query, id)

			err = row.Scan(&tab.ID, &tab.Name, &tab.Editor, &tab.Output, &tab.IsActive, &tab.ActiveDBID, &tab.ActiveDB, &tab.ActiveDBColor)
			if err != nil {
				return nil, err
			}

			var output model.Output
			err = json.Unmarshal([]byte(tab.Output), &output)
			if err != nil {
				return nil, err
			}

			tab.Columns = output.Columns
			tab.Rows = output.Rows
		} else {
			isLastTab = true
		}
	}

	// Delete the tab
	query = `DELETE FROM tabs WHERE id = ?`
	_, err = t.DB.Exec(query, id)
	if err != nil {
		return nil, err
	}

	if isActive && !isLastTab {
		return &tab, nil
	}

	return nil, nil
}

func (t *Tabs) UpdateTabEditorContent(id int64, editor string) error {
	query := `UPDATE tabs SET editor = ? WHERE id = ?`
	_, err := t.DB.Exec(query, editor, id)
	if err != nil {
		return err
	}

	return nil
}

func (t *Tabs) SaveActiveDBProps(id int64, activeDBID, activeDB, activeDBColour string) error {
	var active_db_id, active_db, active_db_colour *string
	if activeDBID != "" {
		active_db_id = &activeDBID
	}
	if activeDB != "" {
		active_db = &activeDB
	}
	if activeDBColour != "" {
		active_db_colour = &activeDBColour
	}

	query := `UPDATE tabs SET active_db_id = ?, active_db = ?, active_db_colour = ? WHERE id = ?`
	_, err := t.DB.Exec(query, active_db_id, active_db, active_db_colour, id)
	if err != nil {
		return err
	}

	return nil
}
