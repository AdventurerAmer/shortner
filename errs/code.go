package errs

import "encoding/json"

type Code int

const (
	CodeUnknown Code = iota
	CodeInternal
	CodeValidation
	CodeResourceNotFound
)

var codeToStr = map[Code]string{
	CodeUnknown:          "UNKNOWN",
	CodeInternal:         "INTERNAL",
	CodeValidation:       "VALIDATION",
	CodeResourceNotFound: "RESOURCE_NOT_FOUND",
}
var strToCode = make(map[string]Code)

func init() {
	for c := range codeToStr {
		s := codeToStr[c]
		strToCode[s] = c
	}
}

func (c Code) String() string {
	return codeToStr[c]
}

func (c Code) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}

func (c *Code) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	*c = strToCode[s]
	return nil
}
