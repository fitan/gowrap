package test_data

func queryDTO(v Query) (res map[string]interface{}) {
	res = make(map[string]interface{})

	res["name = ?"] = v.Name
	if v.Age != nil {
		res["age > ?"] = *v.Age
	}
	res["id in ?"] = v.Ids
	res["ip = ?"] = v.VM.Ip

	return
}
