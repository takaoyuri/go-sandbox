package address

import (
	a "github.com/inouet/ken-all/address"
)

type Address struct {
	Zip  string
	Pref string
	Town string
	City string
}

func NewAddress(row a.Row) Address {
	return Address{
		Zip:  row.Zip7,
		Pref: row.Pref,
		City: row.City,
		Town: row.Town,
	}
}
