package loader

import (
	"context"
	"crypto/tls"
	"errors"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/hashicorp/vault/api"
	"github.com/mxmauro/channelcontext"
	"github.com/mxmauro/mergecontext"
)

// -----------------------------------------------------------------------------

type vaultClientHash [32]byte

// vaultClient
type vaultClient struct {
	host string

	headers map[string]string

	tlsConfig *tls.Config

	mountPath string

	accessToken string
	auth        api.AuthMethod

	apiClient *api.Client

	hash vaultClientHash

	mtx          sync.Mutex
	needLogin    bool
	loginAttempt int

	tokenMonitor vaultClientTokenMonitor
}

type vaultClientTokenMonitor struct {
	init      sync.Once
	ctx       context.Context
	cancelCtx context.CancelFunc
	doneCh    chan struct{}
	messageCh chan tokenMonitorMessage
}

type tokenMonitorMessage struct {
	secret       *api.Secret
	loginAttempt int
}

// -----------------------------------------------------------------------------

var (
	randGen = rand.New(rand.NewSource(int64(time.Now().Nanosecond())))
)

// -----------------------------------------------------------------------------

func newVaultClient(host string, headers map[string]string, tlsConfig *tls.Config, accessToken string, auth VaultAuthMethod) (*vaultClient, error) {
	var err error

	if len(host) == 0 {
		return nil, errors.New("host not set")
	}
	if len(accessToken) == 0 && auth == nil {
		return nil, errors.New("authentication method not specified")
	}

	// Vault client accessor
	client := &vaultClient{
		host:        host,
		headers:     headers,
		tlsConfig:   tlsConfig,
		accessToken: accessToken,
		hash:        calculateHash(host, headers, tlsConfig, accessToken, auth), // Calculate configuration hash
		mtx:         sync.Mutex{},
		needLogin:   true,
		tokenMonitor: vaultClientTokenMonitor{
			init:      sync.Once{},
			doneCh:    make(chan struct{}),
			messageCh: make(chan tokenMonitorMessage, 2),
		},
	}
	client.auth, err = auth.create()
	if err != nil {
		return nil, err
	}

	// Setup Vault api client configuration
	vaultCfg := api.Config{
		MaxRetries: 3,
		Address:    "http",
		HttpClient: &http.Client{
			Transport: httpTransport,
		},
	}

	if tlsConfig != nil {
		vaultCfg.Address += "s"

		// Clone our default transport
		t := httpTransport.Clone()
		t.TLSClientConfig = tlsConfig
		vaultCfg.HttpClient.Transport = t
	}

	// Append host to address
	vaultCfg.Address += "://" + host

	// Create Vault client accessor
	client.apiClient, err = api.NewClient(&vaultCfg)
	if err != nil {
		return nil, err
	}
	// Set provided access token (maybe none)
	client.apiClient.SetToken(accessToken)
	// Remove some settings that can be overridden with environment variables
	client.apiClient.ClearNamespace()

	// Add custom headers if provided
	if len(headers) > 0 {
		for key, value := range headers {
			client.apiClient.AddHeader(key, value)
		}
	}

	// Lookup in our weak map, if a similar configuration already exists
	client = loadOrStore(client)

	// Done
	return client, nil
}

func (client *vaultClient) onFinalize() {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	client.stopMonitorAuthToken()
}

func (client *vaultClient) readWithContext(ctx context.Context, path string) (*api.Secret, error) {
	// If the client was configured with an access token, just try to read
	if len(client.accessToken) > 0 {
		return client.apiClient.Logical().ReadWithContext(ctx, path)
	}

	// Serialize access
	client.mtx.Lock()
	defer client.mtx.Unlock()

ExecuteLogin:
	origNeedLogin := client.needLogin
	if client.needLogin {
		client.loginAttempt += 1

		// Execute initial login
		secret, err := client.apiClient.Auth().Login(ctx, client.auth)
		if err != nil {
			return nil, err
		}

		// Keep our token valid
		client.startMonitorAuthToken(secret)

		client.needLogin = false
	}

	// Now we can read the secret
	secret, err := client.apiClient.Logical().ReadWithContext(ctx, path)
	if err != nil {
		// If we get access denied but no login was executed, then it may happen we were renewing the token
		// near to expiration, and it was still pending to process.
		// In this case, assume we need to log in again. Worker code will ignore a potential renewed token.
		if origNeedLogin == false && apiResponseStatusCode(err) == 403 {
			client.needLogin = true
			goto ExecuteLogin
		}

		return nil, err
	}

	// Done
	return secret, nil
}

