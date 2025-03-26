package sources 

type Source interface {
	GetName() string
	GetImageLinks() ([]string, error)
}
