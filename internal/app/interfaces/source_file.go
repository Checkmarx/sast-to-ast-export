package interfaces

type SourceFileRepo interface {
	DownloadSourceFiles(scanID string, sourceFiles []SourceFile, rmvdir string) error
}

type ExcludeFile struct {
	FileName string
}

type SourceFile struct {
	ResultID   string
	RemoteName string
	LocalName  string
}
