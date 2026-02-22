package http

func mapToMetadata(field *string) []byte {
	if field == nil {
		return nil
	}

	return []byte(*field)
}
