package users

type UserConfig struct {
	AllowedUsers  []string `yaml:"allowedUsers"`
	AllowedGroups []string `yaml:"allowedGroups"`
	AllowedEmails []string `yaml:"allowedEmails"`
}
