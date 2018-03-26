package exception

import "errors"

func EnvironmentVariableMissingException(envVar string) error {
	return errors.New("Environment variable not set: " + envVar)
}
