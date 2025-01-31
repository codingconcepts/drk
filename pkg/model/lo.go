package model

func TimesErr[T any](count int, iteratee func(index int) (T, error)) ([]T, error) {
	result := make([]T, count)
	var err error

	for i := 0; i < count; i++ {
		result[i], err = iteratee(i)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}
