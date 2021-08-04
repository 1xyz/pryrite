package executor

import (
	"encoding/json"
	"mime"
	"strings"

	"github.com/aardlabs/terminal-poc/tools"
)

type ContentType struct {
	Type    string
	Subtype string
	Params  map[string]string
}

func Parse(val string) (*ContentType, error) {
	// Be liberal with what we accept, here, as users can enter anything they
	// want and we don't want to prevent normal things like listing.

	name, params, err := mime.ParseMediaType(val)
	if err != nil {
		tools.Log.Warn().Str("contentType", val).Msg("Ignoring failure to parse content-type value")
		name = val
		params = map[string]string{}
	}

	vals := strings.SplitN(name, "/", 2)
	if len(vals) != 2 {
		tools.Log.Warn().Str("name", name).Msg("Ignoring failure to parse name from content-type")
		vals = []string{name, ""}
	}

	vals[1] = strings.ReplaceAll(vals[1], "/", "-")

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
func (ct *ContentType) ParentOf(other *ContentType, requiredKeys []string) bool {
	for _, key := range requiredKeys {
		if _, ok := ct.Params[key]; !ok {
			return false
		}
	}

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
