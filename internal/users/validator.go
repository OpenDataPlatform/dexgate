package users

import (
	"dexgate/internal/config"
	"gopkg.in/yaml.v2"
	"os"
)

type UserValidator interface {
	ValidateUser(claim string) (bool, error)
}

type UserValidatorImpl struct {
	config *UserConfig
	users  map[string]bool
	groups map[string]bool
	emails map[string]bool
}

func NewUserValidator(configFileName string) (UserValidator, error) {
	config.Log.Infof("Will use '%s' for users permissions", configFileName)
	file, err := os.Open(configFileName)
	if err != nil {
		return nil, err
	}
	uc := &UserConfig{}
	decoder := yaml.NewDecoder(file)
	decoder.SetStrict(true)
	if err = decoder.Decode(uc); err != nil {
		return nil, err
	}
	validator := &UserValidatorImpl{
		config: uc,
		users:  make(map[string]bool),
		groups: make(map[string]bool),
	}
	for _, user := range uc.AllowedUsers {
		validator.users[user] = true
	}
	for _, group := range uc.AllowedGroups {
		validator.groups[group] = true
	}
	for _, email := range uc.AllowedEmails {
		validator.emails[email] = true
	}
	return validator, nil
}

type claim struct {
	Name          string   `yaml:"name"`
	Email         string   `yaml:"email"`
	EmailVerified bool     `yaml:"email_verified"`
	Groups        []string `yaml:"groups"`
}

func (this *UserValidatorImpl) ValidateUser(claimJson string) (bool, error) {
	var claim claim
	err := yaml.Unmarshal([]byte(claimJson), &claim)
	if err != nil {
		return false, err
	}
	if claim.Name != "" {
		if _, ok := this.users[claim.Name]; ok {
			config.Log.Infof("User '%s' is allowed to access", claim.Name)
			return true, nil
		}
	}
	if claim.Groups != nil {
		for _, group := range claim.Groups {
			if _, ok := this.groups[group]; ok {
				config.Log.Infof("user '%s' belonging to group '%s' is allowed to access", claim.Name, group)
				return true, nil
			}
		}
	}
	if claim.Email != "" {
		if _, ok := this.emails[claim.Email]; ok {
			if claim.EmailVerified {
				config.Log.Infof("User '%s' with confirmed email '%s' is allowed to access", claim.Name, claim.Email)
				return true, nil
			} else {
				config.Log.Infof("Email '%s' (User '%s') is not confirmed, so not taken in account", claim.Email, claim.Name)
			}
		}
	}
	claim2, _ := yaml.Marshal(&claim)
	config.Log.Infof("User '%s' is NOT allowed to access this service. Claim: {\n%s}", claim.Name, claim2)
	return false, nil
}
