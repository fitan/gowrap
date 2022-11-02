package test_data

func NopEndpointWrap(data interface{}, err error) (interface{}, error) {
	return data, err
}
