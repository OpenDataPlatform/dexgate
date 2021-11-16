package configwatcher

type Callback func(data string)

/*
 We need to have a synchronous Get() on launch, to catch initial config error.
 Further errors will be notified, but we will keep old config.
*/

type ConfigWatcher interface {
	Get() (data string, err error)
	Watch(callback func(data string)) error
	GetName() (name string)
	Close()
}
