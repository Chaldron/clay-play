package group

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
)

type Store interface {
	Create(CreateParams) error
	Get(string) (Group, error)
	ListMembers(string) ([]GroupMember, error)
	List() ([]Group, error)
	GetByInvite(string) (Group, error)
	HasMember(string, string) (bool, error)
	AddMember(string, string) error
	Update(UpdateParams) error
	Delete(string) error
	RemoveMember(string, string) error
	RefreshInviteId(string, string) error
}

type store struct {
	db *sqlx.DB
}

func NewStore(db *sqlx.DB) *store {
	return &store{
		db: db,
	}
}

func storeLog(format string, s ...any) {
	log.Printf("group/store.go: %s", fmt.Sprintf(format, s...))
}

type Group struct {
	Id               string    `db:"id"`
	CreatedAt        time.Time `db:"created_at"`
	CreatorId        string    `db:"creator_id"`
	CreatorFullName  string    `db:"creator_full_name"`
	IsDeleted        bool      `db:"is_deleted"`
	Name             string    `db:"name"`
	InviteId         string    `db:"invite_id"`
	TotalMemberCount int       `db:"total_member_count"`
}

type GroupMember struct {
	GroupId      string    `db:"group_id"`
	UserId       string    `db:"user_id"`
	UserFullName string    `db:"user_full_name"`
	CreatedAt    time.Time `db:"created_at"`
}

type GroupDetailed struct {
	Group
	Members []GroupMember
}

type CreateParams struct {
	Id        string
	CreatorId string
	Name      string
	InviteId  string
}

func (s *store) Create(p CreateParams) error {
	stmt := `
        INSERT INTO user_group (id, created_at, creator_id, name, invite_id)
        VALUES (?, ?, ?, ?, ?)
    `
	args := []any{
		p.Id,
		time.Now().UTC(),
		p.CreatorId,
		p.Name,
		p.InviteId,
	}
	storeLog("Create args %v", args)

	_, err := s.db.Exec(stmt, args...)
	if err != nil {
		return err
	}

	return nil
}

func (s *store) Get(id string) (Group, error) {
	stmt := `
        SELECT ug.id, ug.name, ug.invite_id, ug.creator_id
            , u.full_name AS creator_full_name 
        FROM user_group ug
        INNER JOIN user u ON ug.creator_id = u.id
        WHERE ug.id = ? AND is_deleted = FALSE
    `
	args := []any{id}

	var g Group
	err := s.db.Get(&g, stmt, args...)
	return g, err
}

func (s *store) ListMembers(id string) ([]GroupMember, error) {
	stmt := `
        SELECT ugm.group_id, ugm.user_id, u.full_name AS user_full_name FROM user_group_member ugm
        INNER JOIN user u ON u.id = ugm.user_id
        WHERE group_id = ?
        ORDER BY ugm.created_at ASC
    `
	args := []any{id}

	var m []GroupMember
	err := s.db.Select(&m, stmt, args...)
	return m, err
}

func (s *store) List() ([]Group, error) {
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
	err := s.db.Select(&g, stmt)
	return g, err
}

func (s *store) GetByInvite(inviteId string) (Group, error) {
	stmt := `
        SELECT id, name FROM user_group
        WHERE invite_id = ? AND is_deleted = FALSE
    `
	args := []any{inviteId}

	var g Group
	err := s.db.Get(&g, stmt, args...)
	return g, err
}

func (s *store) HasMember(groupId string, userId string) (bool, error) {
	stmt := `
        SELECT 1 FROM user_group_member
        WHERE group_id = ? AND user_id = ?
    `
	args := []any{groupId, userId}

	var i int
	err := s.db.Get(&i, stmt, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}

func (s *store) AddMember(groupId string, userId string) error {
	stmt := `
        INSERT INTO user_group_member (group_id, user_id, created_at)
        VALUES (?, ?, ?)
    `
	args := []any{
		groupId,
		userId,
		time.Now().UTC(),
	}
	storeLog("AddMember args %v", args)

	_, err := s.db.Exec(stmt, args...)
	return err
}

type UpdateParams struct {
	Id   string
	Name string
}

func (s *store) Update(p UpdateParams) error {
	stmt := `
        UPDATE user_group
        SET name = ?
        WHERE id = ?
    `
	args := []any{p.Name, p.Id}
	storeLog("Update args %v", args)

	_, err := s.db.Exec(stmt, args...)
	return err
}

func (s *store) Delete(id string) error {
	stmt := `
        UPDATE user_group
        SET is_deleted = TRUE
        WHERE id = ?
    `
	args := []any{id}
	storeLog("Delete args %v", args)

	_, err := s.db.Exec(stmt, args...)
	return err
}

func (s *store) RemoveMember(groupId string, userId string) error {
	stmt := `
        DELETE FROM user_group_member
        WHERE group_id = ? AND user_id = ?
    `
	args := []any{groupId, userId}
	storeLog("RemoveMember args %v", args)

	_, err := s.db.Exec(stmt, args...)
	return err
}

func (s *store) RefreshInviteId(groupId string, newInviteId string) error {
	stmt := `
        UPDATE user_group
        SET invite_id = ?
        WHERE id = ?
    `
	args := []any{newInviteId, groupId}
	storeLog("RefreshInviteId args %v", args)

	_, err := s.db.Exec(stmt, args...)
	return err
}
