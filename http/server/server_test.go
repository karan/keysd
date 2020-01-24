package server

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/keys-pub/keys"
	"github.com/stretchr/testify/require"
)

type clock struct {
	t    time.Time
	tick time.Duration
}

func newClock() *clock {
	return newClockAt(1234567890000)
}

func (c *clock) setTick(tick time.Duration) {
	c.tick = tick
}

func newClockAt(ts keys.TimeMs) *clock {
	t := keys.TimeFromMillis(ts)
	return &clock{
		t:    t,
		tick: time.Millisecond,
	}
}

// func newClockAtNow() *clock {
// 	return &clock{
// 		t:    time.Now(),
// 		tick: time.Millisecond,
// 	}
// }

func (c *clock) Now() time.Time {
	c.t = c.t.Add(c.tick)
	return c.t
}

type testServer struct {
	Server  *Server
	Handler http.Handler
}

// func testFirestore(t *testing.T) Fire {
// 	opts := []option.ClientOption{option.WithCredentialsFile("credentials.json")}
// 	fs, fsErr := firestore.NewFirestore("firestore://chilltest-3297b", opts...)
// 	require.NoError(t, fsErr)
// 	err := fs.Delete(context.TODO(), "/")
// 	require.NoError(t, err)
// 	return fs
// }

func testFire(t *testing.T, clock *clock) Fire {
	fi := keys.NewMem()
	fi.SetTimeNow(clock.Now)
	return fi
}

func testUserStore(t *testing.T, ds keys.DocumentStore, req keys.Requestor, clock *clock) *keys.UserStore {
	us, err := keys.NewUserStore(ds, keys.NewSigchainStore(ds), []string{keys.Twitter, keys.Github}, req, clock.Now)
	require.NoError(t, err)
	return us
}

func newTestServer(t *testing.T, clock *clock, fs Fire, users *keys.UserStore) *testServer {
	mc := NewMemTestCache(clock.Now)
	server := NewServer(fs, mc, users)
	tasks := NewTestTasks(server)
	server.SetTasks(tasks)
	server.SetInternalAuth(keys.RandString(32))
	server.SetNowFn(clock.Now)
	server.SetAccessFn(func(c AccessContext, resource AccessResource, action AccessAction) Access {
		return AccessAllow()
	})
	handler := NewHandler(server)
	return &testServer{
		Server:  server,
		Handler: handler,
	}
}

func (s *testServer) Serve(req *http.Request) (int, http.Header, string) {
	rr := httptest.NewRecorder()
	s.Handler.ServeHTTP(rr, req)
	return rr.Code, rr.Header(), rr.Body.String()
}

// type devServer struct {
// 	client *http.Client
// 	t      *testing.T
// }

// func newDevServer(t *testing.T) *devServer {
// 	client := &http.Client{}
// 	return &devServer{
// 		t:      t,
// 		client: client,
// 	}
// }

// func (s devServer) Serve(req *http.Request) (int, http.Header, string) {
// 	url, err := url.Parse("http://localhost:8080" + req.URL.RequestURI())
// 	require.NoError(s.t, err)
// 	req.URL = url
// 	resp, err := s.client.Do(req)
// 	require.NoError(s.t, err)
// 	b, err := ioutil.ReadAll(resp.Body)
// 	require.NoError(s.t, err)
// 	return resp.StatusCode, resp.Header, string(b)
// }

func userMock(t *testing.T, users *keys.UserStore, key *keys.SignKey, name string, service string, mock *keys.MockRequestor) *keys.Statement {
	url := ""
	switch service {
	case "github":
		url = fmt.Sprintf("https://gist.github.com/%s/1", name)
	case "twitter":
		url = fmt.Sprintf("https://twitter.com/%s/status/1", name)
	default:
		t.Fatal("unsupported service in test")
	}

	sc := keys.NewSigchain(key.PublicKey())
	usr, err := keys.NewUser(users, key.ID(), service, name, url, sc.LastSeq()+1)
	require.NoError(t, err)
	st, err := keys.GenerateUserStatement(sc, usr, key, users.Now())
	require.NoError(t, err)

	msg, err := usr.Sign(key)
	require.NoError(t, err)
	mock.SetResponse(url, []byte(msg))

	return st
}

func TestAccess(t *testing.T) {
	clock := newClock()
	fi := testFire(t, clock)
	rq := keys.NewMockRequestor()
	users := testUserStore(t, fi, rq, clock)
	srv := newTestServer(t, clock, fi, users)

	alice := keys.NewEd25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))

	upkCount := 0
	scCount := 0
	srv.Server.SetAccessFn(func(c AccessContext, resource AccessResource, action AccessAction) Access {
		switch resource {
		case UserPublicKeyResource:
			if action == Put {
				upkCount++
				if upkCount%2 == 0 {
					return AccessDenyTooManyRequests("")
				}
			}
		case SigchainResource:
			if action == Put {
				scCount++
				if scCount == 2 {
					return AccessDenyTooManyRequests("sigchain deny test")
				}
			}
		}
		return AccessAllow()
	})

	// PUT /sigchain/:kid/:seq (alice, allow)
	aliceSc := keys.NewSigchain(alice.PublicKey())
	aliceSt, err := keys.GenerateStatement(aliceSc, []byte("testing"), alice, "", clock.Now())
	require.NoError(t, err)
	err = aliceSc.Add(aliceSt)
	require.NoError(t, err)
	aliceStBytes := aliceSt.Bytes()
	req, err := http.NewRequest("PUT", fmt.Sprintf("/sigchain/%s/1", alice.ID()), bytes.NewReader(aliceStBytes))
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, http.StatusOK, code)

	// PUT /sigchain/:kid/:seq (alice, deny)
	aliceSt2, err := keys.GenerateStatement(aliceSc, []byte("testing"), alice, "", clock.Now())
	require.NoError(t, err)
	err = aliceSc.Add(aliceSt2)
	require.NoError(t, err)
	aliceStBytes2 := aliceSt2.Bytes()
	req, err = http.NewRequest("PUT", fmt.Sprintf("/sigchain/%s/2", alice.ID()), bytes.NewReader(aliceStBytes2))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusTooManyRequests, code)
	require.Equal(t, `{"error":{"code":429,"message":"sigchain deny test"}}`, body)

	bob := keys.NewEd25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x02}, 32)))

	// PUT /:kid/:seq (bob, allow)
	bobSc := keys.NewSigchain(bob.PublicKey())
	bobSt, err := keys.GenerateStatement(bobSc, []byte("testing"), bob, "", clock.Now())
	require.NoError(t, err)
	bobAddErr := bobSc.Add(bobSt)
	require.NoError(t, bobAddErr)
	bobStBytes := bobSt.Bytes()
	req, err = http.NewRequest("PUT", fmt.Sprintf("/%s/1", bob.ID()), bytes.NewReader(bobStBytes))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)

	// POST /task/check/:kid
	req, err = http.NewRequest("POST", "/task/check/"+alice.ID().String(), nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusForbidden, code)
	require.Equal(t, `{"error":{"code":403,"message":"no auth token specified"}}`, body)

	// POST /task/check/:kid (with auth)
	req, err = http.NewRequest("POST", "/task/check/"+alice.ID().String(), nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", srv.Server.internalAuth)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
}
