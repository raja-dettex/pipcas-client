package service

type PipCasDbService interface {
	CreateSchema() error
	Save(string, string) error
	GetAllBy(string) (*[]string, error)
	GetBy(string) (string, error)
}
