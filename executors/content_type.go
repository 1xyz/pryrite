package executor

import (
	"encoding/json"
	"fmt"
	"mime"
	"strings"

	"github.com/pkg/errors"
)

type ContentType struct {
	Type    string
	Subtype string
	Params  map[string]string
}

func Parse(val string) (*ContentType, error) {
	name, params, err := mime.ParseMediaType(val)
	if err != nil {
		return nil, errors.Wrap(err, val)
	}

	vals := strings.SplitN(name, "/", 2)
	if len(vals) != 2 {
		return nil, fmt.Errorf("failed to parse %s", name)
	}

	return &ContentType{
		Type:    vals[0],
		Subtype: vals[1],
		Params:  params,
	}, nil
}

func (ct *ContentType) Clone() *ContentType {
	newCT := &ContentType{
		Type:    ct.Type,
		Subtype: ct.Subtype,
		Params:  map[string]string{},
	}
	for k, v := range ct.Params {
		newCT.Params[k] = v
	}
	return newCT
}

// ParentOf returns true when the other content-type matches but may have more
// specific parameters (i.e. this content-type is a "parent" of the other).
func (ct *ContentType) ParentOf(other *ContentType) bool {
	if ct.Type == other.Type && ct.Subtype == other.Subtype {
		for myK, myV := range ct.Params {
			otherV, ok := other.Params[myK]
			if !ok || myV != otherV {
				return false
			}
		}

		return true
	}

	return false
}

func (ct *ContentType) String() string {
	return mime.FormatMediaType(ct.Type+"/"+ct.Subtype, ct.Params)
}

func (ct *ContentType) MarshalJSON() ([]byte, error) {
	return json.Marshal(ct.String())
}

func (ct *ContentType) UnmarshalJSON(data []byte) error {
	var val string
	err := json.Unmarshal(data, &val)
	if err != nil {
		return err
	}

	newCT, err := Parse(val)
	if err != nil {
		return err
	}

	ct.Type = newCT.Type
	ct.Subtype = newCT.Subtype
	ct.Params = newCT.Params

	return nil
}