func (client *vaultClient) startMonitorAuthToken(secret *api.Secret) {
	client.tokenMonitor.init.Do(func() {
		client.tokenMonitor.doneCh = make(chan struct{})
		client.tokenMonitor.ctx, client.tokenMonitor.cancelCtx = context.WithCancel(context.Background())

		go client.monitorAuthTokenWorker()
	})

	client.tokenMonitor.messageCh <- tokenMonitorMessage{
		secret:       secret,
		loginAttempt: client.loginAttempt,
	}
}

func (client *vaultClient) stopMonitorAuthToken() {
	if client.tokenMonitor.cancelCtx != nil {
		client.tokenMonitor.cancelCtx()

		<-client.tokenMonitor.doneCh

		close(client.tokenMonitor.doneCh)
		client.tokenMonitor.doneCh = nil
		close(client.tokenMonitor.messageCh)
		client.tokenMonitor.messageCh = nil
		client.tokenMonitor.cancelCtx = nil
		client.tokenMonitor.ctx = nil
	}
}

func (client *vaultClient) monitorAuthTokenWorker() {
MainLoop:
	for {
		var message tokenMonitorMessage
		var renewable bool

		// Dormant zone
		select {
		case <-client.tokenMonitor.ctx.Done():
			break MainLoop

		case message = <-client.tokenMonitor.messageCh:
		}

	ProcessMessage:
		// Check if this message still belongs to our last login attempt
		client.mtx.Lock()
		if message.loginAttempt != client.loginAttempt {
			client.mtx.Unlock()
			continue // Discard and restart
		}
		client.mtx.Unlock()

		// Get TTL
		ttl, err := message.secret.TokenTTL()
		if err != nil {
			continue // This happens only when the secret has an invalid TTL
		}

		// If TTL is 0 then no need to monitor a non-expiring token
		if ttl <= 0 {
			continue
		}

		// Check if renewable
		renewable, err = message.secret.TokenIsRenewable()
		if err != nil {
			continue // This happens only when the secret has invalid data
		}

		// Monitor this secret
	NonRenewable:
		if !renewable {
			newMessage, quit := client.monitorNonRenewableAuthToken(ttl)
			if quit {
				break MainLoop
			}
			if newMessage != nil {
				message = *newMessage
				goto ProcessMessage
			}
		} else {
			newMessage, deniedRemainingTTL, quit := client.monitorRenewableAuthToken(ttl, message.secret.Auth.ClientToken)
			if quit {
				break MainLoop
			} else if deniedRemainingTTL != nil {
				// If the request was denied, probably the token does not have enough privileges to renew itself, so
				// fallback non-renewable
				renewable = false
				ttl = *deniedRemainingTTL
				if ttl > 0 {
					goto NonRenewable
				}
			} else if newMessage != nil {
				// monitorRenewableAuthToken only fills the secret field, preserve the rest
				message.secret = newMessage.secret
				goto ProcessMessage
			}
		}

		// If we reach here, assume our original token was invalid, and we need to log-in again
		client.mtx.Lock()
		if message.loginAttempt == client.loginAttempt {
			client.needLogin = true
		}
		client.mtx.Unlock()
	}

	client.tokenMonitor.doneCh <- struct{}{}
}

func (client *vaultClient) monitorNonRenewableAuthToken(ttl time.Duration) (message *tokenMonitorMessage, quit bool) {
	grace := calculateGracePeriod(ttl)
	toSleep := calculateSleepDuration(ttl, grace)

	select {
	case <-client.tokenMonitor.ctx.Done():
		quit = true
		return

	case msg := <-client.tokenMonitor.messageCh:
		message = &msg
		return

	case <-time.After(toSleep):
		return
	}
}

