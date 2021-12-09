package client

type Account struct {
	Login string
	Balance float64 `json:",string"`
	Assets []UserAsset
}

type UserAsset struct {
	Name string
	Amount float64 `json:",string"`
}

func (acc *Account)GetAsset(name string) *UserAsset {
	for _, ass := range acc.Assets {
		if ass.Name == name {
			return &ass
		}
	}

	return nil
}