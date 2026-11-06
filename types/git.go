package types

// GitChanges represents counts of additions and deletions in Git working tree.
type GitChanges struct {
	Insertions int
	Deletions  int
}

// GitAheadBehind represents ahead and behind commit counts relative to upstream.
type GitAheadBehind struct {
	Ahead  int
	Behind int
}

// GitDataProvider abstracts retrieval of Git repository state.
// Methods should be called lazily when the corresponding widget requires data.
type GitDataProvider interface {
	GetBranch(ctx RenderContext) (branch string, dirty bool, err error)
	GetChanges(ctx RenderContext) (GitChanges, error)
	GetAheadBehind(ctx RenderContext) (GitAheadBehind, error)
	GetRepoName(ctx RenderContext) (string, error)
}
