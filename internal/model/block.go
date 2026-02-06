package model

type Block struct {
	ID      int
	UserID  int
	TypeID  int
	Title   string
	Data    []byte
	Nonce   []byte
	Salt    []byte
	Profile string
	Type    *Type
}
