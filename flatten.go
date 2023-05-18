package karma

import (
	"errors"
	"fmt"
	"strings"
)

func Flatten(err error) error {
	if err, ok := err.(Karma); ok {
		messages := []string{err.GetMessage()}
		keyvalues := err.GetContext().GetKeyValuePairs()

		err.Descend(
			func(reason Reason) {
				switch reason := reason.(type) {
				case Karma:
					messages = append(messages, reason.GetMessage())

					values := reason.GetContext().GetKeyValuePairs()
					if len(values) > 0 {
						for i := 0; i < len(values); i += 2 {
							keyvalues = append(keyvalues, values[i], values[i+1])
						}
					}
				default:
					messages = append(messages, fmt.Sprint(reason))
				}
			},
		)

		if len(keyvalues) > 0 {
			pairs := make([]string, len(keyvalues)/2)
			for i := 0; i < len(keyvalues); i += 2 {
				pairs[i/2] = fmt.Sprintf("%s=%v", keyvalues[i], keyvalues[i+1])
			}

			return errors.New(strings.Join(messages, ": ") + " | " + strings.Join(pairs, " "))
		} else {
			return errors.New(strings.Join(messages, ": "))
		}
	}

	return err
}
