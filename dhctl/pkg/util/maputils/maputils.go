package maputils

func ExcludeKeys(m map[string]string, excludeKeys ...string) map[string]string {
	excludeKeysSet := make(map[string]struct{})
	for _, k := range excludeKeys {
		excludeKeysSet[k] = struct{}{}
	}

	res := make(map[string]string)

	for hostName, address := range m {
		if _, ok := excludeKeysSet[hostName]; ok {
			continue
		}

		res[hostName] = address
	}

	return res
}
