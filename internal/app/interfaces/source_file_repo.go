package interfaces

type SourceFileRepo interface {
	DownloadSourceFiles(scanID string, sourceFiles map[string]string) error
}
