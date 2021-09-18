package column

type Selector interface {
	Select([]string) ([]string, error)
}
