package struts

import "encoding/json"

func StructToJson(s interface{}) (out string, err error) {
	data, err := json.Marshal(s)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func StructToMap(s interface{}) (out map[string]interface{}, err error) {
	data, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	var m = make(map[string]interface{})
	err = json.Unmarshal(data, &m)
	if err != nil {
		return m, err
	}
	return m, nil
}

func JsonToStruct(in string, dest interface{}) (err error) {
	err = json.Unmarshal([]byte(in), dest)
	if err != nil {
		return err
	}
	return nil
}

func MapToStruct(in map[string]interface{}) (dest interface{}, err error) {
	data, err := json.Marshal(&in)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, dest)
	if err != nil {
		return nil, err
	}
	return dest, nil
}

func JsonToMap(s string) (m map[string]interface{}, err error) {
	m = make(map[string]interface{})
	err = json.Unmarshal([]byte(s), &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func JsonToMapWithModel(s string, model interface{}) (m map[string]interface{}, err error) {
	err = json.Unmarshal([]byte(s), &model)
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(model)
	if err != nil {
		return nil, err
	}
	m = make(map[string]interface{})
	err = json.Unmarshal(data, &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}
