package libp2p

import (
	"context"
	"errors"
	"fmt"
	"github.com/ipfs/go-ipfs/repo"
	"github.com/libp2p/go-libp2p-core/network"
	"go.uber.org/fx"
	"os"
	"path/filepath"

	"github.com/libp2p/go-libp2p"
	rcmgr "github.com/libp2p/go-libp2p-resource-manager"
)

func ResourceManager() func(fx.Lifecycle, repo.Repo) (network.ResourceManager, Libp2pOpts, error) {
	return func(lc fx.Lifecycle, repo repo.Repo) (network.ResourceManager, Libp2pOpts, error) {
		var limiter *rcmgr.BasicLimiter
		var opts Libp2pOpts

		// FIXME(BLOCKING): Get repo path.
		// repo.Path()
		limitsIn, err := os.Open(filepath.Join(".", "limits.json"))
		switch {
		case err == nil:
			defer limitsIn.Close()
			limiter, err = rcmgr.NewDefaultLimiterFromJSON(limitsIn)
			if err != nil {
				return nil, opts, fmt.Errorf("error parsing limit file: %w", err)
			}
		case errors.Is(err, os.ErrNotExist):
			limiter = rcmgr.NewDefaultLimiter()
		default:
			return nil, opts, err
		}

		libp2p.SetDefaultServiceLimits(limiter)

		rcmgr, err := rcmgr.NewResourceManager(limiter)
		if err != nil {
			return nil, opts, fmt.Errorf("error creating resource manager: %w", err)
		}
		opts.Opts = append(opts.Opts, libp2p.ResourceManager(rcmgr))

		lc.Append(fx.Hook{
			OnStop: func(_ context.Context) error {
				return rcmgr.Close()
			}})

		return rcmgr, opts, nil
	}
}
