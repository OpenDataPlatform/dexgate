package users

import (
	"dexgate/internal/config"
	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v2"
	"os"
)

type UserFilter interface {
	ValidateUser(claim string) (bool, error)
	Close()
}

type userFilterImpl struct {
	validator *userValidator
	watcher   *fsnotify.Watcher
}

func (this *userFilterImpl) ValidateUser(claim string) (bool, error) {
	return this.validator.validateUser(claim)
}

func (this *userFilterImpl) Close() {
	if this.watcher != nil {
		_ = this.watcher.Close()
	}
}

func NewUserFilter() (UserFilter, error) {
	validator, err := newUserValidator()
	if err != nil {
		return nil, err
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	impl := &userFilterImpl{
		validator: validator,
		watcher:   watcher,
	}

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					config.Log.Errorf("Users config file reload watcher has been closed. No more automatic reload!")
					return
				}
				//config.Log.Debugf("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					config.Log.Debugf("modified file:", event.Name)
					validator, err := newUserValidator()
					if err != nil {
						config.Log.Errorf("Error on reloading users config file: '%v'. Keep old version", err)
					}
					impl.validator = validator
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					config.Log.Errorf("Users config file reload watcher has been closed. No more automatic reload!")
					return
				}
				config.Log.Errorf("Error on users config file reload watcher :%v", err)
			}
		}
	}()
	err = watcher.Add(config.Conf.UserConfigFile)
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

func newUserValidator() (*userValidator, error) {
	config.Log.Infof("Will use '%s' for users permissions", config.Conf.UserConfigFile)
	file, err := os.Open(config.Conf.UserConfigFile)
	if err != nil {
		return nil, err
	}
	uc := &UserConfig{}
	decoder := yaml.NewDecoder(file)
	decoder.SetStrict(true)
	if err = decoder.Decode(uc); err != nil {
		return nil, err
	}
	validator := &userValidator{
		config: uc,
		users:  make(map[string]bool),
		groups: make(map[string]bool),
		emails: make(map[string]bool),
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
