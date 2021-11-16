package users

import (
	"dexgate/internal/config"
	"dexgate/pkg/configwatcher"
	"fmt"
	"gopkg.in/yaml.v2"
)

type UserFilter interface {
	ValidateUser(claim string) (bool, error)
	Close()
}

type userFilterImpl struct {
	validator *userValidator
	watcher   configwatcher.ConfigWatcher
}

func (this *userFilterImpl) ValidateUser(claim string) (bool, error) {
	return this.validator.validateUser(claim)
}

func (this *userFilterImpl) Close() {
	if this.watcher != nil {
		this.watcher.Close()
	}
}

func NewUserFilter() (UserFilter, error) {
	var userWatcher configwatcher.ConfigWatcher
	var err error
	if config.Conf.UsersConfigFile != "" {
		userWatcher, err = configwatcher.NewConfigFileWatcher(config.Conf.UsersConfigFile, config.Log)
	} else if config.Conf.UsersConfigMap.ConfigMapName != "" {
		userWatcher, err = configwatcher.NewConfigMapWatcher(nil, config.Conf.UsersConfigMap.Namespace, config.Conf.UsersConfigMap.ConfigMapName, config.Conf.UsersConfigMap.ConfigMapKey, config.Log)
	} else {
		err = fmt.Errorf("Missing users watcher in configuration")
	}
	if err != nil {
		return nil, err
	}
	data, err := userWatcher.Get()
	if err != nil {
		return nil, err
	}
	validator, err := newUserValidator(data)
	if err != nil {
		return nil, err
	}
	config.Log.Infof("Sucessfully loaded initial users configuration from '%s'", userWatcher.GetName())

	impl := &userFilterImpl{
		validator: validator,
		watcher:   userWatcher,
	}
	usersCallback := func(data string) {
		v, err := newUserValidator(data)
		if err != nil {
			config.Log.Errorf("watcher on '%s': Error on reloading user configuration: '%v'. Keep old version", userWatcher.GetName(), err)
		} else {
			// Substitute the new validator
			config.Log.Infof("Sucessfully reloaded users configuration from '%s'", userWatcher.GetName())
			impl.validator = v
		}
	}

	err = userWatcher.Watch(usersCallback)
	if err != nil {
		return nil, err
	}
	return impl, nil
}

type userValidator struct {
	config *UserConfig
	users  map[string]bool
	groups map[string]bool
	emails map[string]bool
}

func newUserValidator(json string) (*userValidator, error) {
	uc := &UserConfig{}
	if err := yaml.UnmarshalStrict([]byte(json), uc); err != nil {
		return nil, fmt.Errorf("Error in parsing users yaml file: '%v'", err)
	}
	validator := &userValidator{
		config: uc,
		users:  make(map[string]bool),
		groups: make(map[string]bool),
		emails: make(map[string]bool),
	}
	// Transform lists in sets
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

func (this *userValidator) validateUser(claimJson string) (bool, error) {
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
				config.Log.Infof("user '%s' as belonging to group '%s' is allowed to access", claim.Name, group)
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
