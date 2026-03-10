package keeper_test

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"

	"github.com/davidsbond/keeper/internal/server"
	"github.com/davidsbond/keeper/pkg/keeper"
)

func setupTest(t *testing.T) *keeper.Client {
	t.Helper()

	dir, err := os.MkdirTemp(os.TempDir(), t.Name())
	require.NoError(t, err)

	key := base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{0}, 32))
	port := 49152 + rand.Intn(16384)

	group, ctx := errgroup.WithContext(t.Context())
	group.Go(func() error {
		return server.Run(ctx, server.Config{
			HTTP: server.HTTPConfig{
				Bind: fmt.Sprintf("0.0.0.0:%d", port),
			},
			Database: server.DatabaseConfig{
				Path:      dir,
				TTL:       time.Hour,
				MasterKey: key,
			},
			JWT: server.JWTConfig{
				Issuer:     "keeper.dsb.dev/test",
				TTL:        time.Hour,
				SigningKey: key,
				Audience:   "keeper.dsb.dev/test",
			},
		})
	})

	t.Cleanup(func() {
		require.NoError(t, group.Wait())
	})

	<-time.After(time.Second)
	return keeper.NewClient(fmt.Sprintf("http://0.0.0.0:%d", port))
}
