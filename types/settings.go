package types

type Override []struct {
	In        string
	Out       string
	Extension string
	Template  string
}

type ProjectSettings struct {
	Override Override
}
type Config struct {
	Settings struct {
		Ignore  []string
		Project ProjectSettings
	}
}
