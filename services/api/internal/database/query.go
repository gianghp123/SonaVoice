package database

import "gorm.io/gorm"

type Query struct {
	Filters map[string]interface{}
	Page    int
	Limit   int
	OrderBy string
}

func NewQuery() *Query {
	return &Query{
		Filters: make(map[string]interface{}),
		Page:    1,
		Limit:   10,
	}
}

func (q *Query) SetFilter(key string, value interface{}) *Query {
	q.Filters[key] = value
	return q
}

func (q *Query) SetPage(page int) *Query {
	if page > 0 {
		q.Page = page
	}
	return q
}

func (q *Query) SetLimit(limit int) *Query {
	if limit > 0 && limit <= 100 {
		q.Limit = limit
	}
	return q
}

func (q *Query) SetOrderBy(order string) *Query {
	q.OrderBy = order
	return q
}

func (q *Query) Count(tx *gorm.DB) (int64, error) {
	var total int64
	if err := tx.Where(q.Filters).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (q *Query) Apply(tx *gorm.DB) *gorm.DB {
	tx = tx.Where(q.Filters)
	if q.OrderBy != "" {
		tx = tx.Order(q.OrderBy)
	}
	offset := (q.Page - 1) * q.Limit
	return tx.Offset(offset).Limit(q.Limit)
}
