package interfaces

type SourceFileRepo interface {
	DownloadSourceFiles(scanID string, sourceFiles []SourceFile) error
}

type SourceFile struct {
	RemoteName string
	LocalName  string
}
