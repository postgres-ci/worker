package common

type Build struct {
	WorkingDir    string `db:"-"`
	BuildID       int64  `db:"build_id"`
	Branch        string `db:"branch"`
	Revision      string `db:"revision"`
	RepositoryURL string `db:"repository_url"`
	Config        struct {
		Scripts []string
	}
}

func (b *Build) LoadConfig() error {

	b.Config.Scripts = []string{"a", "b", "c"}

	return nil
}
