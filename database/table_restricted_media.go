package database

import (
	"database/sql"
	"errors"

	"github.com/t2bot/matrix-media-repo/common/rcontext"
)

type RestrictedCondition string

const RestrictedRequiresAuth RestrictedCondition = "io.t2bot.requires_authentication" // Internal extension

type DbRestrictedMedia struct {
	Origin         string
	MediaId        string
	Condition      RestrictedCondition
	ConditionValue string
}

// restricted_media 存储媒体访问限制规则（condition_type + condition_value）。
// 当前主要用于 requires_authentication：下载/缩略图管线会先检查该表，未授权时拒绝访问。
// 这是一张围绕 media 的策略扩展表，不保存媒体本体，仅定义读取时的鉴权条件。
const insertRestrictedMedia = "INSERT INTO restricted_media (origin, media_id, condition_type, condition_value) VALUES ($1, $2, $3, $4);"
const updateRestrictedMedia = "UPDATE restricted_media SET condition_type = $3, condition_value = $4 WHERE origin = $1 AND media_id = $2;"
const selectRestrictedMedia = "SELECT origin, media_id, condition_type, condition_value FROM restricted_media WHERE origin = $1 AND media_id = $2;"

type restrictedMediaTableStatements struct {
	insertRestrictedMedia *sql.Stmt
	updateRestrictedMedia *sql.Stmt
	selectRestrictedMedia *sql.Stmt
}

type restrictedMediaTableWithContext struct {
	statements *restrictedMediaTableStatements
	ctx        rcontext.RequestContext
}

func prepareRestrictedMediaTables(db *sql.DB) (*restrictedMediaTableStatements, error) {
	var err error
	var stmts = &restrictedMediaTableStatements{}

	if stmts.insertRestrictedMedia, err = db.Prepare(insertRestrictedMedia); err != nil {
		return nil, errors.New("error preparing insertRestrictedMedia: " + err.Error())
	}
	if stmts.updateRestrictedMedia, err = db.Prepare(updateRestrictedMedia); err != nil {
		return nil, errors.New("error preparing updateRestrictedMedia: " + err.Error())
	}
	if stmts.selectRestrictedMedia, err = db.Prepare(selectRestrictedMedia); err != nil {
		return nil, errors.New("error preparing selectRestrictedMedia: " + err.Error())
	}

	return stmts, nil
}

func (s *restrictedMediaTableStatements) Prepare(ctx rcontext.RequestContext) *restrictedMediaTableWithContext {
	return &restrictedMediaTableWithContext{
		statements: s,
		ctx:        ctx,
	}
}

func (s *restrictedMediaTableWithContext) Insert(origin string, mediaId string, condition RestrictedCondition, conditionValue string) error {
	_, err := s.statements.insertRestrictedMedia.ExecContext(s.ctx, origin, mediaId, condition, conditionValue)
	return err
}

func (s *restrictedMediaTableWithContext) Update(origin string, mediaId string, condition RestrictedCondition, conditionValue string) error {
	_, err := s.statements.updateRestrictedMedia.ExecContext(s.ctx, origin, mediaId, condition, conditionValue)
	return err
}

func (s *restrictedMediaTableWithContext) GetAllForId(origin string, mediaId string) ([]*DbRestrictedMedia, error) {
	results := make([]*DbRestrictedMedia, 0)
	rows, err := s.statements.selectRestrictedMedia.QueryContext(s.ctx, origin, mediaId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return results, nil
		}
		return nil, err
	}
	for rows.Next() {
		val := &DbRestrictedMedia{}
		if err = rows.Scan(&val.Origin, &val.MediaId, &val.Condition, &val.ConditionValue); err != nil {
			return nil, err
		}
		results = append(results, val)
	}
	return results, nil
}
