package models

type Music struct {
	UploadDate int64  `json:"upload_date"`
	FileID     string `json:"file_id"`
	FilePath   string `json:"file_path"`
	IpAddress  string `json:"ip"`
	FileSize   int    `json:"file_size"`
	Type       string `json:"type"`
	Title      string `json:"title"`
	Filename   string `json:"filename"`
	Extractor  string `json:"extractor"`
}
