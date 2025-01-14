package interfaces

type SourceFileRepo interface {
	DownloadSourceFiles(scanID string, sourceFiles []SourceFile) error
}

type SourceFile struct {
	ResultID   string
	RemoteName string
	LocalName  string
}
