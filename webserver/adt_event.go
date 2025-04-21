package webserver

import "encoding/json"

type Event struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Data any    `json:"data"`
}

func (e *Event) ToBytes() []byte {
	data, err := json.Marshal(e.Data)

	if err != nil {
		panic(err)
	}

	event := ""

	if e.ID != "" {
		event += "id: " + e.ID + "\n"
	}

	event += "event: " + e.Name + "\ndata: "

	return append([]byte(event), data...)
}

func (e *Event) ToString() string {
	return string(e.ToBytes())
}
