package tcp

import (
	"context"

	"github.com/DaiYuANg/arcgo/clientx"
	clientcodec "github.com/DaiYuANg/arcgo/clientx/codec"
)

type Client interface {
	clientx.Closer
	clientx.Dialer
	DialCodec(ctx context.Context, codec clientcodec.Codec, framer clientcodec.Framer) (*CodecConn, error)
}

var _ Client = (*DefaultClient)(nil)
var _ clientx.Dialer = (*DefaultClient)(nil)
