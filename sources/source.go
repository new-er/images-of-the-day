package sources 

type Source interface {
	GetImageLinks() ([]string, error)
}
