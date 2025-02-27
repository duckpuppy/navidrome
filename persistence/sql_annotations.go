package persistence

import (
	"time"

	. "github.com/Masterminds/squirrel"
	"github.com/astaxie/beego/orm"
	"github.com/deluan/navidrome/log"
	"github.com/deluan/navidrome/model"
	"github.com/google/uuid"
)

type annotation struct {
	AnnID     string    `json:"annID"        orm:"pk;column(ann_id)"`
	UserID    string    `json:"userID"       orm:"pk;column(user_id)"`
	ItemID    string    `json:"itemID"       orm:"pk;column(item_id)"`
	ItemType  string    `json:"itemType"`
	PlayCount int       `json:"playCount"`
	PlayDate  time.Time `json:"playDate"`
	Rating    int       `json:"rating"`
	Starred   bool      `json:"starred"`
	StarredAt time.Time `json:"starredAt"`
}

const annotationTable = "annotation"

func (r sqlRepository) newSelectWithAnnotation(idField string, options ...model.QueryOptions) SelectBuilder {
	return r.newSelect(options...).
		LeftJoin("annotation on ("+
			"annotation.item_id = "+idField+
			" AND annotation.item_type = '"+r.tableName+"'"+
			" AND annotation.user_id = '"+userId(r.ctx)+"')").
		Columns("starred", "starred_at", "play_count", "play_date", "rating")
}

func (r sqlRepository) annId(itemID ...string) And {
	return And{
		Eq{"user_id": userId(r.ctx)},
		Eq{"item_type": r.tableName},
		Eq{"item_id": itemID},
	}
}

func (r sqlRepository) annUpsert(values map[string]interface{}, itemIDs ...string) error {
	upd := Update(annotationTable).Where(r.annId(itemIDs...))
	for f, v := range values {
		upd = upd.Set(f, v)
	}
	c, err := r.executeSQL(upd)
	if c == 0 || err == orm.ErrNoRows {
		for _, itemID := range itemIDs {
			id, _ := uuid.NewRandom()
			values["ann_id"] = id.String()
			values["user_id"] = userId(r.ctx)
			values["item_type"] = r.tableName
			values["item_id"] = itemID
			ins := Insert(annotationTable).SetMap(values)
			_, err = r.executeSQL(ins)
			if err != nil {
				return err
			}
		}
	}
	return err
}

func (r sqlRepository) SetStar(starred bool, ids ...string) error {
	starredAt := time.Now()
	return r.annUpsert(map[string]interface{}{"starred": starred, "starred_at": starredAt}, ids...)
}

func (r sqlRepository) SetRating(rating int, itemID string) error {
	return r.annUpsert(map[string]interface{}{"rating": rating}, itemID)
}

func (r sqlRepository) IncPlayCount(itemID string, ts time.Time) error {
	upd := Update(annotationTable).Where(r.annId(itemID)).
		Set("play_count", Expr("play_count+1")).
		Set("play_date", ts)
	c, err := r.executeSQL(upd)

	if c == 0 || err == orm.ErrNoRows {
		id, _ := uuid.NewRandom()
		values := map[string]interface{}{}
		values["ann_id"] = id.String()
		values["user_id"] = userId(r.ctx)
		values["item_type"] = r.tableName
		values["item_id"] = itemID
		values["play_count"] = 1
		values["play_date"] = ts
		ins := Insert(annotationTable).SetMap(values)
		_, err = r.executeSQL(ins)
		if err != nil {
			return err
		}
	}
	return err
}

func (r sqlRepository) cleanAnnotations() error {
	del := Delete(annotationTable).Where(Eq{"item_type": r.tableName}).Where("item_id not in (select id from " + r.tableName + ")")
	c, err := r.executeSQL(del)
	if err != nil {
		return err
	}
	if c > 0 {
		log.Debug(r.ctx, "Clean-up annotations", "table", r.tableName, "totalDeleted", c)
	}
	return nil
}
