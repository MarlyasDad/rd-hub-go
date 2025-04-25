package bot

type ServiceStart interface {
	Start(tgID int64, firstName string, lastName string, username string, languageCode string) (string, error)
}
