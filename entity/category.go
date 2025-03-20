package entity

import "time"

type Category struct {
	CategoryId   int64     `json:"category_id,omitempty"`
	CategoryUID  string    `json:"category_uid,omitempty"`
	Image        string    `json:"image,omitempty"`
	ParentId     int64     `json:"parent_id,omitempty"`
	ParentUID    string    `json:"parent_uid,omitempty"`
	Top          int       `json:"top,omitempty"`
	Column       int       `json:"column,omitempty"`
	SortOrder    int       `json:"sort_order,omitempty"`
	Status       int       `json:"status,omitempty"`
	DateAdded    time.Time `json:"date_added,omitempty"`
	DateModified time.Time `json:"date_modified,omitempty"`
}

func CategoryFromCategoryData(category *CategoryData) *Category {
	var status = 0
	if category.Active {
		status = 1
	}

	return &Category{
		CategoryUID:  category.CategoryUID,
		ParentId:     0,
		ParentUID:    category.ParentUID,
		Top:          category.Top,
		Column:       1,
		SortOrder:    category.SortOrder,
		Status:       status,
		DateModified: time.Now(),
	}
}
