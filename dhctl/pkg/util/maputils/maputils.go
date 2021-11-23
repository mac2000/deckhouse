package maputils

func ExcludeKeys(m map[string]string, excludeKeys ...string) map[string]string {
	excludeKeysSet := make(map[string]struct{})
	for _, k := range excludeKeys {
		excludeKeysSet[k] = struct{}{}
	}

	res := make(map[string]string)

	for k, v := range m {
		if _, ok := excludeKeysSet[k]; ok {
			continue
		}

		res[k] = v
	}

	return res
}

func Values(m map[string]string) []string {
	keysList := make([]string, 0, len(m))
	for _, v := range m {
		keysList = append(keysList, v)
	}

	return keysList
}

func NotExistsKeys(m map[string]string, keys ...string) []string {
	res := make([]string, 0)
	for k := range m {
		exists := false
		for _, kk := range keys {
			if kk == k {
				exists = true
				break
			}
		}

		if !exists {
			res = append(res, k)
		}
	}

	return res
}
