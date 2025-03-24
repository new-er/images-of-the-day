package sources 

type Source interface {
	SaveImages(destination string) error
}
