package service

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/ds"
	"github.com/keys-pub/keys/user"
	"github.com/keys-pub/keys/util"
	"github.com/keys-pub/keysd/http/server"
	"github.com/stretchr/testify/require"
)

func testConfig(t *testing.T, appName string, serverURL string, keyringType string) (*Config, CloseFn) {
	if appName == "" {
		appName = "KeysTest-" + randName()
	}
	cfg, err := NewConfig(appName)
	require.NoError(t, err)
	cfg.Set("server", serverURL)
	cfg.Set("keyring", keyringType)

	closeFn := func() {
		removeErr := os.RemoveAll(cfg.AppDir())
		require.NoError(t, removeErr)
	}
	return cfg, closeFn
}

func randName() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(buf)
}

func writeTestFile(t *testing.T) string {
	inPath := keys.RandTempPath(".txt")
	writeErr := ioutil.WriteFile(inPath, []byte("test message"), 0644)
	require.NoError(t, writeErr)
	return inPath
}

func testFire(t *testing.T, clock *clock) server.Fire {
	fi := ds.NewMem()
	fi.SetTimeNow(clock.Now)
	return fi
}

type testEnv struct {
	clock *clock
	fi    server.Fire
	req   *util.MockRequestor
	users *user.Store
}

func newTestEnv(t *testing.T) *testEnv {
	clock := newClock()
	fi := testFire(t, clock)
	req := util.NewMockRequestor()
	users := testUserStore(t, fi, keys.NewSigchainStore(fi), req, clock)
	return &testEnv{
		clock: clock,
		fi:    fi,
		req:   req,
		users: users,
	}
}

func testUserStore(t *testing.T, dst ds.DocumentStore, scs keys.SigchainStore, req *util.MockRequestor, clock *clock) *user.Store {
	ust, err := user.NewStore(dst, scs, req, clock.Now)
	require.NoError(t, err)
	return ust
}

func newTestService(t *testing.T, env *testEnv, appName string) (*service, CloseFn) {
	return newTestServiceWithOpts(t, env, appName, "mem")
}

func newTestServiceWithOpts(t *testing.T, env *testEnv, appName string, keyringType string) (*service, CloseFn) {
	serverEnv := newTestServerEnv(t, env)
	cfg, closeCfg := testConfig(t, appName, serverEnv.url, keyringType)
	st, err := newKeyringStore(cfg)
	require.NoError(t, err)
	auth, err := newAuth(cfg, st)
	require.NoError(t, err)
	svc, err := newService(cfg, Build{Version: "1.2.3", Commit: "deadbeef"}, auth, env.req, env.clock.Now)
	require.NoError(t, err)

	closeFn := func() {
		serverEnv.closeFn()
		svc.Close()
		err := auth.keyring.Reset()
		require.NoError(t, err)
		closeCfg()
	}

	return svc, closeFn
}

func testAuthSetup(t *testing.T, service *service) {
	password := "testpassword"
	_, err := service.AuthSetup(context.TODO(), &AuthSetupRequest{
		Password: password,
	})
	require.NoError(t, err)
}

func testImportKey(t *testing.T, service *service, key *keys.EdX25519Key) {
	saltpack, err := keys.EncodeKeyToSaltpack(key, "testpassword")
	require.NoError(t, err)
	_, err = service.KeyImport(context.TODO(), &KeyImportRequest{
		In:       []byte(saltpack),
		Password: "testpassword",
	})
	require.NoError(t, err)
}

func testImportID(t *testing.T, service *service, kid keys.ID) {
	_, err := service.KeyImport(context.TODO(), &KeyImportRequest{
		In: []byte(kid.String()),
	})
	require.NoError(t, err)
}

func userSetupGithub(env *testEnv, service *service, key *keys.EdX25519Key, username string) error {
	resp, err := service.UserSign(context.TODO(), &UserSignRequest{
		KID:     key.ID().String(),
		Service: "github",
		Name:    username,
	})
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://gist.github.com/%s/1", username)
	env.req.SetResponse(url, []byte(resp.Message))

	_, err = service.UserAdd(context.TODO(), &UserAddRequest{
		KID:     key.ID().String(),
		Service: "github",
		Name:    username,
		URL:     url,
	})
	return err
}

func testUserSetupGithub(t *testing.T, env *testEnv, service *service, key *keys.EdX25519Key, username string) {
	err := userSetupGithub(env, service, key, username)
	require.NoError(t, err)
}

func userSetupReddit(env *testEnv, service *service, key *keys.EdX25519Key, username string) error {
	resp, err := service.UserSign(context.TODO(), &UserSignRequest{
		KID:     key.ID().String(),
		Service: "reddit",
		Name:    username,
	})
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://reddit.com/r/keyspubmsgs/comments/123/%s", username)
	rmsg := mockRedditMessage(username, resp.Message, "keyspubmsgs")
	env.req.SetResponse(url+".json", []byte(rmsg))

	_, err = service.UserAdd(context.TODO(), &UserAddRequest{
		KID:     key.ID().String(),
		Service: "reddit",
		Name:    username,
		URL:     url,
	})
	return err
}

