package workspace

// Info contains information about a workspace directory
type Info struct {
	Path      string
	Exists    bool
	FileCount int
	TotalSize int64
}
