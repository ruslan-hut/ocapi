package entity

import "time"

type Category struct {
	CategoryId   int64     `json:"category_id,omitempty"`
	CategoryUID  int64     `json:"category_uid,omitempty"`
	Image        string    `json:"image,omitempty"`
	ParentId     int64     `json:"parent_id,omitempty"`
	ParentUID    int64     `json:"parent_uid,omitempty"`
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

	var top = 0
	if category.Menu {
		top = 1
	}

	return &Category{
		CategoryUID:  category.CategoryUID,
		ParentId:     0,
		ParentUID:    category.ParentUID,
		Top:          top,
		Column:       1,
		SortOrder:    category.SortOrder,
		Status:       status,
		DateAdded:    time.Now(),
		DateModified: time.Now(),
	}
}

//uidcategory uidparent(add if no) sortorder top status
// uidcat langid name description
