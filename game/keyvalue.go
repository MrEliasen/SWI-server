package game

type KeyIntValue struct {
	Key   string `json:"key"`
	Value int32  `json:"value"`
}

type KeyStringValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type KeyFloatValue struct {
	Key   string  `json:"key"`
	Value float32 `json:"value"`
}