func testUserSetupReddit(t *testing.T, env *testEnv, service *service, key *keys.EdX25519Key, username string) {
	err := userSetupReddit(env, service, key, username)
	require.NoError(t, err)
}

func mockRedditMessage(author string, msg string, subreddit string) string {
	msg = strings.ReplaceAll(msg, "\n", " ")
	return `[{   
		"kind": "Listing",
		"data": {
			"children": [
				{
					"kind": "t3",
					"data": {
						"author": "` + author + `",
						"selftext": "` + msg + `",
						"subreddit": "` + subreddit + `"
					}
				}
			]
		}
    }]`
}

func mockRedditURL(url string) string {
	return url + ".json"
}

// func testRemoveKey(t *testing.T, service *service, key *keys.EdX25519Key) {
// 	_, err := service.KeyRemove(context.TODO(), &KeyRemoveRequest{
// 		KID: key.ID().String(),
// 	})
// 	require.NoError(t, err)
// }

func testPush(t *testing.T, service *service, key *keys.EdX25519Key) {
	_, err := service.Push(context.TODO(), &PushRequest{
		Identity: key.ID().String(),
	})
	require.NoError(t, err)
}

func testPull(t *testing.T, service *service, kid keys.ID) {
	_, err := service.Pull(context.TODO(), &PullRequest{
		Identity: kid.String(),
	})
	require.NoError(t, err)
}

// func testUnlock(t *testing.T, service *service) {
// 	_, err := service.AuthUnlock(context.TODO(), &AuthUnlockRequest{
// 		Password: keys.RandPassphrase(12),
// 	})
// 	require.NoError(t, err)
// }

type clock struct {
	t time.Time
}

func newClock() *clock {
	t := util.TimeFromMillis(1234567890000)
	return &clock{
		t: t,
	}
}

func (c *clock) Now() time.Time {
	c.t = c.t.Add(time.Millisecond)
	return c.t
}

func (c *clock) Add(dt time.Duration) {
	c.t = c.t.Add(dt)
}

type serverEnv struct {
	url     string
	closeFn func()
}

func newTestServerEnv(t *testing.T, env *testEnv) *serverEnv {
	mc := server.NewMemTestCache(env.clock.Now)
	srv := server.NewServer(env.fi, mc, env.users, logger)
	srv.SetNowFn(env.clock.Now)
	tasks := server.NewTestTasks(srv)
	srv.SetTasks(tasks)
	srv.SetInternalAuth("testtoken")
	srv.SetAccessFn(func(c server.AccessContext, resource server.AccessResource, action server.AccessAction) server.Access {
		return server.AccessAllow()
	})
	handler := server.NewHandler(srv)
	testServer := httptest.NewServer(handler)
	srv.URL = testServer.URL

	closeFn := func() {
		testServer.Close()
	}
	return &serverEnv{
		url:     srv.URL,
		closeFn: closeFn,
	}
}

// func spewService(t *testing.T, service *service) {
// 	iter, iterErr := service.db.Documents(context.TODO(), "", nil)
// 	require.NoError(t, iterErr)
// 	spew, err := ds.Spew(iter, nil)
// 	require.NoError(t, err)
// 	t.Logf(spew.String())
// }

func TestRuntimeStatus(t *testing.T) {
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env, "")
	defer closeFn()

	resp, err := service.RuntimeStatus(context.TODO(), &RuntimeStatusRequest{})
	require.NoError(t, err)
	require.Equal(t, "1.2.3", resp.Version)
}

func TestKeyringFS(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	// keys.SetLogger(NewLogger(DebugLevel))
	// keyring.SetLogger(NewLogger(DebugLevel))

	env := newTestEnv(t)
	service, closeFn := newTestServiceWithOpts(t, env, "", "fs")
	defer closeFn()

	testAuthSetup(t, service)

	resp, err := service.RuntimeStatus(context.TODO(), &RuntimeStatusRequest{})
	require.NoError(t, err)
	require.Equal(t, "1.2.3", resp.Version)

	keys, err := service.Keys(context.TODO(), &KeysRequest{})
	require.NoError(t, err)
	require.Equal(t, 0, len(keys.Keys))
}

func TestCheckUpdate(t *testing.T) {
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env, "")
	defer closeFn()

	testAuthSetup(t, service)

	testImportKey(t, service, alice)
	testUserSetupGithub(t, env, service, alice, "alice")
	testPush(t, service, alice)

	err := service.checkForKeyUpdates(context.TODO())
	require.NoError(t, err)
}
