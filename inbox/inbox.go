package inbox

import (
	"git.sr.ht/~rockorager/go-jmap"
	_ "git.sr.ht/~rockorager/go-jmap/mail"
	"sync"
)

type JMAPClient struct {
	jmap.Client
	//tokenGetter                 TokenGetterFunc
	mu                          sync.RWMutex
	userId                      jmap.ID
	userEmail                   string
	mailboxIds                  map[string]jmap.ID
	lastComputedCacheAtUnixTime int64
	cachedRenewalErr            error
}
