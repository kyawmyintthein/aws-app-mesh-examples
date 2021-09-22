package bodymodifier

import (
	"github.com/google/martian/parse"
	"github.com/kyawmyintthein/aws-app-mesh-examples/colorapp/gateway/plugins/bodymodifier/modifier"
)

func init() {
	parse.Register("body.CustomModifier", FromJSON)
}

func FromJSON(b []byte) (*parse.Result, error) {
	msg, err := modifier.FromJSON(b)
	if err != nil {
		return nil, err
	}

	return parse.NewResult(msg, []parse.ModifierType{parse.Response})
}