func (client *vaultClient) monitorRenewableAuthToken(ttl time.Duration, clientToken string) (message *tokenMonitorMessage, deniedRemainingTTL *time.Duration, quit bool) {
	var retryBackoff *backoff.ExponentialBackOff
	var toSleep time.Duration
	var accessDenied bool

	initialTime := time.Now()
	grace := calculateGracePeriod(ttl)

	for {
		if retryBackoff == nil {
			remainingTTL := initialTime.Add(ttl).Sub(time.Now())
			toSleep = calculateSleepDuration(remainingTTL, grace)
		} else {
			toSleep = retryBackoff.NextBackOff()
			if toSleep == backoff.Stop {
				// Could not renew, assume it expired
				return
			}
		}
		if toSleep == 0 {
			toSleep = 0
		}

		// Wait some time before renewal
		select {
		case <-client.tokenMonitor.ctx.Done():
			quit = true
			return

		case msg := <-client.tokenMonitor.messageCh:
			message = &msg
			return

		case <-time.After(toSleep):
		}

		message, accessDenied, quit = client.tryRenewAuthToken(clientToken)
		if quit || message != nil {
			return
		} else if accessDenied {
			remainingTTL := initialTime.Add(ttl).Sub(time.Now())
			deniedRemainingTTL = &remainingTTL
			return
		}

		// If we reach here, we were unable to renew (maybe due to an error), so let's start a retry backoff if
		// not done yet
		if retryBackoff == nil {
			remainingTTL := initialTime.Add(ttl).Sub(time.Now())

			retryBackoff = &backoff.ExponentialBackOff{
				MaxElapsedTime:      remainingTTL,
				RandomizationFactor: backoff.DefaultRandomizationFactor,
				InitialInterval:     10 * time.Second,
				MaxInterval:         5 * time.Minute,
				Multiplier:          2,
				Clock:               backoff.SystemClock,
			}
			retryBackoff.Reset()
		}
	}
}

func (client *vaultClient) tryRenewAuthToken(clientToken string) (message *tokenMonitorMessage, accessDenied bool, quit bool) {
	// Convert the secret changed signal to a context
	secretChangedCtx, secretChangedCancelCtx := channelcontext.New[tokenMonitorMessage](client.tokenMonitor.messageCh)
	defer secretChangedCancelCtx()

	// Try to renew the token
	ctx := mergecontext.New(client.tokenMonitor.ctx, secretChangedCtx)
	renewedSecret, err := client.apiClient.Auth().Token().RenewTokenAsSelfWithContext(ctx, clientToken, 0)

	// Check if our context was signalled
	select {
	case <-ctx.Done():
		switch ctx.DoneIndex() {
		case 0:
			// Quitting ...
			quit = true

		case 1:
			// The secret changed signal was triggered
			msg := secretChangedCtx.DoneValue()
			message = &msg

		default:
			// panic("please report this")
		}

	default:
		// If we didn't get any error, assume our token was successfully renewed
		if err == nil {
			message = &tokenMonitorMessage{
				secret: renewedSecret,
			}
		} else {
			if apiResponseStatusCode(err) == 403 {
				accessDenied = true
			}
		}
	}

	// On error, we will, eventually, retry the operation
	return
}

// -----------------------------------------------------------------------------

// calculateGracePeriod calculates the grace period based on the minimum of the remaining lease duration and the token
// increment value; it also adds some jitter to not have clients be in sync
func calculateGracePeriod(ttl time.Duration) time.Duration {
	if ttl <= 0 {
		return 0
	}

	jitterMax := float64(ttl.Nanoseconds()) * 0.1

	// For a given lease duration, we want to allow 80-90% of that to elapse,
	// so the remaining amount is the grace period
	return time.Duration(jitterMax) + time.Duration(uint64(randGen.Int63())%uint64(jitterMax))
}

// calculateSleepDuration calculates the amount of time the LifeTimeWatcher should sleep before re-entering its loop
func calculateSleepDuration(ttl, grace time.Duration) time.Duration {
	// The sleep duration is set to 2/3 of the current lease duration plus
	// 1/3 of the current grace period, which adds jitter.
	return time.Duration(float64(ttl.Nanoseconds())*2/3 + float64(grace.Nanoseconds())/3)
}

func apiResponseStatusCode(err error) int {
	if err != nil {
		var apiErr *api.ResponseError

		if errors.As(err, &apiErr) {
			return apiErr.StatusCode
		}
	}
	return 0
}
