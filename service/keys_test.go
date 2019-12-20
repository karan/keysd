package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKeys(t *testing.T) {
	env := newTestEnv(t)
	ctx := context.TODO()

	// Alice
	aliceService, aliceCloseFn := newTestService(t, env)
	defer aliceCloseFn()

	testAuthSetup(t, aliceService, alice, true)
	testUserSetup(t, env, aliceService, alice.ID(), "alice", true)

	testRecoverKey(t, aliceService, charlie, true)
	testRecoverKey(t, aliceService, group, true)

	resp, err := aliceService.Keys(ctx, &KeysRequest{
		SortField: "kid",
	})
	require.NoError(t, err)
	require.Equal(t, "kid", resp.SortField)
	require.Equal(t, SortAsc, resp.SortDirection)
	require.Equal(t, 3, len(resp.Keys))
	require.Equal(t, "a6MtPHR36F9wG5orC8bhm8iPCE2xrXK41iZLwPZcLzqo", resp.Keys[0].KID)
	require.Equal(t, "cBYSYNgt45ZULLrVAseoFnmCt87mycnqDF5psywZ53VB", resp.Keys[1].KID)
	require.Equal(t, "gqPhYydcdbTzHUdqVrrqBnnAJK9tv3gYbrPKPBynjciM", resp.Keys[2].KID)
	require.Equal(t, 1, len(resp.Keys[0].Users))
	require.Equal(t, "alice", resp.Keys[0].Users[0].Name)
	require.Equal(t, PrivateKeyType, resp.Keys[0].Type)
	require.Equal(t, int64(1234567890001), resp.Keys[0].CreatedAt)
	require.Equal(t, int64(1234567890003), resp.Keys[0].PublishedAt)
	require.Equal(t, int64(1234567890002), resp.Keys[0].SavedAt)

	// Bob
	bobService, bobCloseFn := newTestService(t, env)
	defer bobCloseFn()

	testAuthSetup(t, bobService, bob, true)
	testUserSetup(t, env, bobService, bob.ID(), "bob", true)

	pullResp, err := bobService.Pull(ctx, &PullRequest{All: true})
	require.NoError(t, err)
	require.Equal(t, 4, len(pullResp.KIDs))
	require.Equal(t, "a6MtPHR36F9wG5orC8bhm8iPCE2xrXK41iZLwPZcLzqo", pullResp.KIDs[0])
	require.Equal(t, "cBYSYNgt45ZULLrVAseoFnmCt87mycnqDF5psywZ53VB", pullResp.KIDs[1])
	require.Equal(t, "gqPhYydcdbTzHUdqVrrqBnnAJK9tv3gYbrPKPBynjciM", pullResp.KIDs[2])
	require.Equal(t, "bDM13g2wsoBE8WN2jrPdLRHg2LFgNt2ZrLcP2bG4iuNi", pullResp.KIDs[3])

	resp, err = bobService.Keys(ctx, &KeysRequest{
		SortField: "kid",
	})
	require.NoError(t, err)
	require.Equal(t, "kid", resp.SortField)
	require.Equal(t, SortAsc, resp.SortDirection)
	require.Equal(t, 4, len(resp.Keys))
	require.Equal(t, "a6MtPHR36F9wG5orC8bhm8iPCE2xrXK41iZLwPZcLzqo", resp.Keys[0].KID)
	require.Equal(t, "bDM13g2wsoBE8WN2jrPdLRHg2LFgNt2ZrLcP2bG4iuNi", resp.Keys[1].KID)
	require.Equal(t, 1, len(resp.Keys[1].Users))
	require.Equal(t, "bob", resp.Keys[1].Users[0].Name)
	require.Equal(t, PrivateKeyType, resp.Keys[1].Type)
	require.Equal(t, "cBYSYNgt45ZULLrVAseoFnmCt87mycnqDF5psywZ53VB", resp.Keys[2].KID)
	require.Equal(t, PublicKeyType, resp.Keys[2].Type)
	require.Equal(t, "gqPhYydcdbTzHUdqVrrqBnnAJK9tv3gYbrPKPBynjciM", resp.Keys[3].KID)
	require.Equal(t, PublicKeyType, resp.Keys[3].Type)

	resp, err = bobService.Keys(ctx, &KeysRequest{
		SortField:     "kid",
		SortDirection: SortDesc,
	})
	require.NoError(t, err)
	require.Equal(t, "kid", resp.SortField)
	require.Equal(t, SortDesc, resp.SortDirection)
	require.Equal(t, 4, len(resp.Keys))
	require.Equal(t, "gqPhYydcdbTzHUdqVrrqBnnAJK9tv3gYbrPKPBynjciM", resp.Keys[0].KID)
	require.Equal(t, "cBYSYNgt45ZULLrVAseoFnmCt87mycnqDF5psywZ53VB", resp.Keys[1].KID)
	require.Equal(t, "bDM13g2wsoBE8WN2jrPdLRHg2LFgNt2ZrLcP2bG4iuNi", resp.Keys[2].KID)
	require.Equal(t, "a6MtPHR36F9wG5orC8bhm8iPCE2xrXK41iZLwPZcLzqo", resp.Keys[3].KID)

	resp, err = bobService.Keys(ctx, &KeysRequest{
		SortField: "user",
	})
	require.NoError(t, err)
	require.Equal(t, "user", resp.SortField)
	require.Equal(t, SortAsc, resp.SortDirection)
	require.Equal(t, 4, len(resp.Keys))
	// 0: alice
	require.Equal(t, "a6MtPHR36F9wG5orC8bhm8iPCE2xrXK41iZLwPZcLzqo", resp.Keys[0].KID)
	require.Equal(t, 1, len(resp.Keys[0].Users))
	require.Equal(t, "alice", resp.Keys[0].Users[0].Name)
	// 1: bob
	require.Equal(t, "bDM13g2wsoBE8WN2jrPdLRHg2LFgNt2ZrLcP2bG4iuNi", resp.Keys[1].KID)
	require.Equal(t, 1, len(resp.Keys[1].Users))
	require.Equal(t, "bob", resp.Keys[1].Users[0].Name)
	// 2: charlie
	require.Equal(t, "cBYSYNgt45ZULLrVAseoFnmCt87mycnqDF5psywZ53VB", resp.Keys[2].KID)
	require.Equal(t, 0, len(resp.Keys[2].Users))
	// 3: group
	require.Equal(t, "gqPhYydcdbTzHUdqVrrqBnnAJK9tv3gYbrPKPBynjciM", resp.Keys[3].KID)
	require.Equal(t, 0, len(resp.Keys[3].Users))

	resp, err = bobService.Keys(ctx, &KeysRequest{
		SortField:     "user",
		SortDirection: SortDesc,
	})
	require.NoError(t, err)
	require.Equal(t, "user", resp.SortField)
	require.Equal(t, SortDesc, resp.SortDirection)
	require.Equal(t, 4, len(resp.Keys))
	// 0: bob
	require.Equal(t, "bDM13g2wsoBE8WN2jrPdLRHg2LFgNt2ZrLcP2bG4iuNi", resp.Keys[0].KID)
	require.Equal(t, 1, len(resp.Keys[0].Users))
	require.Equal(t, "bob", resp.Keys[0].Users[0].Name)
	// 1: alice
	require.Equal(t, "a6MtPHR36F9wG5orC8bhm8iPCE2xrXK41iZLwPZcLzqo", resp.Keys[1].KID)
	require.Equal(t, 1, len(resp.Keys[1].Users))
	require.Equal(t, "alice", resp.Keys[1].Users[0].Name)
	// 2: group
	require.Equal(t, "gqPhYydcdbTzHUdqVrrqBnnAJK9tv3gYbrPKPBynjciM", resp.Keys[2].KID)
	// 3: charlie
	require.Equal(t, "cBYSYNgt45ZULLrVAseoFnmCt87mycnqDF5psywZ53VB", resp.Keys[3].KID)

	resp, err = bobService.Keys(ctx, &KeysRequest{
		SortField: "type",
	})
	require.NoError(t, err)
	require.Equal(t, "type", resp.SortField)
	require.Equal(t, SortAsc, resp.SortDirection)
	require.Equal(t, 4, len(resp.Keys))
	// 0: bob
	require.Equal(t, "bDM13g2wsoBE8WN2jrPdLRHg2LFgNt2ZrLcP2bG4iuNi", resp.Keys[0].KID)
	require.Equal(t, 1, len(resp.Keys[0].Users))
	require.Equal(t, "bob", resp.Keys[0].Users[0].Name)
	// 1: alice
	require.Equal(t, "a6MtPHR36F9wG5orC8bhm8iPCE2xrXK41iZLwPZcLzqo", resp.Keys[1].KID)
	require.Equal(t, 1, len(resp.Keys[1].Users))
	require.Equal(t, "alice", resp.Keys[1].Users[0].Name)
	// 2: charlie
	require.Equal(t, "cBYSYNgt45ZULLrVAseoFnmCt87mycnqDF5psywZ53VB", resp.Keys[2].KID)
	require.Equal(t, 0, len(resp.Keys[2].Users))
	// 3: group
	require.Equal(t, "gqPhYydcdbTzHUdqVrrqBnnAJK9tv3gYbrPKPBynjciM", resp.Keys[3].KID)
	require.Equal(t, 0, len(resp.Keys[3].Users))
}

func TestKeysMissingSigchain(t *testing.T) {
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env)
	defer closeFn()
	ctx := context.TODO()

	testAuthSetup(t, service, alice, true)
	testUserSetup(t, env, service, alice.ID(), "alice", true)

	_, err := service.scs.DeleteSigchain(alice.ID())
	require.NoError(t, err)

	resp, err := service.Keys(ctx, &KeysRequest{})
	require.NoError(t, err)
	require.Equal(t, 1, len(resp.Keys))
}
