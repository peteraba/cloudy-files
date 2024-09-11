package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/peteraba/cloudy-files/apperr"
)

// csrfTime represents the time a CSRF token is valid.
const csrfTime = time.Hour

// CSRFModel represents a CSRF model.
type CSRFModel struct {
	Token   string `json:"token"`
	Expires int64  `json:"expires"`
}

// CSRFModels represents a CSRF model list.
type CSRFModels []CSRFModel

// CSRFModelMap represents a CSRF model map.
type CSRFModelMap map[string]CSRFModels

// CSRF represents a CSRF repository.
type CSRF struct {
	store   Store
	lock    *sync.Mutex
	entries CSRFModelMap
}

// NewCSRF creates a new CSRF instance.
func NewCSRF(store Store) *CSRF {
	return &CSRF{
		store:   store,
		lock:    &sync.Mutex{},
		entries: make(CSRFModelMap),
	}
}

// Create creates a new user.
func (c *CSRF) Create(ctx context.Context, ipAddress, token string) error {
	err := c.readForWrite(ctx)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}
	defer c.store.Unlock(ctx)

	c.lock.Lock()
	defer c.lock.Unlock()

	if _, ok := c.entries[ipAddress]; !ok {
		c.entries[ipAddress] = CSRFModels{}
	}

	c.entries[ipAddress] = append(c.entries[ipAddress], CSRFModel{
		Token:   token,
		Expires: time.Now().Add(csrfTime).Unix(),
	})

	err = c.writeAfterRead(ctx)
	if err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	return nil
}

// Use checks if the CSRF token is valid and drops it so that it can not be reused.
func (c *CSRF) Use(ctx context.Context, ipAddress, token string) error {
	err := c.readForWrite(ctx)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}
	defer c.store.Unlock(ctx)

	c.lock.Lock()
	defer c.lock.Unlock()

	now := time.Now().Unix()

	ipAddressCsrf, ok := c.entries[ipAddress]
	if !ok {
		return fmt.Errorf("no CSRF token found for IP address '%s', err: %w", ipAddress, apperr.ErrAccessDenied)
	}

	for _, csrfModel := range ipAddressCsrf {
		if csrfModel.Token == token && csrfModel.Expires > now {
			delete(c.entries, ipAddress)

			err = c.writeAfterRead(ctx)
			if err != nil {
				return fmt.Errorf("error writing file: %w", err)
			}

			return nil
		}
	}

	return fmt.Errorf("no CSRF token found for IP address '%s', err: %w", ipAddress, apperr.ErrAccessDenied)
}

// readForWrite reads the session data from the store and creates entries.
// IMPORTANT!!! Do not forget to unlock the store after writing!
// Note: This function assumes that the store is NOT locked!
func (c *CSRF) readForWrite(ctx context.Context) error {
	data, err := c.store.ReadForWrite(ctx)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	err = c.createEntries(data)
	if err != nil {
		return fmt.Errorf("error creating entries: %w", err)
	}

	return nil
}

// writeAfterRead writes the current session data to the store.
// Note: This function assumes that the store is locked.
func (c *CSRF) writeAfterRead(ctx context.Context) error {
	data, _ := json.Marshal(c.entries) //nolint:errchkjson // We are sure that the data can be marshaled correctly

	err := c.store.WriteLocked(ctx, data)
	if err != nil {
		return fmt.Errorf("error storing data: %w", err)
	}

	return nil
}

// createEntries creates entries from data retrieved from store.
func (c *CSRF) createEntries(data []byte) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	entries := make(CSRFModelMap)

	if len(data) > 0 {
		err := json.Unmarshal(data, &entries)
		if err != nil {
			return fmt.Errorf("error unmarshaling data: %w", err)
		}
	}

	c.entries = entries

	return nil
}
