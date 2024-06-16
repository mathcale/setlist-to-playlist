package strategies

type PersistenceStrategyInterface interface {
	Read() ([]byte, error)
	Write(data []byte) error
}
