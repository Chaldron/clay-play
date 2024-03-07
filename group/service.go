package group

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

type service struct {
	db *sqlx.DB
}

func NewService(db *sqlx.DB) *service {
	return &service{
		db: db,
	}
}

func serviceLog(format string, s ...any) {
	log.Printf("group/service.go: %s", fmt.Sprintf(format, s...))
}

func (s *service) Get(id string) (Group, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return Group{}, err
	}
	defer tx.Rollback()

	g, err := get(tx, id)
	return g, err
}

func (s *service) GetDetailed(id string) (GroupDetailed, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return GroupDetailed{}, err
	}
	defer tx.Rollback()

	g, err := get(tx, id)
	if err != nil {
		return GroupDetailed{}, err
	}

	m, err := listMembers(tx, id)
	if err != nil {
		return GroupDetailed{}, err
	}

	return GroupDetailed{
		Group:   g,
		Members: m,
	}, err
}

func (s *service) List() ([]Group, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return []Group{}, err
	}
	defer tx.Rollback()

	g, err := list(tx)
	return g, err
}

type CreateParams struct {
	CreatorId string
	Name      string
}

func (s *service) CreateAndAddMember(p CreateParams) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	id, err := create(tx, p)
	if err != nil {
		return err
	}

	err = addMember(tx, id, p.CreatorId)

	return tx.Commit()
}

type UpdateParams struct {
	Id   string
	Name string
}

func (s *service) Update(p UpdateParams) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = update(tx, p)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *service) Delete(id string) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = delete(tx, id)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *service) AddMemberFromInvite(inviteId string, userId string) (Group, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return Group{}, err
	}
	defer tx.Rollback()

	g, err := getByInvite(tx, inviteId)
	if err != nil {
		return Group{}, err
	}

	err = addMember(tx, g.Id, userId)
	if err != nil {
		return Group{}, err
	}

	err = tx.Commit()
	if err != nil {
		return Group{}, err
	}

	return g, nil
}

func (s *service) RemoveMember(groupId string, userId string) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = removeMember(tx, groupId, userId)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *service) UserCanAccess(groupId sql.NullString, userId string) (bool, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	exists, err := hasMember(tx, groupId.String, userId)
	return exists, err
}

func (s *service) UserCanAccessError(groupId sql.NullString, userId string) error {
	ok, err := s.UserCanAccess(groupId, userId)
	if err != nil {
		return err
	}
	if !ok {
		return ErrNoAccess
	}

	return nil
}

func (s *service) RefreshInviteId(groupId string) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = refreshInviteId(tx, groupId)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func get(tx *sqlx.Tx, id string) (Group, error) {
	stmt := `
        SELECT ug.id, ug.name, ug.invite_id, ug.creator_id
            , u.full_name AS creator_full_name
        FROM user_group ug
        INNER JOIN user u ON ug.creator_id = u.id
        WHERE ug.id = ? AND is_deleted = FALSE
    `
	args := []any{id}

	var g Group
	err := tx.Get(&g, stmt, args...)
	return g, err
}

func listMembers(tx *sqlx.Tx, id string) ([]GroupMember, error) {
	stmt := `
        SELECT ugm.group_id, ugm.user_id, u.full_name AS user_full_name FROM user_group_member ugm
        INNER JOIN user u ON u.id = ugm.user_id
        WHERE group_id = ?
        ORDER BY ugm.created_at ASC
    `
	args := []any{id}

	var m []GroupMember
	err := tx.Select(&m, stmt, args...)
	return m, err
}

func list(tx *sqlx.Tx) ([]Group, error) {
	stmt := `
        SELECT
            ug.id, ug.name
            , COALESCE(ugc.total_member_count, 0) AS total_member_count
        FROM user_group AS ug
        LEFT JOIN (
            SELECT group_id, COUNT(*) AS total_member_count FROM user_group_member
            GROUP BY group_id
        ) AS ugc ON ug.id = ugc.group_id
        WHERE is_deleted = FALSE
        ORDER BY created_at ASC
    `

	var g []Group
	err := tx.Select(&g, stmt)
	return g, err
}

func getByInvite(tx *sqlx.Tx, inviteId string) (Group, error) {
	stmt := `
        SELECT id, name FROM user_group
        WHERE invite_id = ? AND is_deleted = FALSE
    `
	args := []any{inviteId}

	var g Group
	err := tx.Get(&g, stmt, args...)
	return g, err
}

func create(tx *sqlx.Tx, p CreateParams) (string, error) {
	id, err := gonanoid.New()
	if err != nil {
		return "", err
	}

	inviteId, err := gonanoid.New()
	if err != nil {
		return "", err
	}

	stmt := `
        INSERT INTO user_group (id, created_at, creator_id, name, invite_id)
        VALUES (?, ?, ?, ?, ?)
    `
	args := []any{
		id,
		time.Now().UTC(),
		p.CreatorId,
		p.Name,
		inviteId,
	}
	serviceLog("create args %v", args)

	_, err = tx.Exec(stmt, args...)
	if err != nil {
		return "", err
	}

	return id, nil
}

func addMember(tx *sqlx.Tx, groupId string, userId string) error {
	stmt := `
        INSERT INTO user_group_member (group_id, user_id, created_at)
        VALUES (?, ?, ?)
        ON CONFLICT (group_id, user_id) DO NOTHING 
    `
	args := []any{
		groupId,
		userId,
		time.Now().UTC(),
	}
	serviceLog("addMember args %v", args)

	_, err := tx.Exec(stmt, args...)
	return err
}

func update(tx *sqlx.Tx, p UpdateParams) error {
	stmt := `
        UPDATE user_group
        SET name = ?
        WHERE id = ?
    `
	args := []any{p.Name, p.Id}
	serviceLog("update args %v", args)

	_, err := tx.Exec(stmt, args...)
	return err
}

func delete(tx *sqlx.Tx, id string) error {
	stmt := `
        UPDATE user_group
        SET is_deleted = TRUE
        WHERE id = ?
    `
	args := []any{id}
	serviceLog("delete args %v", args)

	_, err := tx.Exec(stmt, args...)
	return err
}

func removeMember(tx *sqlx.Tx, groupId string, userId string) error {
	g, err := get(tx, groupId)
	if g.CreatorId == userId {
		return errors.New("cannot remove the creator of a group")
	}

	stmt := `
        DELETE FROM user_group_member
        WHERE group_id = ? AND user_id = ?
    `
	args := []any{groupId, userId}
	serviceLog("removeMember args %v", args)

	_, err = tx.Exec(stmt, args...)
	return err
}

func hasMember(tx *sqlx.Tx, groupId string, userId string) (bool, error) {
	stmt := `
        SELECT 1 FROM user_group_member
        WHERE group_id = ? AND user_id = ?
    `
	args := []any{groupId, userId}

	var i int
	err := tx.Get(&i, stmt, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}

func refreshInviteId(tx *sqlx.Tx, groupId string) error {
	inviteId, err := gonanoid.New()
	if err != nil {
		return err
	}

	stmt := `
        UPDATE user_group
        SET invite_id = ?
        WHERE id = ?
    `
	args := []any{inviteId, groupId}
	serviceLog("RefreshInviteId args %v", args)

	_, err = tx.Exec(stmt, args...)
	return err
}
