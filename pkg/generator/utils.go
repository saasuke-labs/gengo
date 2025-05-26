package generator

func merge(metadata1, metadata2 map[string]string) map[string]string {
	merged := make(map[string]string)

	for k, v := range metadata1 {
		merged[k] = v
	}

	for k, v := range metadata2 {
		merged[k] = v

	}

	return merged
}
