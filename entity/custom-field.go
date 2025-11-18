package entity

type CustomField struct {
	FieldName  string `json:"field_name" validate:"required"`
	FieldValue string `json:"field_value"`
}
