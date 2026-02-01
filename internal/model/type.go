package model

type TypeNameString string

const (
	TypeNameText        TypeNameString = "text"
	TypeNameCredentials TypeNameString = "credentials"
	TypeNameCard        TypeNameString = "bank_card"
	TypeNameFile        TypeNameString = "file"
)

type Type struct {
	ID          int
	TypeName    string
	Description string
}
